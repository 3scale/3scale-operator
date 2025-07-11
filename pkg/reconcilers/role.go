package reconcilers

import (
	"fmt"
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RoleRuleMutator(existingObj, desiredObj client.Object) (bool, error) {
	existing, ok := existingObj.(*rbacv1.Role)
	if !ok {
		return false, fmt.Errorf("%T is not a *rbacv1.Role", existingObj)
	}
	desired, ok := desiredObj.(*rbacv1.Role)
	if !ok {
		return false, fmt.Errorf("%T is not a *rbacv1.Role", desiredObj)
	}

	updated := false

	if !reflect.DeepEqual(existing.Rules, desired.Rules) {
		existing.Rules = desired.Rules
		updated = true
	}

	return updated, nil
}
