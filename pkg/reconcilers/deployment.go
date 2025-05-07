package reconcilers

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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

func DeploymentWorkerEnvMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	update := true
	// Always set env var CONFIG_REDIS_ASYNC to 1 this logic is only hit when you don't have logical redis db
	for envId, envVar := range existing.Spec.Template.Spec.Containers[0].Env {
		if envVar.Name == "CONFIG_REDIS_ASYNC" {
			if envVar.Value == "0" {
				existing.Spec.Template.Spec.Containers[0].Env[envId].Value = "1"
				update = true
				return update, nil
			}
			update = false

		}
	}
	// Adds the env CONFIG_REDIS_ASYNC if not present
	if update {
		existing.Spec.Template.Spec.Containers[0].Env = append(existing.Spec.Template.Spec.Containers[0].Env,
			helper.EnvVarFromValue("CONFIG_REDIS_ASYNC", "1"))
	}
	return update, nil
}

func DeploymentWorkerDisableAsyncEnvMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	update := true
	// Always set env var CONFIG_REDIS_ASYNC to 1 this logic is only hit when you don't have logical redis db
	for envId, envVar := range existing.Spec.Template.Spec.Containers[0].Env {
		if envVar.Name == "CONFIG_REDIS_ASYNC" {
			if envVar.Value == "1" {
				existing.Spec.Template.Spec.Containers[0].Env[envId].Value = "0"
				update = true
				return update, nil
			}
			update = false

		}
	}
	// Adds the env CONFIG_REDIS_ASYNC if not present
	if update {
		existing.Spec.Template.Spec.Containers[0].Env = append(existing.Spec.Template.Spec.Containers[0].Env,
			helper.EnvVarFromValue("CONFIG_REDIS_ASYNC", "0"))
	}
	return update, nil
}

func removeEnvVar(envVars []corev1.EnvVar, name string) []corev1.EnvVar {
	var newEnvVars []corev1.EnvVar
	for _, envVar := range envVars {
		if envVar.Name != name {
			newEnvVars = append(newEnvVars, envVar)
		}
	}
	return newEnvVars
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

func DeploymentSyncVolumesAndMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	changed := false

	// Ensure Volumes slice is initialized
	if existing.Spec.Template.Spec.Volumes == nil {
		existing.Spec.Template.Spec.Volumes = []corev1.Volume{}
	}

	// Add missing Volumes
	for _, desiredVolume := range desired.Spec.Template.Spec.Volumes {
		if !volumeExists(existing.Spec.Template.Spec.Volumes, desiredVolume.Name) {
			existing.Spec.Template.Spec.Volumes = append(existing.Spec.Template.Spec.Volumes, desiredVolume)
			changed = true
		}
	}

	// Sync VolumeMounts for Containers
	for cIdx := range existing.Spec.Template.Spec.Containers {
		updated, newVolumeMounts := syncVolumeMounts(existing.Spec.Template.Spec.Containers[cIdx].VolumeMounts, desired.Spec.Template.Spec.Containers[cIdx].VolumeMounts)
		if updated {
			existing.Spec.Template.Spec.Containers[cIdx].VolumeMounts = newVolumeMounts
			changed = true
		}
	}

	// Sync VolumeMounts for InitContainers
	for cIdx := range existing.Spec.Template.Spec.InitContainers {
		updated, newVolumeMounts := syncVolumeMounts(existing.Spec.Template.Spec.InitContainers[cIdx].VolumeMounts, desired.Spec.Template.Spec.InitContainers[cIdx].VolumeMounts)
		if updated {
			existing.Spec.Template.Spec.InitContainers[cIdx].VolumeMounts = newVolumeMounts
			changed = true
		}
	}

	return changed, nil
}

// Helper function: Check if a volume exists
func volumeExists(volumes []corev1.Volume, name string) bool {
	for _, v := range volumes {
		if v.Name == name {
			return true
		}
	}
	return false
}

// Helper function: Sync Volume Mounts (Add missing)
func syncVolumeMounts(existingMounts, desiredMounts []corev1.VolumeMount) (bool, []corev1.VolumeMount) {
	changed := false
	newVolumeMounts := existingMounts

	// Add missing VolumeMounts from desired
	for _, desiredMount := range desiredMounts {
		if !volumeMountExists(existingMounts, desiredMount.Name) {
			newVolumeMounts = append(newVolumeMounts, desiredMount)
			changed = true
		}
	}

	return changed, newVolumeMounts
}

// Helper function: Check if a volume mount exists
func volumeMountExists(volumeMounts []corev1.VolumeMount, name string) bool {
	for _, vm := range volumeMounts {
		if vm.Name == name {
			return true
		}
	}
	return false
}

func DeploymentRemoveTLSVolumesAndMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// system-database and zync database tls volume mount names in containers and init containers
	volumeNamesToRemove := []string{"writable-tls", "tls-secret"}

	if existing.Spec.Template.Spec.Volumes == nil {
		return false, nil
	}
	volumeModified := false
	// Remove volumes from the deployment spec
	for _, volumeName := range volumeNamesToRemove {
		for idx, volume := range existing.Spec.Template.Spec.Volumes {
			if volume.Name == volumeName {
				// Remove the specified volume
				existing.Spec.Template.Spec.Volumes = append(existing.Spec.Template.Spec.Volumes[:idx], existing.Spec.Template.Spec.Volumes[idx+1:]...)
				volumeModified = true
				break
			}
		}
	}
	// If volumes were removed, ensure volume mounts are also removed from containers
	if volumeModified {
		// For regular containers
		for cIdx, container := range existing.Spec.Template.Spec.Containers {
			for _, volumeName := range volumeNamesToRemove {
				for vIdx, volumeMount := range container.VolumeMounts {
					if volumeMount.Name == volumeName {
						// Remove the volume mount
						container.VolumeMounts = append(container.VolumeMounts[:vIdx], container.VolumeMounts[vIdx+1:]...)
						break
					}
				}
			}
			// Update the container spec with the modified volume mounts
			existing.Spec.Template.Spec.Containers[cIdx] = container
		}
		// For initContainers (if any)
		for cIdx, initContainer := range existing.Spec.Template.Spec.InitContainers {
			for _, volumeName := range volumeNamesToRemove {
				for vIdx, volumeMount := range initContainer.VolumeMounts {
					if volumeMount.Name == volumeName {
						// Remove the volume mount from initContainer
						initContainer.VolumeMounts = append(initContainer.VolumeMounts[:vIdx], initContainer.VolumeMounts[vIdx+1:]...)
						break
					}
				}
			}
			// Update the initContainer spec with the modified volume mounts
			existing.Spec.Template.Spec.InitContainers[cIdx] = initContainer
		}
	}
	// If no modifications were made, return false
	if !volumeModified {
		return false, nil
	}
	return true, nil
}

func DeploymentBackendRedisTLSRemoveEnvMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	tlsEnvVars := []string{
		"CONFIG_REDIS_CA_FILE",
		"CONFIG_REDIS_CERT",
		"CONFIG_REDIS_PRIVATE_KEY",
		"CONFIG_REDIS_SSL",
	}
	for _, varName := range tlsEnvVars {
		existing.Spec.Template.Spec.Containers[0].Env = removeEnvVar(existing.Spec.Template.Spec.Containers[0].Env, varName)
	}
	// Remove environment variables from InitContainer(s)
	for i := range existing.Spec.Template.Spec.InitContainers {
		for _, varName := range tlsEnvVars {
			existing.Spec.Template.Spec.InitContainers[i].Env = removeEnvVar(existing.Spec.Template.Spec.InitContainers[i].Env, varName)
		}
	}
	return true, nil
}

func DeploymentQueuesRedisTLSRemoveEnvMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	tlsEnvVars := []string{
		"CONFIG_QUEUES_CA_FILE",
		"CONFIG_QUEUES_CERT",
		"CONFIG_QUEUES_PRIVATE_KEY",
		"CONFIG_QUEUES_SSL",
	}
	for _, varName := range tlsEnvVars {
		existing.Spec.Template.Spec.Containers[0].Env = removeEnvVar(existing.Spec.Template.Spec.Containers[0].Env, varName)
	}
	// Remove environment variables from InitContainer(s)
	for i := range existing.Spec.Template.Spec.InitContainers {
		for _, varName := range tlsEnvVars {
			existing.Spec.Template.Spec.InitContainers[i].Env = removeEnvVar(existing.Spec.Template.Spec.InitContainers[i].Env, varName)
		}
	}
	return true, nil
}

func DeploymentSystemRedisTLSRemoveEnvMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	tlsEnvVars := []string{
		"REDIS_CA_FILE",
		"REDIS_CLIENT_CERT",
		"REDIS_PRIVATE_KEY",
		"REDIS_SSL",
		"BACKEND_REDIS_CA_FILE",
		"BACKEND_REDIS_CLIENT_CERT",
		"BACKEND_REDIS_PRIVATE_KEY",
		"BACKEND_REDIS_SSL",
	}
	for _, varName := range tlsEnvVars {
		existing.Spec.Template.Spec.Containers[0].Env = removeEnvVar(existing.Spec.Template.Spec.Containers[0].Env, varName)
	}
	return true, nil
}

func DeploymentSystemRedisTLSRemoveVolumesAndMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNamesToRemove := []string{"system-redis-tls", "backend-redis-tls"}
	return removeRedisTLSVolumesAndMounts(existing, volumeNamesToRemove)
}

func DeploymentBackendRedisTLSRemoveVolumesAndMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNamesToRemove := []string{"backend-redis-tls"}
	return removeRedisTLSVolumesAndMounts(existing, volumeNamesToRemove)
}

func DeploymentQueuesRedisTLSRemoveVolumesAndMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNamesToRemove := []string{"queues-redis-tls"}
	return removeRedisTLSVolumesAndMounts(existing, volumeNamesToRemove)
}

func removeRedisTLSVolumesAndMounts(existing *k8sappsv1.Deployment, volumeNamesToRemove []string) (bool, error) {
	if existing.Spec.Template.Spec.Volumes == nil {
		return false, nil
	}
	volumeModified := false
	for _, volumeName := range volumeNamesToRemove {
		for idx, volume := range existing.Spec.Template.Spec.Volumes {
			if volume.Name == volumeName {
				existing.Spec.Template.Spec.Volumes = append(existing.Spec.Template.Spec.Volumes[:idx], existing.Spec.Template.Spec.Volumes[idx+1:]...)
				volumeModified = true
				break
			}
		}
	}
	if volumeModified {
		for cIdx, container := range existing.Spec.Template.Spec.Containers {
			for _, volumeName := range volumeNamesToRemove {
				for vIdx, volumeMount := range container.VolumeMounts {
					if volumeMount.Name == volumeName {
						container.VolumeMounts = append(container.VolumeMounts[:vIdx], container.VolumeMounts[vIdx+1:]...)
						break
					}
				}
			}
			existing.Spec.Template.Spec.Containers[cIdx] = container
		}
		for cIdx, initContainer := range existing.Spec.Template.Spec.InitContainers {
			for _, volumeName := range volumeNamesToRemove {
				for vIdx, volumeMount := range initContainer.VolumeMounts {
					if volumeMount.Name == volumeName {
						initContainer.VolumeMounts = append(initContainer.VolumeMounts[:vIdx], initContainer.VolumeMounts[vIdx+1:]...)
						break
					}
				}
			}
			existing.Spec.Template.Spec.InitContainers[cIdx] = initContainer
		}
	}
	if !volumeModified {
		return false, nil
	}
	return true, nil
}

// Add volumes mutators for Redis TLS

func addRedistTLSVolumesAndMounts(existing *k8sappsv1.Deployment, volumeNamesToAdd []string, volumeMountsToAdd []v1.VolumeMount, items map[string][]v1.KeyToPath, secretName string) (bool, error) {
	volumeModified := false
	if existing.Spec.Template.Spec.Volumes == nil {
		existing.Spec.Template.Spec.Volumes = []v1.Volume{}
	}
	// Iterate through volume names to add
	for _, volumeName := range volumeNamesToAdd {
		// Check if the volume already exists, and find it to remove
		var volumeToRemove *v1.Volume
		for i, volume := range existing.Spec.Template.Spec.Volumes {
			if volume.Name == volumeName {
				volumeToRemove = &existing.Spec.Template.Spec.Volumes[i]
				break
			}
		}
		// If the volume exists, remove it from the list
		if volumeToRemove != nil {
			// Remove the volume by creating a new slice without the old volume
			existing.Spec.Template.Spec.Volumes = append(
				existing.Spec.Template.Spec.Volumes[:0],
				existing.Spec.Template.Spec.Volumes[1:]...,
			)
			volumeModified = true
		}
		// Create the new volume and add it
		var volumeToAdd v1.Volume
		volumeItems, exists := items[volumeName]
		if exists {
			volumeToAdd = v1.Volume{
				Name: volumeName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: secretName,
						Items:      volumeItems,
					},
				},
			}
			volumeModified = true
		}
		if volumeToAdd.Name != "" {
			existing.Spec.Template.Spec.Volumes = append(existing.Spec.Template.Spec.Volumes, volumeToAdd)
		}
	}
	// Add volume mounts to containers and init containers
	for cIdx, container := range existing.Spec.Template.Spec.Containers {
		for _, volumeMount := range volumeMountsToAdd {
			// Check if the volumeMount already exists, and remove it
			var mountToRemove *v1.VolumeMount
			for i, mount := range container.VolumeMounts {
				if mount.Name == volumeMount.Name {
					mountToRemove = &container.VolumeMounts[i]
					break
				}
			}
			// If the volumeMount exists, remove it
			if mountToRemove != nil {
				// Remove the volume mount
				container.VolumeMounts = append(
					container.VolumeMounts[:0],
					container.VolumeMounts[1:]...,
				)
				volumeModified = true
			}
			// Add the new volume mount
			container.VolumeMounts = append(container.VolumeMounts, volumeMount)
			volumeModified = true
		}
		// Update the container in the deployment spec
		existing.Spec.Template.Spec.Containers[cIdx] = container
	}
	// Add volume mounts to init containers
	for cIdx, initContainer := range existing.Spec.Template.Spec.InitContainers {
		for _, volumeMount := range volumeMountsToAdd {
			// Check if the volumeMount already exists, and remove it
			var mountToRemove *v1.VolumeMount
			for i, mount := range initContainer.VolumeMounts {
				if mount.Name == volumeMount.Name {
					mountToRemove = &initContainer.VolumeMounts[i]
					break
				}
			}
			// If the volumeMount exists, remove it
			if mountToRemove != nil {
				// Remove the volume mount
				initContainer.VolumeMounts = append(
					initContainer.VolumeMounts[:0],
					initContainer.VolumeMounts[1:]...,
				)
				volumeModified = true
			}
			// Add the new volume mount
			initContainer.VolumeMounts = append(initContainer.VolumeMounts, volumeMount)
			volumeModified = true
		}
		// Update the init container in the deployment spec
		existing.Spec.Template.Spec.InitContainers[cIdx] = initContainer
	}
	if volumeModified {
		return true, nil
	}
	return false, nil
}

func DeploymentSystemRedisTLSSyncVolumesAndMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNameToAdd := "system-redis-tls"
	volumeMountNameToAdd := "system-redis-tls"

	// Check if the volume already exists in the deployment
	if volumeExists(existing.Spec.Template.Spec.Volumes, volumeNameToAdd) {
		return false, nil // Volume already exists, no mutation needed
	}

	// Check if the volume mount already exists in the containers
	for _, container := range existing.Spec.Template.Spec.Containers {
		if volumeMountExists(container.VolumeMounts, volumeMountNameToAdd) {
			return false, nil // Volume mount already exists, no mutation needed
		}
	}

	// If volume and mount don't exist, we proceed to add them
	volumeMountsToAdd := []v1.VolumeMount{
		{
			Name:      volumeMountNameToAdd,
			ReadOnly:  false,
			MountPath: "/tls/system-redis",
		},
	}

	// Define the secret items to mount
	items := map[string][]v1.KeyToPath{
		volumeNameToAdd: {
			{
				Key:  "REDIS_SSL_CA",
				Path: "system-redis-ca.crt",
			},
			{
				Key:  "REDIS_SSL_CERT",
				Path: "system-redis-client.crt",
			},
			{
				Key:  "REDIS_SSL_KEY",
				Path: "system-redis-private.key",
			},
		},
	}

	secretName := "system-redis"

	// Call the addRedistTLSVolumesAndMounts helper to add the volume and volume mount
	return addRedistTLSVolumesAndMounts(existing, []string{volumeNameToAdd}, volumeMountsToAdd, items, secretName)
}

func DeploymentBackendRedisTLSSyncVolumesAndMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNameToAdd := "backend-redis-tls"
	volumeMountNameToAdd := "backend-redis-tls"

	// Check if the volume already exists in the deployment
	if volumeExists(existing.Spec.Template.Spec.Volumes, volumeNameToAdd) {
		return false, nil // Volume already exists, no mutation needed
	}

	// Check if the volume mount already exists in the containers
	for _, container := range existing.Spec.Template.Spec.Containers {
		if volumeMountExists(container.VolumeMounts, volumeMountNameToAdd) {
			return false, nil // Volume mount already exists, no mutation needed
		}
	}

	// If volume and mount don't exist, we proceed to add them
	volumeMountsToAdd := []v1.VolumeMount{
		{
			Name:      volumeMountNameToAdd,
			ReadOnly:  false,
			MountPath: "/tls/backend-redis",
		},
	}

	// Define the secret items to mount
	items := map[string][]v1.KeyToPath{
		volumeNameToAdd: {
			{
				Key:  "REDIS_SSL_CA",
				Path: "backend-redis-ca.crt",
			},
			{
				Key:  "REDIS_SSL_CERT",
				Path: "backend-redis-client.crt",
			},
			{
				Key:  "REDIS_SSL_KEY",
				Path: "backend-redis-private.key",
			},
		},
	}

	secretName := "backend-redis"

	// Call the addRedistTLSVolumesAndMounts helper to add the volume and volume mount
	return addRedistTLSVolumesAndMounts(existing, []string{volumeNameToAdd}, volumeMountsToAdd, items, secretName)
}

func DeploymentQueuesRedisTLSSyncVolumesAndMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNameToAdd := "queues-redis-tls"
	volumeMountNameToAdd := "queues-redis-tls"

	// Check if the volume already exists in the deployment
	if volumeExists(existing.Spec.Template.Spec.Volumes, volumeNameToAdd) {
		return false, nil // Volume already exists, no mutation needed
	}

	// Check if the volume mount already exists in the containers
	for _, container := range existing.Spec.Template.Spec.Containers {
		if volumeMountExists(container.VolumeMounts, volumeMountNameToAdd) {
			return false, nil // Volume mount already exists, no mutation needed
		}
	}

	// If volume and mount don't exist, we proceed to add them
	volumeMountsToAdd := []v1.VolumeMount{
		{
			Name:      volumeMountNameToAdd,
			ReadOnly:  false,
			MountPath: "/tls/queues",
		},
	}

	// Define the secret items to mount
	items := map[string][]v1.KeyToPath{
		volumeNameToAdd: {
			{
				Key:  "REDIS_SSL_QUEUES_CA",
				Path: "config-queues-ca.crt",
			},
			{
				Key:  "REDIS_SSL_QUEUES_CERT",
				Path: "config-queues-client.crt",
			},
			{
				Key:  "REDIS_SSL_QUEUES_KEY",
				Path: "config-queues-private.key",
			},
		},
	}

	secretName := "backend-redis"

	// Call the addRedistTLSVolumesAndMounts helper to add the volume and volume mount
	return addRedistTLSVolumesAndMounts(existing, []string{volumeNameToAdd}, volumeMountsToAdd, items, secretName)
}
