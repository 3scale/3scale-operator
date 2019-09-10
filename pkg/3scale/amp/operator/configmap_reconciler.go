package operator

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type ConfigMapReconciler interface {
	IsUpdateNeeded(desired, existing *v1.ConfigMap) bool
}

type ConfigMapBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler ConfigMapReconciler
}

func NewConfigMapBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler ConfigMapReconciler) *ConfigMapBaseReconciler {
	return &ConfigMapBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *ConfigMapBaseReconciler) Reconcile(desired *v1.ConfigMap) error {
	objectInfo := ObjectInfo(desired)
	existing := &v1.ConfigMap{}
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

func (r *ConfigMapBaseReconciler) isUpdateNeeded(desired, existing *v1.ConfigMap) (bool, error) {
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

type CreateOnlyConfigMapReconciler struct {
}

func NewCreateOnlyConfigMapReconciler() *CreateOnlyConfigMapReconciler {
	return &CreateOnlyConfigMapReconciler{}
}

func (r *CreateOnlyConfigMapReconciler) IsUpdateNeeded(desired, existing *v1.ConfigMap) bool {
	return false
}

func ConfigMapReconcileField(desired, existing *v1.ConfigMap, fieldName string) bool {
	updated := false

	if existingVal, ok := existing.Data[fieldName]; !ok {
		existing.Data[fieldName] = desired.Data[fieldName]
		updated = true
	} else {
		if desired.Data[fieldName] != existingVal {
			existing.Data[fieldName] = desired.Data[fieldName]
			updated = true
		}
	}
	return updated
}
