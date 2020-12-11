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
type DCMutateFn func(desired, existing *appsv1.DeploymentConfig) bool

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
			tmpUpdate := opt(desired, existing)
			update = update || tmpUpdate
		}

		return update, nil
	}
}

func GenericDeploymentConfigMutator() MutateFn {
	return DeploymentConfigMutator(
		DeploymentConfigReplicasMutator,
		DeploymentConfigContainerResourcesMutator,
		DeploymentConfigAffinityMutator,
		DeploymentConfigTolerationsMutator,
	)
}

func DeploymentConfigReplicasMutator(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	if desired.Spec.Replicas != existing.Spec.Replicas {
		existing.Spec.Replicas = desired.Spec.Replicas
		update = true
	}

	return update
}

func DeploymentConfigAffinityMutator(desired, existing *appsv1.DeploymentConfig) bool {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity) {
		diff := cmp.Diff(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity)
		log.Info(fmt.Sprintf("%s spec.template.spec.Affinity has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec.Template.Spec.Affinity = desired.Spec.Template.Spec.Affinity
		updated = true
	}

	return updated
}

func DeploymentConfigTolerationsMutator(desired, existing *appsv1.DeploymentConfig) bool {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		diff := cmp.Diff(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations)
		log.Info(fmt.Sprintf("%s spec.template.spec.Tolerations has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec.Template.Spec.Tolerations = desired.Spec.Template.Spec.Tolerations
		updated = true
	}

	return updated
}

func DeploymentConfigContainerResourcesMutator(desired, existing *appsv1.DeploymentConfig) bool {
	desiredName := common.ObjectInfo(desired)
	update := false

	if len(desired.Spec.Template.Spec.Containers) != 1 {
		panic(fmt.Sprintf("%s desired spec.template.spec.containers length changed to '%d', should be 1", desiredName, len(desired.Spec.Template.Spec.Containers)))
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

	return update
}

func DeploymentConfigEnvVarMergeMutator(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	if len(existing.Spec.Template.Spec.Containers) != len(desired.Spec.Template.Spec.Containers) {
		existing.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		update = true
	}

	for idx := 0; idx < len(desired.Spec.Template.Spec.Containers); idx++ {
		existingContainer := &existing.Spec.Template.Spec.Containers[idx]
		desiredContainer := &desired.Spec.Template.Spec.Containers[idx]

		for _, desiredEnvVar := range desiredContainer.Env {
			envVarIdx := FindEnvVar(existingContainer.Env, desiredEnvVar.Name)
			if envVarIdx < 0 {
				existingContainer.Env = append(existingContainer.Env, desiredEnvVar)
				update = true
			} else if !reflect.DeepEqual(existingContainer.Env[idx], desiredEnvVar) {
				existingContainer.Env[envVarIdx] = desiredEnvVar
				update = true
			}
		}
	}

	return update
}
