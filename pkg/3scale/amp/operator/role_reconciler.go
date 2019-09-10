package operator

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/helper"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type RoleReconciler interface {
	IsUpdateNeeded(desired, existing *rbacv1.Role) bool
}

type RoleBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler RoleReconciler
}

func NewRoleBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler RoleReconciler) *RoleBaseReconciler {
	return &RoleBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *RoleBaseReconciler) Reconcile(desired *rbacv1.Role) error {
	objectInfo := ObjectInfo(desired)
	existing := &rbacv1.Role{}
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

func (r *RoleBaseReconciler) isUpdateNeeded(desired, existing *rbacv1.Role) (bool, error) {
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

type CreateOnlyRoleReconciler struct {
}

func NewCreateOnlyRoleReconciler() *CreateOnlyRoleReconciler {
	return &CreateOnlyRoleReconciler{}
}

func (r *CreateOnlyRoleReconciler) IsUpdateNeeded(desired, existing *rbacv1.Role) bool {
	return false
}
