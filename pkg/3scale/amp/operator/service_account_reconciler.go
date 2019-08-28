package operator

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type ServiceAccountReconciler interface {
	IsUpdateNeeded(desired, existing *v1.ServiceAccount) bool
}

type ServiceAccountBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler ServiceAccountReconciler
}

func NewServiceAccountBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler ServiceAccountReconciler) *ServiceAccountBaseReconciler {
	return &ServiceAccountBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *ServiceAccountBaseReconciler) Reconcile(desired *v1.ServiceAccount) error {
	objectInfo := ObjectInfo(desired)
	existing := &v1.ServiceAccount{}
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

func (r *ServiceAccountBaseReconciler) isUpdateNeeded(desired, existing *v1.ServiceAccount) (bool, error) {
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

type CreateOnlyServiceAccountReconciler struct {
}

func NewCreateOnlyServiceAccountReconciler() *CreateOnlyServiceAccountReconciler {
	return &CreateOnlyServiceAccountReconciler{}
}

func (r *CreateOnlyServiceAccountReconciler) IsUpdateNeeded(desired, existing *v1.ServiceAccount) bool {
	return false
}
