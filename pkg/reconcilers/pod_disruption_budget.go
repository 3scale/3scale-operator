package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"
	policyv1 "k8s.io/api/policy/v1"
)

func GenericPDBMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*policyv1.PodDisruptionBudget)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.PodDisruptionBudget", existingObj)
	}
	desired, ok := desiredObj.(*policyv1.PodDisruptionBudget)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.PodDisruptionBudget", desiredObj)
	}

	updated := false
	if !reflect.DeepEqual(desired.Spec, existing.Spec) {
		existing.Spec = desired.Spec
		updated = true
	}

	return updated, nil
}
