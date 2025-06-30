package controllers

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

type methodData struct {
	item threescaleapi.MethodItem
	spec capabilitiesv1beta1.MethodSpec
}

func (t *ProductThreescaleReconciler) syncMethods(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.Methods))
	for systemName := range t.resource.Spec.Methods {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingMap := map[string]threescaleapi.MethodItem{}
	existingList, err := t.productEntity.Methods()
	if err != nil {
		return fmt.Errorf("error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	existingKeys := make([]string, 0, len(existingList.Methods))
	for _, existing := range existingList.Methods {
		systemName := existing.Element.SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	//
	// Deleted existing and not desired
	//
	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	t.logger.V(1).Info("syncMethods", "notDesiredExistingKeys", notDesiredExistingKeys)
	notDesiredMap := map[string]threescaleapi.MethodItem{}
	for _, systemName := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		notDesiredMap[systemName] = existingMap[systemName]
	}
	err = t.processNotDesiredMethods(notDesiredMap)
	if err != nil {
		return fmt.Errorf("error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Reconcile existing and changed
	//
	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	t.logger.V(1).Info("syncMethods", "matchedKeys", matchedKeys)
	matchedMap := map[string]methodData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = methodData{
			item: existingMap[systemName],
			spec: t.resource.Spec.Methods[systemName],
		}
	}

	err = t.reconcileMatchedMethods(matchedMap)
	if err != nil {
		return fmt.Errorf("error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Create not existing and desired
	//
	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	t.logger.V(1).Info("syncMethods", "desiredNewKeys", desiredNewKeys)
	desiredNewMap := map[string]capabilitiesv1beta1.MethodSpec{}
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.Method map key set
		desiredNewMap[systemName] = t.resource.Spec.Methods[systemName]
	}
	err = t.createNewMethods(desiredNewMap)
	if err != nil {
		return fmt.Errorf("error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	return nil
}

func (t *ProductThreescaleReconciler) createNewMethods(desiredNewMap map[string]capabilitiesv1beta1.MethodSpec) error {
	for systemName, method := range desiredNewMap {
		params := threescaleapi.Params{
			"friendly_name": method.Name,
			"system_name":   systemName,
		}
		if len(method.Description) > 0 {
			params["description"] = method.Description
		}
		err := t.productEntity.CreateMethod(params)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ProductThreescaleReconciler) processNotDesiredMethods(notDesiredMap map[string]threescaleapi.MethodItem) error {
	for _, notDesiredMethod := range notDesiredMap {
		err := t.productEntity.DeleteMethod(notDesiredMethod.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ProductThreescaleReconciler) reconcileMatchedMethods(matchedMap map[string]methodData) error {
	for _, data := range matchedMap {
		params := threescaleapi.Params{}
		if data.spec.Name != data.item.Name {
			params["friendly_name"] = data.spec.Name
		}

		if data.spec.Description != data.item.Description {
			params["description"] = data.spec.Description
		}

		if len(params) > 0 {
			err := t.productEntity.UpdateMethod(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("error reconcile product methods: %w", err)
			}
		}
	}

	return nil
}
