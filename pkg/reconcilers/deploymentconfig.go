package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// DCMutateFn is a function which mutates the existing DeploymentConfig into it's desired state.
type DCMutateFn func(desired, existing *appsv1.DeploymentConfig) (bool, error)

func DeploymentConfigMutator(opts ...DCMutateFn) MutateFn {
	return func(existingObj, desiredObj common.KubernetesObject) (bool, error) {
		existing, ok := existingObj.(*appsv1.DeploymentConfig)
		if !ok {
			return false, fmt.Errorf("%T is not a *appsv1.DeploymentConfig", existingObj)
		}
		desired, ok := desiredObj.(*appsv1.DeploymentConfig)
		if !ok {
			return false, fmt.Errorf("%T is not a *appsv1.DeploymentConfig", desiredObj)
		}

		update := false

		// Loop through each option
		for _, opt := range opts {
			tmpUpdate, err := opt(desired, existing)
			if err != nil {
				return false, err
			}
			update = update || tmpUpdate
		}

		return update, nil
	}
}

// GenericBackendMutators returns the generic mutators for backend
func GenericBackendMutators() []DCMutateFn {
	return []DCMutateFn{
		DeploymentConfigImageChangeTriggerMutator,
		DeploymentConfigContainerResourcesMutator,
		DeploymentConfigAffinityMutator,
		DeploymentConfigTolerationsMutator,
		DeploymentConfigPodTemplateLabelsMutator,
	}
}

func DeploymentConfigReplicasMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	update := false

	if desired.Spec.Replicas != existing.Spec.Replicas {
		existing.Spec.Replicas = desired.Spec.Replicas
		update = true
	}

	return update, nil
}

func DeploymentConfigAffinityMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity) {
		diff := cmp.Diff(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity)
		log.Info(fmt.Sprintf("%s spec.template.spec.Affinity has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec.Template.Spec.Affinity = desired.Spec.Template.Spec.Affinity
		updated = true
	}

	return updated, nil
}

func DeploymentConfigTolerationsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		diff := cmp.Diff(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations)
		log.Info(fmt.Sprintf("%s spec.template.spec.Tolerations has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec.Template.Spec.Tolerations = desired.Spec.Template.Spec.Tolerations
		updated = true
	}

	return updated, nil
}

func DeploymentConfigContainerResourcesMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	desiredName := common.ObjectInfo(desired)
	update := false

	if len(desired.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("%s desired spec.template.spec.containers length changed to '%d', should be 1", desiredName, len(desired.Spec.Template.Spec.Containers))
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		log.Info(fmt.Sprintf("%s spec.template.spec.containers length changed to '%d', recreating dc", desiredName, len(existing.Spec.Template.Spec.Containers)))
		existing.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		update = true
	}

	if !helper.CmpResources(&existing.Spec.Template.Spec.Containers[0].Resources, &desired.Spec.Template.Spec.Containers[0].Resources) {
		diff := cmp.Diff(existing.Spec.Template.Spec.Containers[0].Resources, desired.Spec.Template.Spec.Containers[0].Resources, cmpopts.IgnoreUnexported(resource.Quantity{}))
		log.Info(fmt.Sprintf("%s spec.template.spec.containers[0].resources have changed: %s", desiredName, diff))
		existing.Spec.Template.Spec.Containers[0].Resources = desired.Spec.Template.Spec.Containers[0].Resources
		update = true
	}

	return update, nil
}

// DeploymentConfigEnvVarReconciler implements basic env var reconcilliation for single container deployment configs.
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing DC
func DeploymentConfigEnvVarReconciler(desired, existing *appsv1.DeploymentConfig, envVar string) bool {
	update := false

	existingContainer := &existing.Spec.Template.Spec.Containers[0]
	desiredContainer := desired.Spec.Template.Spec.Containers[0]

	desiredIdx := helper.FindEnvVar(desiredContainer.Env, envVar)
	existingIdx := helper.FindEnvVar(existingContainer.Env, envVar)

	if desiredIdx < 0 && existingIdx >= 0 {
		// env var exists in existing and does not exist in desired => Remove from the list
		// shift all of the elements at the right of the deleting index by one to the left
		existingContainer.Env = append(existingContainer.Env[:existingIdx], existingContainer.Env[existingIdx+1:]...)
		update = true
	} else if desiredIdx < 0 && existingIdx < 0 {
		// env var does not exist in existing and does not exist in desired => NOOP
	} else if desiredIdx >= 0 && existingIdx < 0 {
		// env var does not exist in existing and exists in desired => ADD it
		existingContainer.Env = append(existingContainer.Env, desiredContainer.Env[desiredIdx])
		update = true
	} else {
		// env var exists in existing and exists in desired
		if !reflect.DeepEqual(existingContainer.Env[existingIdx], desiredContainer.Env[desiredIdx]) {
			existingContainer.Env[existingIdx] = desiredContainer.Env[desiredIdx]
			update = true
		}
	}
	return update
}

// DeploymentConfigImageChangeTriggerMutator ensures image change triggers are reconciled
func DeploymentConfigImageChangeTriggerMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	desiredDeploymentTriggerImageChangePos, err := helper.FindDeploymentTriggerOnImageChange(desired.Spec.Triggers)
	if err != nil {
		return false, fmt.Errorf("unexpected: '%s' in DeploymentConfig '%s'", err, desired.Name)

	}
	existingDeploymentTriggerImageChangePos, err := helper.FindDeploymentTriggerOnImageChange(existing.Spec.Triggers)
	if err != nil {
		return false, fmt.Errorf("unexpected: '%s' in DeploymentConfig '%s'", err, existing.Name)
	}

	desiredDeploymentTriggerImageChangeParams := desired.Spec.Triggers[desiredDeploymentTriggerImageChangePos].ImageChangeParams
	existingDeploymentTriggerImageChangeParams := existing.Spec.Triggers[existingDeploymentTriggerImageChangePos].ImageChangeParams

	if !reflect.DeepEqual(existingDeploymentTriggerImageChangeParams.From.Name, desiredDeploymentTriggerImageChangeParams.From.Name) {
		existingDeploymentTriggerImageChangeParams.From.Name = desiredDeploymentTriggerImageChangeParams.From.Name
		return true, nil
	}

	return false, nil
}

// DeploymentConfigPodTemplateLabelsMutator ensures pod template labels are reconciled
func DeploymentConfigPodTemplateLabelsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	helper.MergeMapStringString(&updated, &existing.Spec.Template.Labels, desired.Spec.Template.Labels)

	return updated, nil
}

// DeploymentConfigRemoveDuplicateEnvVarMutator ensures pod env vars are not duplicated
func DeploymentConfigRemoveDuplicateEnvVarMutator(_, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false
	for idx := range existing.Spec.Template.Spec.Containers {
		prunedEnvs := helper.RemoveDuplicateEnvVars(existing.Spec.Template.Spec.Containers[idx].Env)
		if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[idx].Env, prunedEnvs) {
			existing.Spec.Template.Spec.Containers[idx].Env = prunedEnvs
			updated = true
		}
	}

	return updated, nil
}
