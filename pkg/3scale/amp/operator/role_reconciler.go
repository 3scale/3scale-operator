package operator

import (
	"context"
	"fmt"
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type RoleReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewRoleReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) RoleReconciler {
	return RoleReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *RoleReconciler) Reconcile(desiredRole *rbacv1.Role) error {
	objectInfo := ObjectInfo(desiredRole)
	existingRole := &rbacv1.Role{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredRole.Name, Namespace: desiredRole.Namespace}, existingRole)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredRole)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureRole(existingRole, desiredRole)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating Role %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingRole)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Role %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *RoleReconciler) ensureRole(updated, desired *rbacv1.Role) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	if !reflect.DeepEqual(updated.Rules, desired.Rules) {
		updated.Rules = desired.Rules
	}

	return changed, nil
}
