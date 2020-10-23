package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"

	"github.com/google/go-cmp/cmp"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
)

func GenericGrafanaDashboardsMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*grafanav1alpha1.GrafanaDashboard)
	if !ok {
		return false, fmt.Errorf("%T is not a *grafanav1alpha1.GrafanaDashboard", existingObj)
	}
	desired, ok := desiredObj.(*grafanav1alpha1.GrafanaDashboard)
	if !ok {
		return false, fmt.Errorf("%T is not a *grafanav1alpha1.GrafanaDashboard", desiredObj)
	}

	updated := false

	if !reflect.DeepEqual(existing.Spec, desired.Spec) {
		diff := cmp.Diff(existing.Spec, desired.Spec)
		log.V(1).Info(fmt.Sprintf("%s spec has changed: %s", common.ObjectInfo(desired), diff))
		existing.Spec = desired.Spec
		updated = true
	}

	return updated, nil
}
