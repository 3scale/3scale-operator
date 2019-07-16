package operator

import (
	"context"
	"fmt"
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type RoleBindingReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewRoleBindingReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) RoleBindingReconciler {
	return RoleBindingReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *RoleBindingReconciler) Reconcile(desiredRoleBinding *rbacv1.RoleBinding) error {
	objectInfo := ObjectInfo(desiredRoleBinding)
	existingRoleBinding := &rbacv1.RoleBinding{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredRoleBinding.Name, Namespace: desiredRoleBinding.Namespace}, existingRoleBinding)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredRoleBinding)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureRole(existingRoleBinding, desiredRoleBinding)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating RoleBinding %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingRoleBinding)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating RoleBinding %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *RoleBindingReconciler) ensureRole(updated, desired *rbacv1.RoleBinding) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	if !reflect.DeepEqual(updated.Subjects, desired.Subjects) {
		updated.Subjects = desired.Subjects
	}

	if !reflect.DeepEqual(updated.RoleRef, desired.RoleRef) {
		updated.RoleRef = desired.RoleRef
	}

	return changed, nil
}
