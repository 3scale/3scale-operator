package controllers

import (
	"encoding/json"
	"fmt"
	"reflect"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *ProductThreescaleReconciler) syncPolicies(_ interface{}) error {
	existing, err := t.productEntity.Policies()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] policies: %w", t.resource.Spec.SystemName, err)
	}

	desired, err := t.convertResourcePolicies()
	if err != nil {
		return err
	}

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
func (t *ProductThreescaleReconciler) convertResourcePolicies() (*threescaleapi.PoliciesConfigList, error) {
	policies := &threescaleapi.PoliciesConfigList{
		Policies: []threescaleapi.PolicyConfig{},
	}

	for _, crdPolicy := range t.resource.Spec.Policies {
		configuration, err := t.convertPolicyConfiguration(crdPolicy)
		if err != nil {
			return nil, err
		}

		policies.Policies = append(policies.Policies, threescaleapi.PolicyConfig{
			Name:          crdPolicy.Name,
			Version:       crdPolicy.Version,
			Enabled:       crdPolicy.Enabled,
			Configuration: configuration,
		})
	}

	return policies, nil
}

func (t *ProductThreescaleReconciler) convertPolicyConfiguration(crdPolicy capabilitiesv1beta1.PolicyConfig) (map[string]interface{}, error) {
	configuration := map[string]interface{}{}

	// If plain value is not the default - use plain value as precedence over secret
	if string(crdPolicy.Configuration.Raw) != capabilitiesv1beta1.ProductPolicyConfigurationDefault {
		// CRD validation ensures no error happens
		// "configuration` type is object
		// properties:
		//  configuration:
		//    description: Configuration defines the policy configuration
		//    type: object
		//    x-kubernetes-preserve-unknown-fields: true
		_ = json.Unmarshal(crdPolicy.Configuration.Raw, &configuration)

		return configuration, nil
	}

	// If policy is defined in secretRef
	if crdPolicy.ConfigurationRef.Name != "" {
		// Get configuration from secret reference
		secret := &corev1.Secret{}
		namespace := t.resource.Namespace
		if crdPolicy.ConfigurationRef.Namespace != "" {
			namespace = crdPolicy.ConfigurationRef.Namespace
		}

		if err := t.Client().Get(t.Context(), types.NamespacedName{Name: crdPolicy.ConfigurationRef.Name, Namespace: namespace}, secret); err != nil {
			return nil, err
		}

		configurationByteArray, ok := secret.Data[capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField]
		if !ok {
			return nil, fmt.Errorf("not found configuration field in secret (ns: %s, name: %s) field: %s",
				namespace, crdPolicy.ConfigurationRef.Name, capabilitiesv1beta1.ProductPolicyConfigurationPasswordSecretField)
		}

		if err := json.Unmarshal(configurationByteArray, &configuration); err != nil {
			return nil, err
		}
	}

	return configuration, nil
}
