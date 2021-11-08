package controllers

import (
	"encoding/json"
	"fmt"
	"reflect"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/google/go-cmp/cmp"
)

func (t *ProductThreescaleReconciler) syncPolicies(_ interface{}) error {
	existing, err := t.productEntity.Policies()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] policies: %w", t.resource.Spec.SystemName, err)
	}

	desired := t.convertResourcePolicies()

	// Compare Go unmarshalled objects (not byte arrays)
	// resilient to serialization differences like map key order differences or quotes.
	// Policies order matters. If order does not match, will be updated
	if !reflect.DeepEqual(desired, existing) {
		diff := cmp.Diff(desired, existing)
		t.logger.V(1).Info("syncPolicies", "policies not equal", diff)
		err = t.productEntity.UpdatePolicies(desired)
		if err != nil {
			return fmt.Errorf("Error sync product [%s] policies: %w", t.resource.Spec.SystemName, err)
		}
	}

	return nil
}

// Convert Policies from []capabilitiesv1beta1.PolicyConfig to *threescaleapi.PoliciesConfigList to be comparable
func (t *ProductThreescaleReconciler) convertResourcePolicies() *threescaleapi.PoliciesConfigList {
	policies := &threescaleapi.PoliciesConfigList{
		Policies: []threescaleapi.PolicyConfig{},
	}

	for _, crdPolicy := range t.resource.Spec.Policies {
		var configuration map[string]interface{}
		// CRD validation ensures no error happens
		// "configuration` type is object
		//properties:
		//  configuration:
		//    description: Configuration defines the policy configuration
		//    type: object
		//    x-kubernetes-preserve-unknown-fields: true
		_ = json.Unmarshal(crdPolicy.Configuration.Raw, &configuration)

		policies.Policies = append(policies.Policies, threescaleapi.PolicyConfig{
			Name:          crdPolicy.Name,
			Version:       crdPolicy.Version,
			Enabled:       crdPolicy.Enabled,
			Configuration: configuration,
		})
	}

	return policies
}
