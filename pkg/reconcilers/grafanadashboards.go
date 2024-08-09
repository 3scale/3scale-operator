package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	grafanav1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
)

func GenericGrafanaDashboardsMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	var existingSpec, desiredSpec interface{}
	var existingType, desiredType string

	// Determine the type of the existing object
	switch existing := existingObj.(type) {
	case *grafanav1alpha1.GrafanaDashboard:
		existingSpec = existing.Spec
		existingType = "alpha"
	case *grafanav1beta1.GrafanaDashboard:
		existingSpec = existing.Spec
		existingType = "beta"
	default:
		return false, fmt.Errorf("%T is not a supported GrafanaDashboard type", existingObj)
	}

	// Determine the type of the desired object
	switch desired := desiredObj.(type) {
	case *grafanav1alpha1.GrafanaDashboard:
		desiredSpec = desired.Spec
		desiredType = "alpha"
	case *grafanav1beta1.GrafanaDashboard:
		desiredSpec = desired.Spec
		desiredType = "beta"
	default:
		return false, fmt.Errorf("%T is not a supported GrafanaDashboard type", desiredObj)
	}

	// Check if the types match
	if existingType != desiredType {
		return false, fmt.Errorf("mismatched types: existing is %s, desired is %s", existingType, desiredType)
	}

	updated := false

	// Compare and update specs
	if !reflect.DeepEqual(existingSpec, desiredSpec) {
		switch existing := existingObj.(type) {
		case *grafanav1alpha1.GrafanaDashboard:
			existing.Spec = desiredSpec.(grafanav1alpha1.GrafanaDashboardSpec)
		case *grafanav1beta1.GrafanaDashboard:
			existing.Spec = desiredSpec.(grafanav1beta1.GrafanaDashboardSpec)
		}
		updated = true
	}

	return updated, nil
}
