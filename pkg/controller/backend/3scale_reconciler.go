package backend

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

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
	t.backendAPIEntity = helper.NewBackendAPIEntity(backendAPIObj, t.threescaleAPIClient)

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
	desiredMethodsKeys := make([]string, len(t.backendResource.Spec.Methods))
	for methodSystemName := range t.backendResource.Spec.Methods {
		desiredMethodsKeys = append(desiredMethodsKeys, methodSystemName)
	}

	existingMethods := map[string]threescaleapi.MethodItem{}
	existingMethodList, err := t.backendAPIEntity.Methods()
	if err != nil {
		return fmt.Errorf("Error reading backend methods: %w", err)
	}

	for _, existingMethod := range existingMethodList.Methods {
		systemName := existingMethod.Element.SystemName
		existingMethods[systemName] = existingMethod.Element
	}

	existingMethodsKeys := make([]string, len(existingMethods))
	for methodSystemName := range existingMethods {
		existingMethodsKeys = append(existingMethodsKeys, methodSystemName)
	}

	desiredNewMethods := helper.ArrayStringDifference(desiredMethodsKeys, existingMethodsKeys)
	err = t.createNewMethods(desiredNewMethods)
	if err != nil {
		return fmt.Errorf("Error creating backend methods: %w", err)
	}

	notDesiresExistingMethods := helper.ArrayStringDifference(existingMethodsKeys, desiredMethodsKeys)
	err = t.processNotDesiredMethods(notDesiresExistingMethods)
	if err != nil {
		return fmt.Errorf("Error processing not desired backend methods: %w", err)
	}

	matchedMethods := helper.ArrayStringIntersection(existingMethodsKeys, desiredMethodsKeys)
	err = t.reconcileMatchedMethods(matchedMethods)
	if err != nil {
		return fmt.Errorf("Error reconciling matched backend methods: %w", err)
	}

	return nil
}

func (t *ThreescaleReconciler) createNewMethods(_newMethodNames []string) error {
	// TODO
	return nil
}

func (t *ThreescaleReconciler) processNotDesiredMethods(_newMethodNames []string) error {
	// TODO
	return nil
}

func (t *ThreescaleReconciler) reconcileMatchedMethods(_newMethodNames []string) error {
	// TODO
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
