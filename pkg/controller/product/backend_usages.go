package product

import (
	"fmt"
	"strconv"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
)

type backendUsageData struct {
	item threescaleapi.BackendAPIUsageItem
	spec capabilitiesv1beta1.BackendUsageSpec
}

type newBackendUsageData struct {
	item threescaleapi.BackendApiItem
	spec capabilitiesv1beta1.BackendUsageSpec
}

func (t *ThreescaleReconciler) syncBackendUsage(_ interface{}) error {
	backendAPIs, err := t.threescaleAPIClient.ListBackendApis()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] backendusages: %w", t.resource.Spec.SystemName, err)
	}

	backendMapByID := map[int64]threescaleapi.BackendApiItem{}
	backendMapBySystemName := map[string]threescaleapi.BackendApiItem{}
	for _, backend := range backendAPIs.Backends {
		backendMapByID[backend.Element.ID] = backend.Element
		backendMapBySystemName[backend.Element.SystemName] = backend.Element
	}

	desiredKeys := make([]string, 0, len(t.resource.Spec.BackendUsages))
	for systemName := range t.resource.Spec.BackendUsages {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingList, err := t.entity.BackendUsages()
	if err != nil {
		return fmt.Errorf("Error sync product [%s] backendusages: %w", t.resource.Spec.SystemName, err)
	}

	existingKeys := make([]string, 0, len(existingList))
	existingMap := map[string]threescaleapi.BackendAPIUsageItem{}
	for _, existing := range existingList {
		// backend usage ID should exist in the backend list
		systemName := backendMapByID[existing.Element.BackendAPIID].SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	//
	// Deleted existing and not desired
	//
	notDesiredExistingKeys := helper.ArrayStringDifference(existingKeys, desiredKeys)
	notDesiredList := make([]threescaleapi.BackendAPIUsageItem, 0, len(notDesiredExistingKeys))
	for _, systemName := range notDesiredExistingKeys {
		// key is expected to exist
		// notDesiredExistingKeys is a subset of the existingMap key set
		notDesiredList = append(notDesiredList, existingMap[systemName])
	}
	err = t.processNotDesiredBackendUsages(notDesiredList)
	if err != nil {
		return fmt.Errorf("Error sync product [%s] backendusages: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Reconcile existing and changed
	//
	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	matchedMap := map[string]backendUsageData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = backendUsageData{
			item: existingMap[systemName],
			spec: t.resource.Spec.BackendUsages[systemName],
		}
	}
	err = t.reconcileMatchedBackendUsages(matchedMap)
	if err != nil {
		return fmt.Errorf("Error sync product [%s] backendusages: %w", t.resource.Spec.SystemName, err)
	}

	//
	// Create not existing and desired
	// Spec validation makes sure all backend resources referenced by backend usage
	// exist and are sync'ed. Thus, they should exist in the backend entity map
	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	desiredNewList := make([]newBackendUsageData, 0, len(desiredNewKeys))
	for _, backendSystemName := range desiredNewKeys {
		// backendSystemName is expected to exist
		// desiredNewKeys is a subset of the Spec.BackendUsages map key set
		desiredNewList = append(desiredNewList, newBackendUsageData{
			spec: t.resource.Spec.BackendUsages[backendSystemName],
			// desiredNewKeys is a set of backend resources systemName's
			// which have been validated of being sync'ed.
			// Thus, the map of existing Backend entities should contain desired systemName
			item: backendMapBySystemName[backendSystemName],
		})
	}
	err = t.createNewBackendUsage(desiredNewList)
	if err != nil {
		return fmt.Errorf("Error sync product [%s] backendusages: %w", t.resource.Spec.SystemName, err)
	}
	return nil
}

func (t *ThreescaleReconciler) processNotDesiredBackendUsages(notDesiredList []threescaleapi.BackendAPIUsageItem) error {
	for _, item := range notDesiredList {
		err := t.entity.DeleteBackendUsage(item.ID)
		if err != nil {
			return fmt.Errorf("Error deleting product backendusage: %w", err)
		}
	}
	return nil
}

func (t *ThreescaleReconciler) reconcileMatchedBackendUsages(matchedMap map[string]backendUsageData) error {
	for _, data := range matchedMap {
		params := threescaleapi.Params{}
		if data.spec.Path != data.item.Path {
			params["path"] = data.spec.Path
		}

		if len(params) > 0 {
			err := t.entity.UpdateBackendUsage(data.item.ID, params)
			if err != nil {
				return fmt.Errorf("Error updating product backendusage: %w", err)
			}
		}
	}

	return nil
}

func (t *ThreescaleReconciler) createNewBackendUsage(matchedList []newBackendUsageData) error {
	for _, data := range matchedList {
		params := threescaleapi.Params{
			"path":           data.spec.Path,
			"backend_api_id": strconv.FormatInt(data.item.ID, 10),
		}
		err := t.entity.CreateBackendUsage(params)
		if err != nil {
			return fmt.Errorf("Error creating product backendusage: %w", err)
		}
	}

	return nil
}
