package operator

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type PVCReconciler interface {
	IsUpdateNeeded(desired, existing *v1.PersistentVolumeClaim) bool
}

type PVCBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler PVCReconciler
}

func NewPVCBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler PVCReconciler) *PVCBaseReconciler {
	return &PVCBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *PVCBaseReconciler) Reconcile(desired *v1.PersistentVolumeClaim) error {
	objectInfo := ObjectInfo(desired)
	existing := &v1.PersistentVolumeClaim{}
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

func (r *PVCBaseReconciler) isUpdateNeeded(desired, existing *v1.PersistentVolumeClaim) (bool, error) {
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

type CreateOnlyPVCReconciler struct {
}

func NewCreateOnlyPVCReconciler() *CreateOnlyPVCReconciler {
	return &CreateOnlyPVCReconciler{}
}

func (r *CreateOnlyPVCReconciler) IsUpdateNeeded(desired, existing *v1.PersistentVolumeClaim) bool {
	return false
}
