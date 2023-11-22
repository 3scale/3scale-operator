package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
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
		DeploymentConfigPriorityClassMutator,
		DeploymentConfigTopologySpreadConstraintsMutator,
		DeploymentConfigPodTemplateAnnotationsMutator,
		DeploymentConfigArgsMutator,
		DeploymentConfigProbesMutator,
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

// DeploymentConfigEnvVarReconciler implements basic env var reconcilliation deployment configs.
// Existing and desired DC must have same number of containers
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing DC
func DeploymentConfigEnvVarReconciler(desired, existing *appsv1.DeploymentConfig, envVar string) bool {
	updated := false

	if len(desired.Spec.Template.Spec.Containers) != len(existing.Spec.Template.Spec.Containers) {
		log.Info("[WARNING] not reconciling deployment config",
			"name", client.ObjectKeyFromObject(desired),
			"reason", "existing and desired do not have same number of containers")
		return false
	}

	if len(desired.Spec.Template.Spec.InitContainers) != len(existing.Spec.Template.Spec.InitContainers) {
		log.Info("[WARNING] not reconciling deployment config",
			"name", client.ObjectKeyFromObject(desired),
			"reason", "existing and desired do not have same number of init containers")
		return false
	}

	// Init Containers
	for idx := range existing.Spec.Template.Spec.InitContainers {
		tmpChanged := helper.EnvVarReconciler(
			desired.Spec.Template.Spec.InitContainers[idx].Env,
			&existing.Spec.Template.Spec.InitContainers[idx].Env,
			envVar)
		updated = updated || tmpChanged
	}

	// Containers
	for idx := range existing.Spec.Template.Spec.Containers {
		tmpChanged := helper.EnvVarReconciler(
			desired.Spec.Template.Spec.Containers[idx].Env,
			&existing.Spec.Template.Spec.Containers[idx].Env,
			envVar)
		updated = updated || tmpChanged
	}

	// Pre Hook pod
	if existing.Spec.Strategy.RollingParams != nil &&
		existing.Spec.Strategy.RollingParams.Pre != nil &&
		existing.Spec.Strategy.RollingParams.Pre.ExecNewPod != nil &&
		desired.Spec.Strategy.RollingParams != nil &&
		desired.Spec.Strategy.RollingParams.Pre != nil &&
		desired.Spec.Strategy.RollingParams.Pre.ExecNewPod != nil {
		// reconcile Pre Hook
		tmpChanged := helper.EnvVarReconciler(
			desired.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env,
			&existing.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env,
			envVar)
		updated = updated || tmpChanged
	}

	// Post Hook pod
	if existing.Spec.Strategy.RollingParams != nil &&
		existing.Spec.Strategy.RollingParams.Post != nil &&
		existing.Spec.Strategy.RollingParams.Post.ExecNewPod != nil &&
		desired.Spec.Strategy.RollingParams != nil &&
		desired.Spec.Strategy.RollingParams.Post != nil &&
		desired.Spec.Strategy.RollingParams.Post.ExecNewPod != nil {
		// reconcile Pre Hook
		tmpChanged := helper.EnvVarReconciler(
			desired.Spec.Strategy.RollingParams.Post.ExecNewPod.Env,
			&existing.Spec.Strategy.RollingParams.Post.ExecNewPod.Env,
			envVar)
		updated = updated || tmpChanged
	}

	return updated
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

// DeploymentConfigPriorityClassMutator ensures priorityclass is reconciled
func DeploymentConfigPriorityClassMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	if existing.Spec.Template.Spec.PriorityClassName != desired.Spec.Template.Spec.PriorityClassName {
		existing.Spec.Template.Spec.PriorityClassName = desired.Spec.Template.Spec.PriorityClassName
		updated = true
	}

	return updated, nil
}

// DeploymentConfigStrategyMutator ensures desired strategy
func DeploymentConfigStrategyMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Strategy, desired.Spec.Strategy) {
		existing.Spec.Strategy = desired.Spec.Strategy
		updated = true
	}

	return updated, nil
}

// DeploymentConfigTopologySpreadConstraintsMutator ensures TopologySpreadConstraints is reconciled
func DeploymentConfigTopologySpreadConstraintsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints) {
		diff := cmp.Diff(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints)
		log.Info(fmt.Sprintf("%s spec.template.spec.TopologySpreadConstraints has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec.Template.Spec.TopologySpreadConstraints = desired.Spec.Template.Spec.TopologySpreadConstraints
		updated = true
	}

	return updated, nil
}

// DeploymentConfigPodTemplateAnnotationsMutator ensures Pod Template Annotations is reconciled
func DeploymentConfigPodTemplateAnnotationsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	helper.MergeMapStringString(&updated, &existing.Spec.Template.Annotations, desired.Spec.Template.Annotations)

	return updated, nil
}

// DeploymentConfigArgsMutator ensures Args are reconciled
func DeploymentConfigArgsMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	for i, desiredContainer := range desired.Spec.Template.Spec.Containers {
		existingContainer := &existing.Spec.Template.Spec.Containers[i]

		if !reflect.DeepEqual(existingContainer.Args, desiredContainer.Args) {
			existingContainer.Args = desiredContainer.Args
			updated = true
		}
	}

	return updated, nil
}

// DeploymentConfigProbesMutator ensures probes are reconciled
func DeploymentConfigProbesMutator(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := false

	for i, desiredContainer := range desired.Spec.Template.Spec.Containers {
		existingContainer := &existing.Spec.Template.Spec.Containers[i]

		if !reflect.DeepEqual(existingContainer.LivenessProbe, desiredContainer.LivenessProbe) {
			existingContainer.LivenessProbe = desiredContainer.LivenessProbe
			updated = true
		}

		if !reflect.DeepEqual(existingContainer.ReadinessProbe, desiredContainer.ReadinessProbe) {
			existingContainer.ReadinessProbe = desiredContainer.ReadinessProbe
			updated = true
		}
	}

	return updated, nil
}
