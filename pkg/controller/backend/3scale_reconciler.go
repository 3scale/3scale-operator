package backend

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type methodData struct {
	methodAPIItem threescaleapi.MethodItem
	methodSpec    capabilitiesv1beta1.Methodpec
}

type ThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	backendResource     *capabilitiesv1beta1.Backend
	backendAPIEntity    *helper.BackendAPIEntity
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func NewThreescaleReconciler(b *reconcilers.BaseReconciler, backendResource *capabilitiesv1beta1.Backend, threescaleAPIClient *threescaleapi.ThreeScaleClient) *ThreescaleReconciler {
	return &ThreescaleReconciler{
		BaseReconciler:      b,
		backendResource:     backendResource,
		threescaleAPIClient: threescaleAPIClient,
		logger:              b.Logger().WithValues("3scale Reconciler", backendResource.Name),
	}
}

func (t *ThreescaleReconciler) Reconcile() (*helper.BackendAPIEntity, error) {
	taskRunner := helper.NewTaskRunner(nil, t.logger)
	taskRunner.AddTask("SyncBackend", t.syncBackend)
	// First methods and metrics, then mapping rules.
	// Mapping rules reference methods and metrics.
	// When a method/metric is deleted,
	// any orphan mapping rule will be deleted automatically by 3scale
	taskRunner.AddTask("SyncMethods", t.syncMethods)
	taskRunner.AddTask("SyncMetrics", t.syncMetrics)
	taskRunner.AddTask("SyncMappingRules", t.syncMappingRules)

	err := taskRunner.Run()
	if err != nil {
		return nil, err
	}

	return t.backendAPIEntity, nil
}

func (t *ThreescaleReconciler) syncBackend(_ interface{}) error {
	listObj, err := t.threescaleAPIClient.ListBackendApis()
	if err != nil {
		return fmt.Errorf("Error reading product list: %w", err)
	}

	// Find backend in the list by system name
	idx, exists := func(pList []threescaleapi.BackendApi) (int, bool) {
		for i, item := range pList {
			if item.Element.SystemName == t.backendResource.Spec.SystemName {
				return i, true
			}
		}
		return -1, false
	}(listObj.Backends)

	var backendAPIObj *threescaleapi.BackendApi
	if exists {
		backendAPIObj = &listObj.Backends[idx]
	} else {
		// Create backend using system_name.
		// it cannot be modified later
		params := threescaleapi.Params{
			"system_name":      t.backendResource.Spec.SystemName,
			"name":             t.backendResource.Spec.Name,
			"private_endpoint": t.backendResource.Spec.PrivateBaseURL,
		}
		backend, err := t.threescaleAPIClient.CreateBackendApi(params)
		if err != nil {
			return fmt.Errorf("Error creating product %s: %w", t.backendResource.Spec.SystemName, err)
		}

		backendAPIObj = backend
	}

	// Will be used by coming steps
	t.backendAPIEntity = helper.NewBackendAPIEntity(backendAPIObj, t.threescaleAPIClient, t.logger)

	updatedParams := threescaleapi.Params{}

	if backendAPIObj.Element.Name != t.backendResource.Spec.Name {
		updatedParams["name"] = t.backendResource.Spec.Name
	}

	if backendAPIObj.Element.Description != t.backendResource.Spec.Description {
		updatedParams["description"] = t.backendResource.Spec.Description
	}

	if backendAPIObj.Element.PrivateEndpoint != t.backendResource.Spec.PrivateBaseURL {
		updatedParams["private_endpoint"] = t.backendResource.Spec.PrivateBaseURL
	}

	if len(updatedParams) > 0 {
		err = t.backendAPIEntity.Update(updatedParams)
		if err != nil {
			return fmt.Errorf("Error updating backendAPI: %w", err)
		}
	}

	return nil
}

func (t *ThreescaleReconciler) syncMethods(_ interface{}) error {
	desiredKeys := make([]string, len(t.backendResource.Spec.Methods))
	for systemName := range t.backendResource.Spec.Methods {
		desiredKeys = append(desiredKeys, systemName)
	}

	existingMap := map[string]threescaleapi.MethodItem{}
	existingList, err := t.backendAPIEntity.Methods()
	if err != nil {
		return fmt.Errorf("Error reading backend methods: %w", err)
	}

	existingKeys := make([]string, len(existingList.Methods))
	for _, existing := range existingList.Methods {
		systemName := existing.Element.SystemName
		existingKeys = append(existingKeys, systemName)
		existingMap[systemName] = existing.Element
	}

	desiredNewKeys := helper.ArrayStringDifference(desiredKeys, existingKeys)
	desiredNewMap := map[string]capabilitiesv1beta1.Methodpec{}
	for _, systemName := range desiredNewKeys {
		// key is expected to exist
		// desiredNewMethods is a subset of the Spec.Method map key set
		desiredNewMap[systemName] = t.backendResource.Spec.Methods[systemName]
	}
	err = t.createNewMethods(desiredNewMap)
	if err != nil {
		return fmt.Errorf("Error creating backend methods: %w", err)
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
		return fmt.Errorf("Error processing not desired backend methods: %w", err)
	}

	matchedKeys := helper.ArrayStringIntersection(existingKeys, desiredKeys)
	matchedMap := map[string]methodData{}
	for _, systemName := range matchedKeys {
		matchedMap[systemName] = methodData{
			methodAPIItem: existingMap[systemName],
			methodSpec:    t.backendResource.Spec.Methods[systemName],
		}
	}

	err = t.reconcileMatchedMethods(matchedMap)
	if err != nil {
		return fmt.Errorf("Error reconciling matched backend methods: %w", err)
	}

	return nil
}

func (t *ThreescaleReconciler) createNewMethods(newMethodSystemNames map[string]capabilitiesv1beta1.Methodpec) error {
	for systemName, method := range newMethodSystemNames {
		params := threescaleapi.Params{
			"friendly_name": method.Name,
			"system_name":   systemName,
		}
		if len(method.Description) > 0 {
			params["description"] = method.Description
		}
		err := t.backendAPIEntity.CreateMethod(params)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ThreescaleReconciler) processNotDesiredMethods(methodMap map[string]threescaleapi.MethodItem) error {
	for _, method := range methodMap {
		err := t.backendAPIEntity.DeleteMethod(method.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *ThreescaleReconciler) reconcileMatchedMethods(matchedMap map[string]methodData) error {
	for _, methodData := range matchedMap {
		params := threescaleapi.Params{}
		if methodData.methodSpec.Name != methodData.methodAPIItem.Name {
			params["friendly_name"] = methodData.methodSpec.Name
		}

		if methodData.methodSpec.Description != methodData.methodAPIItem.Description {
			params["description"] = methodData.methodSpec.Description
		}

		if len(params) > 0 {
			err := t.backendAPIEntity.UpdateMethod(methodData.methodAPIItem.ID, params)
			if err != nil {
				return fmt.Errorf("Error updating backendAPI method: %w", err)
			}
		}
	}

	return nil
}

func (t *ThreescaleReconciler) syncMetrics(_ interface{}) error {
	// TODO
	return nil
}

func (t *ThreescaleReconciler) syncMappingRules(_ interface{}) error {
	// TODO
	return nil
}
