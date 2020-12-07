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

type PolicyThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta1.Policy
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	providerAccountHost string
	logger              logr.Logger
}

func NewPolicyThreescaleReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.Policy, threescaleAPIClient *threescaleapi.ThreeScaleClient, providerAccountHost string, logger logr.Logger) *PolicyThreescaleReconciler {
	return &PolicyThreescaleReconciler{
		BaseReconciler:      b,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		providerAccountHost: providerAccountHost,
		logger:              logger.WithValues("3scale Reconciler", providerAccountHost),
	}
}

func (s *PolicyThreescaleReconciler) Reconcile() (*threescaleapi.APIcastPolicy, error) {
	s.logger.V(1).Info("START")

	var remotePolicy *threescaleapi.APIcastPolicy

	remotePolicies, err := s.threescaleAPIClient.ListAPIcastPolicies()
	if err != nil {
		return nil, err
	}

	for idx := range remotePolicies.Items {
		// Look for ID. If it does not exist, look for Name && Version
		foundByID := s.resource.Status.ID != nil && *remotePolicies.Items[idx].Element.ID == *s.resource.Status.ID
		foundByNameVersion := *remotePolicies.Items[idx].Element.Name == s.resource.Spec.Name && *remotePolicies.Items[idx].Element.Version == s.resource.Spec.Version
		if foundByID || foundByNameVersion {
			// found
			remotePolicy = &remotePolicies.Items[idx]
			break
		}
	}

	desiredConfiguration := json.RawMessage(s.resource.Spec.Schema.Configuration.Raw)

	if remotePolicy == nil {
		newPolicy := &threescaleapi.APIcastPolicy{
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

		remotePolicy, err = s.threescaleAPIClient.CreateAPIcastPolicy(newPolicy)
		if err != nil {
			return nil, err
		}
	}

	update := false
	// minimum required for update
	updatedPolicy := &threescaleapi.APIcastPolicy{
		Element: threescaleapi.APIcastPolicyItem{
			ID: remotePolicy.Element.ID,
			Schema: &threescaleapi.APIcastPolicySchema{
				Name:          &s.resource.Spec.Schema.Name,
				Summary:       &s.resource.Spec.Schema.Summary,
				Schema:        &s.resource.Spec.Schema.Schema,
				Version:       &s.resource.Spec.Schema.Version,
				Configuration: &desiredConfiguration,
			},
		},
	}

	if remotePolicy.Element.Name == nil || *remotePolicy.Element.Name != s.resource.Spec.Name {
		s.logger.V(1).Info("update Name", "Difference", cmp.Diff(remotePolicy.Element.Name, s.resource.Spec.Name))
		updatedPolicy.Element.Name = &s.resource.Spec.Name
		update = true
	}

	if remotePolicy.Element.Version == nil || *remotePolicy.Element.Version != s.resource.Spec.Version {
		s.logger.V(1).Info("update Version", "Difference", cmp.Diff(remotePolicy.Element.Version, s.resource.Spec.Version))
		updatedPolicy.Element.Version = &s.resource.Spec.Version
		update = true
	}
	if remotePolicy.Element.Schema != nil {
		if remotePolicy.Element.Schema.Name == nil || !reflect.DeepEqual(s.resource.Spec.Schema.Name, *remotePolicy.Element.Schema.Name) {
			s.logger.V(1).Info("update schema name", "Difference", cmp.Diff(remotePolicy.Element.Schema.Name, s.resource.Spec.Schema.Name))
			// No need to update, already in updatedPolicy structure
			update = true
		}

		if remotePolicy.Element.Schema.Version == nil || !reflect.DeepEqual(s.resource.Spec.Schema.Version, *remotePolicy.Element.Schema.Version) {
			s.logger.V(1).Info("update schema version", "Difference", cmp.Diff(remotePolicy.Element.Schema.Version, s.resource.Spec.Schema.Version))
			// No need to update, already in updatedPolicy structure
			update = true
		}

		if remotePolicy.Element.Schema.Schema == nil || !reflect.DeepEqual(s.resource.Spec.Schema.Schema, *remotePolicy.Element.Schema.Schema) {
			s.logger.V(1).Info("update schema $schema", "Difference", cmp.Diff(remotePolicy.Element.Schema.Schema, s.resource.Spec.Schema.Schema))
			// No need to update, already in updatedPolicy structure
			update = true
		}

		if remotePolicy.Element.Schema.Summary == nil || !reflect.DeepEqual(s.resource.Spec.Schema.Summary, *remotePolicy.Element.Schema.Summary) {
			s.logger.V(1).Info("update schema summary", "Difference", cmp.Diff(remotePolicy.Element.Schema.Summary, s.resource.Spec.Schema.Summary))
			// No need to update, already in updatedPolicy structure
			update = true
		}

		if s.resource.Spec.Schema.Description != nil && !reflect.DeepEqual(s.resource.Spec.Schema.Description, remotePolicy.Element.Schema.Description) {
			s.logger.V(1).Info("update schema description ", "Difference", cmp.Diff(remotePolicy.Element.Schema.Description, s.resource.Spec.Schema.Description))
			updatedPolicy.Element.Schema.Description = s.resource.Spec.Schema.Description
			update = true
		}

		if remotePolicy.Element.Schema.Configuration != nil {
			// Compare parsed configuration
			// Avoid detecting differences from serialization
			var existingConfGenericObj interface{}
			err := json.Unmarshal(*remotePolicy.Element.Schema.Configuration, &existingConfGenericObj)
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
				// No need to update, already in updatedPolicy structure
				update = true
			}
		} else {
			// Very unexpected scenario
			// remotePolicy.Element.Schema.Configuration is nill
			s.logger.V(1).Info("update schema to desired configuration, existing one is nil")
			// No need to update, already in updatedPolicy structure
			update = true
		}
	}

	if update {
		s.logger.V(1).Info("Desired Policy needs sync")
		remotePolicy, err = s.threescaleAPIClient.UpdateAPIcastPolicy(updatedPolicy)
		if err != nil {
			return nil, err
		}
	}

	return remotePolicy, nil
}
