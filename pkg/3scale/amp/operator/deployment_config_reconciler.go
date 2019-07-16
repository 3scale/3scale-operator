package operator

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

type DeploymentConfigReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

type APIManagerDeploymentConfigReconciler struct {
	DeploymentConfigReconciler
}

func NewDeploymentConfigReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) DeploymentConfigReconciler {
	return DeploymentConfigReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *DeploymentConfigReconciler) Reconcile(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	// We assume desired namespace and ownerref are set already
	objectInfo := ObjectInfo(desiredDeploymentConfig)
	existingDeploymentConfig := &appsv1.DeploymentConfig{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredDeploymentConfig.Name, Namespace: desiredDeploymentConfig.Namespace}, existingDeploymentConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredDeploymentConfig)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	// Set in the desired some fields that are automatically set
	// by Kubernetes controllers as defaults that are not defined in
	// our logic
	// TODO

	// reconcile update
	needsUpdate, err := r.ensureDeploymentConfig(existingDeploymentConfig, desiredDeploymentConfig)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating DeploymentConfig %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingDeploymentConfig)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating DeploymentConfig %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *DeploymentConfigReconciler) ensureDeploymentConfig(updated, desired *appsv1.DeploymentConfig) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// Reconcile PodTemplateSpec labels
	if r.ObjectMetaMerger.EnsureLabels(updated.Spec.Template.ObjectMeta.Labels, desired.Spec.Template.ObjectMeta.Labels) {
		changed = true
	}

	r.ensurePodContainers(updated.Spec.Template.Spec.Containers, desired.Spec.Template.Spec.Containers)
	if !reflect.DeepEqual(updated.Spec.Template.Spec.Containers, desired.Spec.Template.Spec.Containers) {
		fmt.Println(cmp.Diff(updated.Spec.Template.Spec.Containers, desired.Spec.Template.Spec.Containers, cmpopts.IgnoreUnexported(resource.Quantity{})))
		updated.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		changed = true
	}

	r.ensurePodContainers(updated.Spec.Template.Spec.InitContainers, desired.Spec.Template.Spec.InitContainers)
	if !reflect.DeepEqual(updated.Spec.Template.Spec.InitContainers, desired.Spec.Template.Spec.InitContainers) {
		fmt.Println(cmp.Diff(updated.Spec.Template.Spec.InitContainers, desired.Spec.Template.Spec.InitContainers, cmpopts.IgnoreUnexported(resource.Quantity{})))
		updated.Spec.Template.Spec.InitContainers = desired.Spec.Template.Spec.InitContainers
		changed = true
	}

	r.ensureDeploymentConfigStrategy(&updated.Spec.Strategy, &desired.Spec.Strategy)
	if !reflect.DeepEqual(updated.Spec.Strategy, desired.Spec.Strategy) {
		fmt.Println(cmp.Diff(updated.Spec.Strategy, desired.Spec.Strategy, cmpopts.IgnoreUnexported(resource.Quantity{})))
		updated.Spec.Strategy = desired.Spec.Strategy
		changed = true
	}

	return changed, nil
}

func (r *DeploymentConfigReconciler) ensureDeploymentConfigStrategy(updated, desired *appsv1.DeploymentStrategy) {
	// When no DeploymentConfig Strategy Type is defined OpenShift defines a default one
	// We make sure we have that into desired in that case so the diff does not
	// detect that in the comparison
	if desired.Type == appsv1.DeploymentStrategyType("") {
		*desired = *updated.DeepCopy()
	}

	if desired.Type == appsv1.DeploymentStrategyTypeRolling && desired.RollingParams == nil {
		desired.RollingParams = updated.RollingParams
	}

	if desired.Type == appsv1.DeploymentStrategyTypeRecreate && desired.RecreateParams == nil {
		desired.RecreateParams = updated.RecreateParams
	}

	if desired.Type == appsv1.DeploymentStrategyTypeCustom && desired.CustomParams == nil {
		desired.CustomParams = updated.CustomParams
	}

	if desired.ActiveDeadlineSeconds == nil {
		desired.ActiveDeadlineSeconds = updated.ActiveDeadlineSeconds
	}
}

func (r *DeploymentConfigReconciler) ensurePodContainers(updated, desired []v1.Container) {
	updatedContainerMap := map[string]*v1.Container{}
	for idx := range updated {
		container := &updated[idx]
		updatedContainerMap[container.Name] = container
	}

	for idx := range desired {
		desiredContainer := &desired[idx]
		if updatedContainer, ok := updatedContainerMap[desiredContainer.Name]; ok {
			desiredContainer.Image = updatedContainer.Image

			if desiredContainer.TerminationMessagePath == "" {
				desiredContainer.TerminationMessagePath = updatedContainer.TerminationMessagePath
			}

			if desiredContainer.TerminationMessagePolicy == v1.TerminationMessagePolicy("") {
				desiredContainer.TerminationMessagePolicy = updatedContainer.TerminationMessagePolicy
			}

			if desiredContainer.ImagePullPolicy == v1.PullPolicy("") {
				desiredContainer.ImagePullPolicy = updatedContainer.ImagePullPolicy
			}

			if desiredContainer.ReadinessProbe != nil && updatedContainer.ReadinessProbe != nil {
				r.ensureProbe(updatedContainer.ReadinessProbe, desiredContainer.ReadinessProbe)
			}
			if desiredContainer.LivenessProbe != nil && updatedContainer.LivenessProbe != nil {
				r.ensureProbe(updatedContainer.LivenessProbe, desiredContainer.LivenessProbe)
			}

			r.ensureResourceRequirements(&updatedContainer.Resources, &desiredContainer.Resources)

		}
	}

	sort.Slice(updated, func(i, j int) bool { return updated[i].Name < updated[j].Name })
	sort.Slice(desired, func(i, j int) bool { return desired[i].Name < desired[j].Name })
}

func (r *DeploymentConfigReconciler) ensureResourceRequirements(updated, desired *v1.ResourceRequirements) {
	r.ensureResourceList(updated.Limits, desired.Limits)
	r.ensureResourceList(updated.Requests, desired.Requests)
}

// This method assigns the value of desired to updated when the resource quantities
// are not nil in both sides and are equal. The reason for this is
// because although they are equal internally are stored differently (the units
// might be expressed in a different way) and executing DeepEqual returns
// different even though they are "logically" equal
func (r *DeploymentConfigReconciler) ensureResourceList(updated, desired v1.ResourceList) {
	if !desired.Cpu().IsZero() && !updated.Cpu().IsZero() &&
		desired.Cpu().Cmp(*updated.Cpu()) == 0 {
		desired[v1.ResourceCPU] = *updated.Cpu()
	}
	if !desired.Memory().IsZero() && !updated.Memory().IsZero() &&
		desired.Memory().Cmp(*updated.Memory()) == 0 {
		desired[v1.ResourceMemory] = *updated.Memory()
	}
	if !desired.Pods().IsZero() && !updated.Pods().IsZero() &&
		desired.Pods().Cmp(*updated.Pods()) == 0 {
		desired[v1.ResourcePods] = *updated.Pods()
	}
	if !desired.StorageEphemeral().IsZero() && !updated.StorageEphemeral().IsZero() &&
		desired.StorageEphemeral().Cmp(*updated.StorageEphemeral()) == 0 {
		desired[v1.ResourceEphemeralStorage] = *updated.StorageEphemeral()
	}
}

func (r *DeploymentConfigReconciler) ensureProbe(updated, desired *v1.Probe) {
	if desired.TimeoutSeconds == 0 {
		desired.TimeoutSeconds = updated.TimeoutSeconds
	}

	if desired.PeriodSeconds == 0 {
		desired.PeriodSeconds = updated.PeriodSeconds
	}

	if desired.SuccessThreshold == 0 {
		desired.SuccessThreshold = updated.SuccessThreshold
	}

	if desired.FailureThreshold == 0 {
		desired.FailureThreshold = updated.FailureThreshold
	}

	if desired.Handler.HTTPGet != nil && updated.Handler.HTTPGet != nil {
		if desired.Handler.HTTPGet.Scheme == v1.URIScheme("") {
			desired.Handler.HTTPGet.Scheme = v1.URIScheme("HTTP")
		}
	}
}
