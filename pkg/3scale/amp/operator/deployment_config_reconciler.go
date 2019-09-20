package operator

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

type DeploymentConfigReconciler interface {
	IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool
}

type DeploymentConfigBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler DeploymentConfigReconciler
}

func NewDeploymentConfigBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler DeploymentConfigReconciler) *DeploymentConfigBaseReconciler {
	return &DeploymentConfigBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *DeploymentConfigBaseReconciler) Reconcile(desired *appsv1.DeploymentConfig) error {
	objectInfo := ObjectInfo(desired)
	existing := &appsv1.DeploymentConfig{}
	err := r.Client().Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: r.apiManager.GetNamespace()},
		existing)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.createResource(desired)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			return nil
		}
		return err
	}

	update, err := r.isUpdateNeeded(desired, existing)
	if err != nil {
		return err
	}

	if update {
		return r.updateResource(existing)
	}

	return nil
}

func (r *DeploymentConfigBaseReconciler) isUpdateNeeded(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	updated := helper.EnsureObjectMeta(&existing.ObjectMeta, &desired.ObjectMeta)

	updatedTmp, err := r.ensureOwnerReference(existing)
	if err != nil {
		return false, nil
	}

	updated = updated || updatedTmp

	updatedTmp = r.reconciler.IsUpdateNeeded(desired, existing)
	updated = updated || updatedTmp

	return updated, nil
}

func DeploymentConfigReconcileContainerResources(desired, existing *appsv1.DeploymentConfig, logger logr.Logger) bool {
	desiredName := ObjectInfo(desired)
	update := false

	//
	// Check container resource requirements
	//
	if len(desired.Spec.Template.Spec.Containers) != 1 {
		panic(fmt.Sprintf("%s desired spec.template.spec.containers length changed to '%d', should be 1", desiredName, len(desired.Spec.Template.Spec.Containers)))
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		logger.Info(fmt.Sprintf("%s spec.template.spec.containers length changed to '%d', recreating dc", desiredName, len(existing.Spec.Template.Spec.Containers)))
		existing.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		update = true
	}

	if !helper.CmpResources(&existing.Spec.Template.Spec.Containers[0].Resources, &desired.Spec.Template.Spec.Containers[0].Resources) {
		diff := cmp.Diff(existing.Spec.Template.Spec.Containers[0].Resources, desired.Spec.Template.Spec.Containers[0].Resources, cmpopts.IgnoreUnexported(resource.Quantity{}))
		logger.Info(fmt.Sprintf("%s spec.template.spec.containers[0].resources have changed: %s", desiredName, diff))
		existing.Spec.Template.Spec.Containers[0].Resources = desired.Spec.Template.Spec.Containers[0].Resources
		update = true
	}

	return update
}

func DeploymentConfigReconcileReplicas(desired, existing *appsv1.DeploymentConfig, logger logr.Logger) bool {
	desiredName := ObjectInfo(desired)
	update := false

	if desired.Spec.Replicas != existing.Spec.Replicas {
		logger.Info(fmt.Sprintf("%s spec.replicas differs, recreating dc", desiredName))
		existing.Spec.Replicas = desired.Spec.Replicas
		update = true
	}

	return update
}

type CreateOnlyDCReconciler struct {
}

func NewCreateOnlyDCReconciler() *CreateOnlyDCReconciler {
	return &CreateOnlyDCReconciler{}
}

func (r *CreateOnlyDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	return false
}
