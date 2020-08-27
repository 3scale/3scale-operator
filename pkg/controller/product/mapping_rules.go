package product

import (
	"errors"
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

type mappingRuleData struct {
	item threescaleapi.MappingRuleItem
	spec capabilitiesv1beta1.MappingRuleSpec
}

func (t *ThreescaleReconciler) syncMappingRules(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.MappingRules))
	desiredMap := map[string]capabilitiesv1beta1.MappingRuleSpec{}
	for _, spec := range t.resource.Spec.MappingRules {
		key := fmt.Sprintf("%s:%s", spec.HTTPMethod, spec.Pattern)
		desiredKeys = append(desiredKeys, key)
		desiredMap[key] = spec
	}

	existingKeys := []string{}
	existingMap := map[string]threescaleapi.MappingRuleItem{}
	existingList, err := t.productEntity.MappingRules()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
	}
	for _, item := range existingList.MappingRules {
		key := fmt.Sprintf("%s:%s", item.Element.HTTPMethod, item.Element.Pattern)
		existingKeys = append(existingKeys, key)
		existingMap[key] = item.Element
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

	//
	// Reconcile existing and changed mapping rules
	//

	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	t.logger.V(1).Info("syncMappingRules", "matchedKeys", matchedKeys)
	matchedList := make([]mappingRuleData, 0, len(matchedKeys))
	for _, key := range matchedKeys {
		matchedList = append(matchedList, mappingRuleData{
			item: existingMap[key],
			spec: desiredMap[key],
		})
	}

	err = t.reconcileMatchedMappingRules(matchedList)
	if err != nil {
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Create not existing and desired mapping rules
	//

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	t.logger.V(1).Info("syncMappingRules", "desiredNewKeys", desiredNewKeys)
	desiredNewList := make([]capabilitiesv1beta1.MappingRuleSpec, 0, len(desiredNewKeys))
	for _, key := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the desiredKeys set
		desiredNewList = append(desiredNewList, desiredMap[key])
	}
	err = t.createNewMappingRules(desiredNewList)
	if err != nil {
		return fmt.Errorf("Error sync product [%s] mappingrules: %w", t.resource.Spec.SystemName, err)
	}

	return nil
}

func (t *ThreescaleReconciler) processNotDesiredMappingRules(notDesiredList []threescaleapi.MappingRuleItem) error {
	for _, mappingRule := range notDesiredList {
		err := t.productEntity.DeleteMappingRule(mappingRule.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ThreescaleReconciler) reconcileMatchedMappingRules(matchedList []mappingRuleData) error {
	for _, data := range matchedList {
		params := threescaleapi.Params{}

		//
		// Reconcile metric or method
		//
		metricID, err := t.productEntity.FindMethodMetricIDBySystemName(data.spec.MetricMethodRef)
		if err != nil {
			return fmt.Errorf("Error reconcile product mapping rule: %w", err)
		}

		if metricID < 0 {
			// Should not happen as metric and method references have been validated and should exists
			return errors.New("product metric method ref for mapping rule not found")
		}

		if metricID != data.item.MetricID {
			params["metric_id"] = strconv.FormatInt(metricID, 10)
		}

		//
		// Reconcile delta
		//
		if data.spec.Increment != data.item.Delta {
			params["delta"] = strconv.Itoa(data.spec.Increment)
		}

		//
		// Reconcile last
		//
		desiredLastAttribute := false
		if data.spec.Last != nil {
			desiredLastAttribute = *data.spec.Last
		}

		if desiredLastAttribute != data.item.Last {
			params["last"] = strconv.FormatBool(desiredLastAttribute)
		}

		if len(params) > 0 {
			err := t.productEntity.UpdateMappingRule(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error reconcile product mapping rule: %w", err)
			}
		}
	}

	return nil
}

func (t *ThreescaleReconciler) createNewMappingRules(desiredList []capabilitiesv1beta1.MappingRuleSpec) error {
	for _, spec := range desiredList {
		metricID, err := t.productEntity.FindMethodMetricIDBySystemName(spec.MetricMethodRef)
		if err != nil {
			return fmt.Errorf("Error creating product [%s] mappingrule: %w", t.resource.Spec.SystemName, err)
		}

		if metricID < 0 {
			// Should not happen as metric and method references have been validated and should exists
			return errors.New("product metric method ref for mapping rule not found")
		}

		params := threescaleapi.Params{
			"pattern":     spec.Pattern,
			"http_method": spec.HTTPMethod,
			"metric_id":   strconv.FormatInt(metricID, 10),
			"delta":       strconv.Itoa(spec.Increment),
		}

		if spec.Last != nil {
			params["last"] = strconv.FormatBool(*spec.Last)
		}

		err = t.productEntity.CreateMappingRule(params)
		if err != nil {
			return fmt.Errorf("Error creating product [%s] mappingrule: %w", t.resource.Spec.SystemName, err)
		}
	}
	return nil
}
