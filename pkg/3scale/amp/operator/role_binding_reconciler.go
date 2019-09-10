package operator

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type RoleBindingReconciler interface {
	IsUpdateNeeded(desired, existing *rbacv1.RoleBinding) bool
}

type RoleBindingBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler RoleBindingReconciler
}

func NewRoleBindingBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler RoleBindingReconciler) *RoleBindingBaseReconciler {
	return &RoleBindingBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *RoleBindingBaseReconciler) Reconcile(desired *rbacv1.RoleBinding) error {
	objectInfo := ObjectInfo(desired)
	existing := &rbacv1.RoleBinding{}
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

func (r *RoleBindingBaseReconciler) isUpdateNeeded(desired, existing *rbacv1.RoleBinding) (bool, error) {
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

type CreateOnlyRoleBindingReconciler struct {
}

func NewCreateOnlyRoleBindingReconciler() *CreateOnlyRoleBindingReconciler {
	return &CreateOnlyRoleBindingReconciler{}
}

func (r *CreateOnlyRoleBindingReconciler) IsUpdateNeeded(desired, existing *rbacv1.RoleBinding) bool {
	return false
}
