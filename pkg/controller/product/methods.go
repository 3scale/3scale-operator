package product

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

type methodData struct {
	item threescaleapi.MethodItem
	spec capabilitiesv1beta1.Methodpec
}

func (t *ThreescaleReconciler) syncMethods(_ interface{}) error {
	desiredKeys := make([]string, 0, len(t.resource.Spec.Methods))
	for systemName := range t.resource.Spec.Methods {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingMap := map[string]threescaleapi.MethodItem{}
	existingList, err := t.productEntity.Methods()
	if err != nil {
		return fmt.Errorf("Error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	existingKeys := make([]string, 0, len(existingList.Methods))
	for _, existing := range existingList.Methods {
		systemName := existing.Element.SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	desiredNewMap := map[string]capabilitiesv1beta1.Methodpec{}
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewKeys is a subset of the Spec.Method map key set
		desiredNewMap[systemName] = t.resource.Spec.Methods[systemName]
	}
	err = t.createNewMethods(desiredNewMap)
	if err != nil {
		return fmt.Errorf("Error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	notDesiredMap := map[string]threescaleapi.MethodItem{}
	for _, systemName := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		notDesiredMap[systemName] = existingMap[systemName]
	}
	err = t.processNotDesiredMethods(notDesiredMap)
	if err != nil {
		return fmt.Errorf("Error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	matchedMap := map[string]methodData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = methodData{
			item: existingMap[systemName],
			spec: t.resource.Spec.Methods[systemName],
		}
	}

	err = t.reconcileMatchedMethods(matchedMap)
	if err != nil {
		return fmt.Errorf("Error sync product methods [%s]: %w", t.resource.Spec.SystemName, err)
	}

	return nil
}

func (t *ThreescaleReconciler) createNewMethods(desiredNewMap map[string]capabilitiesv1beta1.Methodpec) error {
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

func (t *ThreescaleReconciler) processNotDesiredMethods(notDesiredMap map[string]threescaleapi.MethodItem) error {
	for _, notDesiredMethod := range notDesiredMap {
		err := t.productEntity.DeleteMethod(notDesiredMethod.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ThreescaleReconciler) reconcileMatchedMethods(matchedMap map[string]methodData) error {
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
				return fmt.Errorf("Error reconcile product methods: %w", err)
			}
		}
	}

	return nil
}
