package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"
	hpa "k8s.io/api/autoscaling/v2"
)

func GenericHPAMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*hpa.HorizontalPodAutoscaler)
	if !ok {
		return false, fmt.Errorf("%T is not a *v2.HorizontalPodAutoscaler", existingObj)
	}
	desired, ok := desiredObj.(*hpa.HorizontalPodAutoscaler)
	if !ok {
		return false, fmt.Errorf("%T is not a *v2.HorizontalPodAutoscaler", desiredObj)
	}

	updated := false
	if !reflect.DeepEqual(desired.Spec, existing.Spec) {
		existing.Spec = desired.Spec
		updated = true
	}

	return updated, nil
}
