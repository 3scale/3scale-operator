package reconcilers

import (
	"fmt"
	"reflect"
	"slices"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
)

const (
	DeploymentKind          = "Deployment"
	DeploymentAPIVersion    = "apps/v1"
	DeploymentLabelSelector = "deployment"
)

type ContainerImage struct {
	Name string
	Tag  string
}

type ImageTriggerFrom struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

type ImageTrigger struct {
	From      ImageTriggerFrom `json:"from"`
	FieldPath string           `json:"fieldPath"`
	// +optional
	Paused bool `json:"paused,omitempty"`
}

// DMutateFn is a function which mutates the existing Deployment into it's desired state.
type DMutateFn func(desired, existing *k8sappsv1.Deployment) (bool, error)

func DeploymentMutator(opts ...DMutateFn) MutateFn {
	return func(existingObj, desiredObj common.KubernetesObject) (bool, error) {
		existing, ok := existingObj.(*k8sappsv1.Deployment)
		if !ok {
			return false, fmt.Errorf("%T is not a *k8sappsv1.Deployment", existingObj)
		}
		desired, ok := desiredObj.(*k8sappsv1.Deployment)
		if !ok {
			return false, fmt.Errorf("%T is not a *k8sappsv1.Deployment", desiredObj)
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

// GenericBackendDeploymentMutators returns the generic mutators for backend
func GenericBackendDeploymentMutators() []DMutateFn {
	return []DMutateFn{
		DeploymentAnnotationsMutator,
		DeploymentContainerResourcesMutator,
		DeploymentAffinityMutator,
		DeploymentTolerationsMutator,
		DeploymentPodTemplateLabelsMutator,
		DeploymentPriorityClassMutator,
		DeploymentTopologySpreadConstraintsMutator,
		DeploymentPodTemplateAnnotationsMutator,
		DeploymentArgsMutator,
		DeploymentProbesMutator,
		DeploymentPodContainerImageMutator,
		DeploymentPodInitContainerImageMutator,
	}
}

// DeploymentAnnotationsMutator ensures Deployment Annotations are reconciled
func DeploymentAnnotationsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	helper.MergeMapStringString(&updated, &existing.ObjectMeta.Annotations, desired.ObjectMeta.Annotations)

	return updated, nil
}

func DeploymentReplicasMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	update := false

	if desired.Spec.Replicas != existing.Spec.Replicas {
		existing.Spec.Replicas = desired.Spec.Replicas
		update = true
	}

	return update, nil
}

func DeploymentAffinityMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity) {
		diff := cmp.Diff(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity)
		log.Info(fmt.Sprintf("%s spec.template.spec.Affinity has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec.Template.Spec.Affinity = desired.Spec.Template.Spec.Affinity
		updated = true
	}

	return updated, nil
}

func DeploymentTolerationsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		diff := cmp.Diff(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations)
		log.Info(fmt.Sprintf("%s spec.template.spec.Tolerations has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec.Template.Spec.Tolerations = desired.Spec.Template.Spec.Tolerations
		updated = true
	}

	return updated, nil
}

func DeploymentContainerResourcesMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	desiredName := common.ObjectInfo(desired)
	update := false

	if len(desired.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("%s desired spec.template.spec.containers length changed to '%d', should be 1", desiredName, len(desired.Spec.Template.Spec.Containers))
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		log.Info(fmt.Sprintf("%s spec.template.spec.containers length changed to '%d', recreating deployment", desiredName, len(existing.Spec.Template.Spec.Containers)))
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

// DeploymentEnvVarReconciler implements basic env var reconciliation deployments.
// Existing and desired Deployment must have same number of containers
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing Deployment
func DeploymentEnvVarReconciler(desired, existing *k8sappsv1.Deployment, envVar string) bool {
	updated := false

	if len(desired.Spec.Template.Spec.Containers) != len(existing.Spec.Template.Spec.Containers) {
		log.Info("[WARNING] not reconciling deployment",
			"name", client.ObjectKeyFromObject(desired),
			"reason", "existing and desired do not have same number of containers")
		return false
	}

	if len(desired.Spec.Template.Spec.InitContainers) != len(existing.Spec.Template.Spec.InitContainers) {
		log.Info("[WARNING] not reconciling deployment",
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

	return updated
}

// DeploymentPodTemplateLabelsMutator ensures pod template labels are reconciled
func DeploymentPodTemplateLabelsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	helper.MergeMapStringString(&updated, &existing.Spec.Template.Labels, desired.Spec.Template.Labels)

	return updated, nil
}

// DeploymentRemoveDuplicateEnvVarMutator ensures pod env vars are not duplicated
func DeploymentRemoveDuplicateEnvVarMutator(_, existing *k8sappsv1.Deployment) (bool, error) {
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

// DeploymentPriorityClassMutator ensures priorityclass is reconciled
func DeploymentPriorityClassMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	if existing.Spec.Template.Spec.PriorityClassName != desired.Spec.Template.Spec.PriorityClassName {
		existing.Spec.Template.Spec.PriorityClassName = desired.Spec.Template.Spec.PriorityClassName
		updated = true
	}

	return updated, nil
}

// DeploymentStrategyMutator ensures desired strategy
func DeploymentStrategyMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Strategy, desired.Spec.Strategy) {
		existing.Spec.Strategy = desired.Spec.Strategy
		updated = true
	}

	return updated, nil
}

// DeploymentTopologySpreadConstraintsMutator ensures TopologySpreadConstraints is reconciled
func DeploymentTopologySpreadConstraintsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints) {
		diff := cmp.Diff(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints)
		log.Info(fmt.Sprintf("%s spec.template.spec.TopologySpreadConstraints has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec.Template.Spec.TopologySpreadConstraints = desired.Spec.Template.Spec.TopologySpreadConstraints
		updated = true
	}

	return updated, nil
}

// DeploymentPodTemplateAnnotationsMutator ensures Pod Template Annotations is reconciled
func DeploymentPodTemplateAnnotationsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	helper.MergeMapStringString(&updated, &existing.Spec.Template.Annotations, desired.Spec.Template.Annotations)

	return updated, nil
}

// DeploymentArgsMutator ensures deployment's containers' args are reconciled
func DeploymentArgsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

// DeploymentProbesMutator ensures probes are reconciled
func DeploymentProbesMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

// DeploymentPodContainerImageMutator ensures that the deployment's pod's containers are reconciled
func DeploymentPodContainerImageMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	for i, desiredContainer := range desired.Spec.Template.Spec.Containers {
		existingContainer := &existing.Spec.Template.Spec.Containers[i]

		if !reflect.DeepEqual(existingContainer.Image, desiredContainer.Image) {
			existingContainer.Image = desiredContainer.Image
			updated = true
		}
	}
	return updated, nil
}

// DeploymentPodInitContainerImageMutator ensures that the deployment's pod's containers are reconciled
func DeploymentPodInitContainerImageMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	for i, desiredContainer := range desired.Spec.Template.Spec.InitContainers {
		if i >= len(existing.Spec.Template.Spec.InitContainers) {
			// Add missing containers from desired to existing
			existing.Spec.Template.Spec.InitContainers = append(existing.Spec.Template.Spec.InitContainers, desiredContainer)
			fmt.Printf("Added missing container: %s\n", desiredContainer.Name)
			updated = true
			continue
		}
		existingContainer := &existing.Spec.Template.Spec.InitContainers[i]

		if !reflect.DeepEqual(existingContainer.Image, desiredContainer.Image) {
			existingContainer.Image = desiredContainer.Image
			updated = true
		}
	}
	return updated, nil
}

// DeploymentPodInitContainerMutator ensures that the deployment's pod's init containers are reconciled
func DeploymentPodInitContainerMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	// Trim excess containers if existing has more than desired
	if len(existing.Spec.Template.Spec.InitContainers) > len(desired.Spec.Template.Spec.InitContainers) {
		existing.Spec.Template.Spec.InitContainers = existing.Spec.Template.Spec.InitContainers[:len(desired.Spec.Template.Spec.InitContainers)]
		updated = true
	}

	// Ensure init containers match
	for i := range desired.Spec.Template.Spec.InitContainers {
		if i >= len(existing.Spec.Template.Spec.InitContainers) {
			// Append missing containers
			existing.Spec.Template.Spec.InitContainers = append(existing.Spec.Template.Spec.InitContainers, desired.Spec.Template.Spec.InitContainers[i])
			updated = true
		} else if !reflect.DeepEqual(existing.Spec.Template.Spec.InitContainers[i], desired.Spec.Template.Spec.InitContainers[i]) {
			// Update mismatched containers
			existing.Spec.Template.Spec.InitContainers[i] = desired.Spec.Template.Spec.InitContainers[i]
			updated = true
		}
	}

	return updated, nil
}

// DeploymentVolumesMutator implements strict Volumes reconcilliation
// Does not allow manually added volumes
func DeploymentVolumesMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	// Copy so can can perform sort
	existingVolumes := existing.Spec.Template.Spec.DeepCopy().Volumes
	desiredVolumes := desired.Spec.Template.Spec.DeepCopy().Volumes

	if len(existingVolumes) != len(desiredVolumes) {
		existing.Spec.Template.Spec.Volumes = desired.Spec.Template.Spec.Volumes
		return true, nil
	}

	// We sort the two volumes slice so that we can perform the
	// comparisons in a deterministic order.
	sort.Slice(existingVolumes, func(i, j int) bool {
		return existingVolumes[i].Name < existingVolumes[j].Name
	})

	sort.Slice(desiredVolumes, func(i, j int) bool {
		return desiredVolumes[i].Name < desiredVolumes[j].Name
	})

	for i, desiredVolume := range desiredVolumes {
		if !reflect.DeepEqual(existingVolumes[i], desiredVolume) {
			existing.Spec.Template.Spec.Volumes = desired.Spec.Template.Spec.Volumes
			updated = true
		}
	}

	return updated, nil
}

// DeploymentInitContainerVolumeMountsMutator implements strict VolumeMounts reconcilliation
// Does not allow manually added volumes
func DeploymentInitContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	if len(desired.Spec.Template.Spec.InitContainers) != len(existing.Spec.Template.Spec.InitContainers) {
		log.Info("[WARNING] not reconciling deployment",
			"name", client.ObjectKeyFromObject(desired),
			"reason", "existing and desired do not have same number of containers")
		return false, nil
	}

	for i, desiredContainer := range desired.Spec.Template.Spec.InitContainers {
		existingContainer := &existing.Spec.Template.Spec.InitContainers[i]

		if !reflect.DeepEqual(existingContainer.VolumeMounts, desiredContainer.VolumeMounts) {
			existingContainer.VolumeMounts = desiredContainer.VolumeMounts
			updated = true
		}
	}

	return updated, nil
}

// DeploymentContainerVolumeMountsMutator implements strict VolumeMounts reconcilliation
// Does not allow manually added volumes
func DeploymentContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	updated := false

	if len(desired.Spec.Template.Spec.Containers) != len(existing.Spec.Template.Spec.Containers) {
		log.Info("[WARNING] not reconciling deployment",
			"name", client.ObjectKeyFromObject(desired),
			"reason", "existing and desired do not have same number of containers")
		return false, nil
	}

	for i, desiredContainer := range desired.Spec.Template.Spec.Containers {
		existingContainer := &existing.Spec.Template.Spec.Containers[i]

		if !reflect.DeepEqual(existingContainer.VolumeMounts, desiredContainer.VolumeMounts) {
			existingContainer.VolumeMounts = desiredContainer.VolumeMounts
			updated = true
		}
	}

	return updated, nil
}

// WeakDeploymentVolumesMutator implements basic Volume reconciliation deployments.
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing Deployment
func WeakDeploymentVolumesMutator(desired, existing *k8sappsv1.Deployment, volumesName []string) (bool, error) {
	updated := false

	// Copy so can can perform sort
	existingSpec := &existing.Spec.Template.Spec
	desiredSpec := &desired.Spec.Template.Spec

	for _, volumeName := range volumesName {
		desiredIdx := helper.FindVolumeByName(desiredSpec.Volumes, volumeName)
		existingIdx := helper.FindVolumeByName(existingSpec.Volumes, volumeName)

		if desiredIdx < 0 && existingIdx >= 0 {
			// env var exists in existing and does not exist in desired => Remove from the list
			// shift all of the elements at the right of the deleting index by one to the left
			existingSpec.Volumes = slices.Delete(existingSpec.Volumes, existingIdx, existingIdx+1)
			updated = true
		} else if desiredIdx < 0 && existingIdx < 0 {
			// env var does not exist in existing and does not exist in desired => NOOP
		} else if desiredIdx >= 0 && existingIdx < 0 {
			// env var does not exist in existing and exists in desired => ADD it
			existingSpec.Volumes = append(existingSpec.Volumes, desiredSpec.Volumes[desiredIdx])
			updated = true
		} else {
			// env var exists in existing and exists in desired
			if !reflect.DeepEqual(existingSpec.Volumes[existingIdx], desiredSpec.Volumes[desiredIdx]) {
				existingSpec.Volumes[existingIdx] = desiredSpec.Volumes[desiredIdx]
				updated = true
			}
		}
	}

	return updated, nil
}

// WeakDeploymentInitContainerVolumeMountsMutator implements basic VolumeMounts reconciliation deployments.
// Existing and desired Deployment must have same number of containers
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing Deployment
func WeakDeploymentInitContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment, volumeMountsName []string) (bool, error) {
	if len(desired.Spec.Template.Spec.InitContainers) != len(existing.Spec.Template.Spec.InitContainers) {
		log.Info("[WARNING] not reconciling deployment",
			"name", client.ObjectKeyFromObject(desired),
			"reason", "existing and desired do not have same number of containers")
		return false, nil
	}

	updated := false
	for idx, desiredContainer := range desired.Spec.Template.Spec.InitContainers {
		existingContainer := &existing.Spec.Template.Spec.InitContainers[idx]
		tmpChanged := weakVolumeMountMutator(&desiredContainer, existingContainer, volumeMountsName)
		updated = updated || tmpChanged
	}

	return updated, nil
}

// WeakDeploymentContainerVolumeMountsMutator implements basic volumes reconciliation deployments.
// Existing and desired Deployment must have same number of containers
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing Deployment
func WeakDeploymentContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment, volumeMountsName []string) (bool, error) {
	if len(desired.Spec.Template.Spec.InitContainers) != len(existing.Spec.Template.Spec.InitContainers) {
		log.Info("[WARNING] not reconciling deployment",
			"name", client.ObjectKeyFromObject(desired),
			"reason", "existing and desired do not have same number of containers")
		return false, nil
	}

	updated := false
	for idx, desiredContainer := range desired.Spec.Template.Spec.Containers {
		existingContainer := &existing.Spec.Template.Spec.Containers[idx]
		tmpChanged := weakVolumeMountMutator(&desiredContainer, existingContainer, volumeMountsName)
		updated = updated || tmpChanged
	}

	return updated, nil
}

func weakVolumeMountMutator(desiredContainer, existingContainer *corev1.Container, volumeMountsName []string) bool {
	updated := false

	for _, volumeMountName := range volumeMountsName {
		desiredIdx := helper.FindVolumeMountByName(desiredContainer.VolumeMounts, volumeMountName)
		existingIdx := helper.FindVolumeMountByName(existingContainer.VolumeMounts, volumeMountName)

		if desiredIdx < 0 && existingIdx >= 0 {
			// env var exists in existing and does not exist in desired => Remove from the list
			// shift all of the elements at the right of the deleting index by one to the left
			existingContainer.VolumeMounts = slices.Delete(existingContainer.VolumeMounts, existingIdx, existingIdx+1)
			updated = true
		} else if desiredIdx < 0 && existingIdx < 0 {
			// env var does not exist in existing and does not exist in desired => NOOP
		} else if desiredIdx >= 0 && existingIdx < 0 {
			// env var does not exist in existing and exists in desired => ADD it
			existingContainer.VolumeMounts = append(existingContainer.VolumeMounts, desiredContainer.VolumeMounts[desiredIdx])
			updated = true
		} else {
			// env var exists in existing and exists in desired
			if !reflect.DeepEqual(existingContainer.VolumeMounts[existingIdx], desiredContainer.VolumeMounts[desiredIdx]) {
				existingContainer.VolumeMounts[existingIdx] = desiredContainer.VolumeMounts[desiredIdx]
				updated = true
			}
		}
	}

	return updated
}
