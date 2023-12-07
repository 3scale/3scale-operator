package controllers

import (
	"fmt"
	"strconv"
	"strings"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

func (t *ProductThreescaleReconciler) syncMappingRules(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.MappingRules))
	desiredMap := map[string]capabilitiesv1beta1.MappingRuleSpec{}
	for _, spec := range t.resource.Spec.MappingRules {
		metricID, err := t.productEntity.FindMethodMetricIDBySystemName(spec.MetricMethodRef)
		if err != nil {
			return fmt.Errorf("error syncMappingRules: %w", err)
		}
		key := fmt.Sprintf("%s:%s:%s", spec.HTTPMethod, spec.Pattern, strconv.FormatInt(metricID, 10))
		desiredKeys = append(desiredKeys, key)
		desiredMap[key] = spec
	}

	existingMap, err := t.getExistingMappingRules()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
	}
	existingKeys := make([]string, 0, len(existingMap))
	for existingKey := range existingMap {
		existingKeys = append(existingKeys, existingKey)
	}

	//
	// Deleted existing and not desired mapping rules
	//
	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	t.logger.V(1).Info("syncMappingRules", "notDesiredExistingKeys", notDesiredExistingKeys)
	notDesiredList := make([]threescaleapi.MappingRuleItem, 0, len(notDesiredExistingKeys))
	for _, key := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingKeys set
		notDesiredList = append(notDesiredList, existingMap[key])
	}
	err = t.processNotDesiredMappingRules(notDesiredList)
	if err != nil {
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
	}

	// If existing non-desired mapping rules have been detected we refetch
	// the existing list after deletion to get a more consistent view
	// of the existing rules attributes after the deletion.
	if len(notDesiredList) > 0 {
		existingMap, err = t.getExistingMappingRules()
		if err != nil {
			return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
		}
	}

	// Reconcile desired mapping rules
	// In order of definition in the custom resource. Create or update
	// the MappingRule being processed depending on whether it already exists
	// in the 3scale API. Specified 'position' attribute of the MappingRule
	// always corresponds to the position in the CR's MappingRules array.
	// Even though when creating/updating a MappingRule the existing MappingRule
	// positions in 3scale change, we always compare the desired keys with the
	// existing MappingRule positions at this point. We do not refetch the list.
	// Although this is sort of temporarily inconsistent the result is consistent
	// because we always update positions in ascending order. The worst case of
	// this is that some updates are performed on MappingRules whose positions
	// are already reconciled, which is unneeded. The alternative would be to
	// refetch all the MappingRules each time we create/update a MappingRule,
	// which would be more inefficient than doing potential unneeded updates as
	// it is being done with this implementation.
	// Additionally, in case the relative position MappingRules' between
	// unmodified MappingRules happens temporarily during the reconciliation that
	// is not an issue due to changes are not effective until the user promotes
	// the configuration
	t.logger.V(1).Info("syncMappingRules", "desiredKeys", desiredKeys)
	for desiredIdxZeroBased, desiredKey := range desiredKeys {
		sMetricID := strings.Split(desiredKey, ":")[2]
		metricID, err := strconv.ParseInt(sMetricID, 10, 64)
		if err != nil {
			return fmt.Errorf("can't convert metricID string : %s to an int64!\n", sMetricID)
		}
		desiredMappingRule := desiredMap[desiredKey]
		// We define the position sent to System starting from one (one-based array)
		// instead of zero-based. The reason for that is that System API does not
		// allow to overwrite an existing MappingRule setting it the position 0.
		// By starting with a minimum value of position 1 we avoid that special
		// implemented behavior by system
		desiredIdx := desiredIdxZeroBased + 1
		if existingMappingRule, ok := existingMap[desiredKey]; ok {
			// Reconcile MappingRule
			t.logger.V(1).Info("syncMappingRules", "desiredMappingRuleToReconcile", desiredKey, "position", desiredIdx)
			err := t.reconcileMappingRuleWithPosition(desiredMappingRule, desiredIdx, existingMappingRule, metricID)
			if err != nil {
				return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
			}
		} else {
			// Create MappingRule
			t.logger.V(1).Info("syncMappingRules", "desiredMappingRuleToCreate", desiredKey, "position", desiredIdx)
			err := t.createNewMappingRuleWithPosition(desiredMappingRule, desiredIdx, metricID)
			if err != nil {
				return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
			}
		}
	}

	return nil
}

func (t *ProductThreescaleReconciler) processNotDesiredMappingRules(notDesiredList []threescaleapi.MappingRuleItem) error {
	for _, mappingRule := range notDesiredList {
		if mappingRule.HTTPMethod == "" || mappingRule.Pattern == "" || mappingRule.Pattern == "/" {
			t.logger.V(1).Info("processNotDesiredMappingRules, delete", "HTTPMethod", mappingRule.HTTPMethod, "Pattern", mappingRule.Pattern, "MetricID", strconv.FormatInt(mappingRule.MetricID, 10))
			err := t.productEntity.DeleteMappingRule(mappingRule.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *ProductThreescaleReconciler) getExistingMappingRules() (map[string]threescaleapi.MappingRuleItem, error) {
	existingMap := map[string]threescaleapi.MappingRuleItem{}
	existingList, err := t.productEntity.MappingRules()
	if err != nil {
		return nil, fmt.Errorf("Error getting product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
	}
	for _, item := range existingList.MappingRules {
		key := fmt.Sprintf("%s:%s:%s", item.Element.HTTPMethod, item.Element.Pattern, strconv.FormatInt(item.Element.MetricID, 10))
		existingMap[key] = item.Element
	}

	return existingMap, nil
}

func (t *ProductThreescaleReconciler) reconcileMappingRuleWithPosition(desired capabilitiesv1beta1.MappingRuleSpec, desiredPosition int, existing threescaleapi.MappingRuleItem, metricID int64) error {
	params := threescaleapi.Params{}

	//
	// Reconcile metric or method
	//

	if metricID != existing.MetricID {
		params["metric_id"] = strconv.FormatInt(metricID, 10)
	}

	//
	// Reconcile delta
	//
	if desired.Increment != existing.Delta {
		params["delta"] = strconv.Itoa(desired.Increment)
	}

	//
	// Reconcile last
	//
	desiredLastAttribute := false
	if desired.Last != nil {
		desiredLastAttribute = *desired.Last
	}

	if desiredLastAttribute != existing.Last {
		params["last"] = strconv.FormatBool(desiredLastAttribute)
	}

	// Reconcile Position
	//
	if desiredPosition != existing.Position {
		params["position"] = strconv.FormatInt(int64(desiredPosition), 10)
	}

	if len(params) > 0 {
		err := t.productEntity.UpdateMappingRule(existing.ID, params)
		if err != nil {
			return fmt.Errorf("Error reconcile product mapping rule: %w", err)
		}
	}

	return nil
}

func (t *ProductThreescaleReconciler) createNewMappingRuleWithPosition(desired capabilitiesv1beta1.MappingRuleSpec, desiredPosition int, metricID int64) error {
	params := threescaleapi.Params{
		"pattern":     desired.Pattern,
		"http_method": desired.HTTPMethod,
		"metric_id":   strconv.FormatInt(metricID, 10),
		"delta":       strconv.Itoa(desired.Increment),
	}

	if desired.Last != nil {
		params["last"] = strconv.FormatBool(*desired.Last)
	}

	params["position"] = strconv.FormatInt(int64(desiredPosition), 10)

	err := t.productEntity.CreateMappingRule(params)
	if err != nil {
		return fmt.Errorf("Error creating product [%s] mappingrule: %w", t.resource.Spec.SystemName, err)
	}

	return nil
}
