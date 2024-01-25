package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	k8sappsv1 "k8s.io/api/apps/v1"
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
