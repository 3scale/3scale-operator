package reconcilers

import (
	"fmt"
	"reflect"

	policyv1 "k8s.io/api/policy/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenericPDBMutator(existingObj, desiredObj client.Object) (bool, error) {
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
