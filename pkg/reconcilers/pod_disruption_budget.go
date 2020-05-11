package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"
	"k8s.io/api/policy/v1beta1"
)

func GenericPDBMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*v1beta1.PodDisruptionBudget)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1beta1.PodDisruptionBudget", existingObj)
	}
	desired, ok := desiredObj.(*v1beta1.PodDisruptionBudget)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1beta1.PodDisruptionBudget", desiredObj)
	}

	updated := false
	if !reflect.DeepEqual(desired.Spec, existing.Spec) {
		existing.Spec = desired.Spec
		updated = true
	}

	return updated, nil
}
