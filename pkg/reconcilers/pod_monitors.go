package reconcilers

import (
	"fmt"
	"reflect"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenericPodMonitorMutator(existingObj, desiredObj client.Object) (bool, error) {
	existing, ok := existingObj.(*monitoringv1.PodMonitor)
	if !ok {
		return false, fmt.Errorf("%T is not a *monitoringv1.PodMonitor", existingObj)
	}
	desired, ok := desiredObj.(*monitoringv1.PodMonitor)
	if !ok {
		return false, fmt.Errorf("%T is not a *monitoringv1.PodMonitor", desiredObj)
	}

	updated := false
	if !reflect.DeepEqual(desired.Spec, existing.Spec) {
		existing.Spec = desired.Spec
		updated = true
	}

	return updated, nil
}
