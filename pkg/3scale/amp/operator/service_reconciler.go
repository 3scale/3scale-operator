package operator

import (
	"context"

	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type ServiceReconciler interface {
	IsUpdateNeeded(desired, existing *v1.Service) bool
}

type ServiceBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler ServiceReconciler
}

func NewServiceBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler ServiceReconciler) *ServiceBaseReconciler {
	return &ServiceBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *ServiceBaseReconciler) Reconcile(desired *v1.Service) error {
	existing := &v1.Service{}
	err := r.Client().Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: r.apiManager.GetNamespace()},
		existing)

	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	if IsObjectTaggedTorDelete(desired) {
		if !apierrors.IsNotFound(err) {
			return r.deleteResource(existing)
		}
		// if not found, nothing else to do
		return nil
	}

	if apierrors.IsNotFound(err) {
		return r.createResource(desired)
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

func (r *ServiceBaseReconciler) isUpdateNeeded(desired, existing *v1.Service) (bool, error) {
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

type CreateOnlySvcReconciler struct {
}

func NewCreateOnlySvcReconciler() *CreateOnlySvcReconciler {
	return &CreateOnlySvcReconciler{}
}

func (r *CreateOnlySvcReconciler) IsUpdateNeeded(desired, existing *v1.Service) bool {
	return false
}
