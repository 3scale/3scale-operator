package reconcilers

import (
	"fmt"
	"github.com/3scale/3scale-operator/pkg/common"
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
)

func RoleRuleMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
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
