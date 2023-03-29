package controllers

import (
	"fmt"

	capabilitiesv1beta2 "github.com/3scale/3scale-operator/apis/capabilities/v1beta2"
	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

type metricData struct {
	item threescaleapi.MetricItem
	spec capabilitiesv1beta2.MetricSpec
}

func (t *ProductThreescaleReconciler) syncMetrics(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.Metrics))
	for systemName := range t.resource.Spec.Metrics {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingMap := map[string]threescaleapi.MetricItem{}
	existingList, err := t.productEntity.Metrics()
	if err != nil {
		return fmt.Errorf("Error sync product metrics [%s]: %w", t.resource.Spec.SystemName, err)
	}

	existingKeys := make([]string, 0, len(existingList.Metrics))
	for _, existing := range existingList.Metrics {
		systemName := existing.Element.SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	//
	// Deleted existing and not desired metrics
	//

	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	t.logger.V(1).Info("syncMetrics", "notDesiredExistingKeys", notDesiredExistingKeys)
	notDesiredMap := map[string]threescaleapi.MetricItem{}
	for _, systemName := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		notDesiredMap[systemName] = existingMap[systemName]
	}
	err = t.processNotDesiredMetrics(notDesiredMap)
	if err != nil {
		return fmt.Errorf("Error sync product metrics [%s]: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Reconcile existing and changed metrics
	//

	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	t.logger.V(1).Info("syncMetrics", "matchedKeys", matchedKeys)
	matchedMap := map[string]metricData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = metricData{
			item: existingMap[systemName],
			spec: t.resource.Spec.Metrics[systemName],
		}
	}

	err = t.reconcileMatchedMetrics(matchedMap)
	if err != nil {
		return fmt.Errorf("Error sync product metrics [%s]: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Create not existing and desired metrics
	//

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	t.logger.V(1).Info("syncMetrics", "desiredNewKeys", desiredNewKeys)
	desiredNewMap := map[string]capabilitiesv1beta2.MetricSpec{}
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.Metrics map key set
		desiredNewMap[systemName] = t.resource.Spec.Metrics[systemName]
	}
	err = t.createNewMetrics(desiredNewMap)
	if err != nil {
		return fmt.Errorf("Error sync product metrics [%s]: %w", t.resource.Spec.SystemName, err)
	}

	return nil
}

func (t *ProductThreescaleReconciler) processNotDesiredMetrics(notDesiredMap map[string]threescaleapi.MetricItem) error {
	for _, metric := range notDesiredMap {
		err := t.productEntity.DeleteMetric(metric.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ProductThreescaleReconciler) reconcileMatchedMetrics(matchedMap map[string]metricData) error {
	for _, data := range matchedMap {
		params := threescaleapi.Params{}
		if data.spec.Name != data.item.Name {
			params["friendly_name"] = data.spec.Name
		}

		if data.spec.Unit != data.item.Unit {
			params["unit"] = data.spec.Unit
		}

		if data.spec.Description != data.item.Description {
			params["description"] = data.spec.Description
		}

		if len(params) > 0 {
			err := t.productEntity.UpdateMetric(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error updating product metric: %w", err)
			}
		}
	}

	return nil
}

func (t *ProductThreescaleReconciler) createNewMetrics(desiredNewMap map[string]capabilitiesv1beta2.MetricSpec) error {
	for systemName, metric := range desiredNewMap {
		params := threescaleapi.Params{
			"friendly_name": metric.Name,
			"unit":          metric.Unit,
			"system_name":   systemName,
		}
		if len(metric.Description) > 0 {
			params["description"] = metric.Description
		}
		err := t.productEntity.CreateMetric(params)
		if err != nil {
			return err
		}
	}
	return nil
}
