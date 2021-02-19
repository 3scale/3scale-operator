package controllers

import (
	"encoding/json"
	"reflect"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
)

type CustomPolicyDefinitionThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta1.CustomPolicyDefinition
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	providerAccountHost string
	logger              logr.Logger
}

func NewCustomPolicyDefinitionThreescaleReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.CustomPolicyDefinition, threescaleAPIClient *threescaleapi.ThreeScaleClient, providerAccountHost string, logger logr.Logger) *CustomPolicyDefinitionThreescaleReconciler {
	return &CustomPolicyDefinitionThreescaleReconciler{
		BaseReconciler:      b,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		providerAccountHost: providerAccountHost,
		logger:              logger.WithValues("3scale Reconciler", providerAccountHost),
	}
}

func (s *CustomPolicyDefinitionThreescaleReconciler) Reconcile() (*threescaleapi.APIcastPolicy, error) {
	s.logger.V(1).Info("START")

	var remoteCustomPolicy *threescaleapi.APIcastPolicy

	remoteCustomPolicies, err := s.threescaleAPIClient.ListAPIcastPolicies()
	if err != nil {
		return nil, err
	}

	for idx := range remoteCustomPolicies.Items {
		// Look for ID. If it does not exist, look for Name && Version
		foundByID := s.resource.Status.ID != nil && *remoteCustomPolicies.Items[idx].Element.ID == *s.resource.Status.ID
		foundByNameVersion := *remoteCustomPolicies.Items[idx].Element.Name == s.resource.Spec.Name && *remoteCustomPolicies.Items[idx].Element.Version == s.resource.Spec.Version
		if foundByID || foundByNameVersion {
			// found
			remoteCustomPolicy = &remoteCustomPolicies.Items[idx]
			break
		}
	}

	desiredConfiguration := json.RawMessage(s.resource.Spec.Schema.Configuration.Raw)

	if remoteCustomPolicy == nil {
		newCustomPolicy := &threescaleapi.APIcastPolicy{
			Element: threescaleapi.APIcastPolicyItem{
				Name:    &s.resource.Spec.Name,
				Version: &s.resource.Spec.Version,
				Schema: &threescaleapi.APIcastPolicySchema{
					Name:          &s.resource.Spec.Schema.Name,
					Summary:       &s.resource.Spec.Schema.Summary,
					Schema:        &s.resource.Spec.Schema.Schema,
					Version:       &s.resource.Spec.Schema.Version,
					Configuration: &desiredConfiguration,
				},
			},
		}

		remoteCustomPolicy, err = s.threescaleAPIClient.CreateAPIcastPolicy(newCustomPolicy)
		if err != nil {
			return nil, err
		}
	}

	update := false
	// minimum required for update
	updatedCustomPolicy := &threescaleapi.APIcastPolicy{
		Element: threescaleapi.APIcastPolicyItem{
			ID: remoteCustomPolicy.Element.ID,
			Schema: &threescaleapi.APIcastPolicySchema{
				Name:          &s.resource.Spec.Schema.Name,
				Summary:       &s.resource.Spec.Schema.Summary,
				Schema:        &s.resource.Spec.Schema.Schema,
				Version:       &s.resource.Spec.Schema.Version,
				Configuration: &desiredConfiguration,
			},
		},
	}

	if remoteCustomPolicy.Element.Name == nil || *remoteCustomPolicy.Element.Name != s.resource.Spec.Name {
		s.logger.V(1).Info("update Name", "Difference", cmp.Diff(remoteCustomPolicy.Element.Name, s.resource.Spec.Name))
		updatedCustomPolicy.Element.Name = &s.resource.Spec.Name
		update = true
	}

	if remoteCustomPolicy.Element.Version == nil || *remoteCustomPolicy.Element.Version != s.resource.Spec.Version {
		s.logger.V(1).Info("update Version", "Difference", cmp.Diff(remoteCustomPolicy.Element.Version, s.resource.Spec.Version))
		updatedCustomPolicy.Element.Version = &s.resource.Spec.Version
		update = true
	}
	if remoteCustomPolicy.Element.Schema != nil {
		if remoteCustomPolicy.Element.Schema.Name == nil || !reflect.DeepEqual(s.resource.Spec.Schema.Name, *remoteCustomPolicy.Element.Schema.Name) {
			s.logger.V(1).Info("update schema name", "Difference", cmp.Diff(remoteCustomPolicy.Element.Schema.Name, s.resource.Spec.Schema.Name))
			// No need to update, already in updatedPolicy structure
			update = true
		}

		if remoteCustomPolicy.Element.Schema.Version == nil || !reflect.DeepEqual(s.resource.Spec.Schema.Version, *remoteCustomPolicy.Element.Schema.Version) {
			s.logger.V(1).Info("update schema version", "Difference", cmp.Diff(remoteCustomPolicy.Element.Schema.Version, s.resource.Spec.Schema.Version))
			// No need to update, already in updatedPolicy structure
			update = true
		}

		if remoteCustomPolicy.Element.Schema.Schema == nil || !reflect.DeepEqual(s.resource.Spec.Schema.Schema, *remoteCustomPolicy.Element.Schema.Schema) {
			s.logger.V(1).Info("update schema $schema", "Difference", cmp.Diff(remoteCustomPolicy.Element.Schema.Schema, s.resource.Spec.Schema.Schema))
			// No need to update, already in updatedPolicy structure
			update = true
		}

		if remoteCustomPolicy.Element.Schema.Summary == nil || !reflect.DeepEqual(s.resource.Spec.Schema.Summary, *remoteCustomPolicy.Element.Schema.Summary) {
			s.logger.V(1).Info("update schema summary", "Difference", cmp.Diff(remoteCustomPolicy.Element.Schema.Summary, s.resource.Spec.Schema.Summary))
			// No need to update, already in updatedPolicy structure
			update = true
		}

		if s.resource.Spec.Schema.Description != nil && !reflect.DeepEqual(s.resource.Spec.Schema.Description, remoteCustomPolicy.Element.Schema.Description) {
			s.logger.V(1).Info("update schema description ", "Difference", cmp.Diff(remoteCustomPolicy.Element.Schema.Description, s.resource.Spec.Schema.Description))
			updatedCustomPolicy.Element.Schema.Description = s.resource.Spec.Schema.Description
			update = true
		}

		if remoteCustomPolicy.Element.Schema.Configuration != nil {
			// Compare parsed configuration
			// Avoid detecting differences from serialization
			var existingConfGenericObj interface{}
			err := json.Unmarshal(*remoteCustomPolicy.Element.Schema.Configuration, &existingConfGenericObj)
			if err != nil {
				return nil, err
			}

			var desiredConfGenericObj interface{}
			err = json.Unmarshal(desiredConfiguration, &desiredConfGenericObj)
			if err != nil {
				return nil, err
			}

			if !reflect.DeepEqual(existingConfGenericObj, desiredConfGenericObj) {
				s.logger.V(1).Info("update schema configuration", "Difference", cmp.Diff(existingConfGenericObj, desiredConfGenericObj))
				// No need to update, already in updatedCustomPolicy structure
				update = true
			}
		} else {
			// Very unexpected scenario
			// remoteCustomPolicy.Element.Schema.Configuration is nill
			s.logger.V(1).Info("update schema to desired configuration, existing one is nil")
			// No need to update, already in updatedCustomPolicy structure
			update = true
		}
	}

	if update {
		s.logger.V(1).Info("Desired CustomPolicy needs sync")
		remoteCustomPolicy, err = s.threescaleAPIClient.UpdateAPIcastPolicy(updatedCustomPolicy)
		if err != nil {
			return nil, err
		}
	}

	return remoteCustomPolicy, nil
}
