package product

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
	resource            *capabilitiesv1beta1.Product
	entity              *helper.ProductEntity
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func NewThreescaleReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.Product, threescaleAPIClient *threescaleapi.ThreeScaleClient) *ThreescaleReconciler {
	return &ThreescaleReconciler{
		BaseReconciler:      b,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		logger:              b.Logger().WithValues("3scale Reconciler", resource.Name),
	}
}

func (t *ThreescaleReconciler) Reconcile() (*helper.ProductEntity, error) {
	taskRunner := helper.NewTaskRunner(nil, t.logger)
	taskRunner.AddTask("SyncProduct", t.syncProduct)
	taskRunner.AddTask("SyncBackendUsage", t.syncBackendUsage)
	taskRunner.AddTask("SyncProxy", t.syncProxy)
	// First methods and metrics, then mapping rules.
	// Mapping rules reference methods and metrics.
	// When a method/metric is deleted,
	// any orphan mapping rule will be deleted automatically by 3scale
	taskRunner.AddTask("SyncMethods", t.syncMethods)
	taskRunner.AddTask("SyncMetrics", t.syncMetrics)
	taskRunner.AddTask("SyncMappingRules", t.syncMappingRules)
	taskRunner.AddTask("SyncApplicationPlans", t.syncApplicationPlans)
	taskRunner.AddTask("SyncPolicies", t.syncPolicies)
	// This should be the last step
	taskRunner.AddTask("BumbProxyVersion", t.bumpProxyVersion)

	err := taskRunner.Run()
	if err != nil {
		return nil, err
	}

	return t.entity, nil
}

func (t *ThreescaleReconciler) syncProduct(_ interface{}) error {
	productList, err := t.threescaleAPIClient.ListProducts()
	if err != nil {
		return fmt.Errorf("Error sync product [%s]: %w", t.resource.Spec.SystemName, err)
	}

	// Find product in the list by system name
	idx, exists := func(pList []threescaleapi.Product) (int, bool) {
		for i, item := range pList {
			if item.Element.SystemName == t.resource.Spec.SystemName {
				return i, true
			}
		}
		return -1, false
	}(productList.Products)

	var productObj *threescaleapi.Product
	if exists {
		productObj = &productList.Products[idx]
	} else {
		// Create product using system_name.
		// it cannot be modified later
		params := threescaleapi.Params{
			"system_name": t.resource.Spec.SystemName,
		}
		product, err := t.threescaleAPIClient.CreateProduct(t.resource.Spec.Name, params)
		if err != nil {
			return fmt.Errorf("Error creating product %s: %w", t.resource.Spec.SystemName, err)
		}

		productObj = product
	}

	// Will be used by coming steps
	t.entity = helper.NewProductEntity(productObj, t.threescaleAPIClient, t.logger)

	params := threescaleapi.Params{}

	if productObj.Element.Name != t.resource.Spec.Name {
		params["name"] = t.resource.Spec.Name
	}

	if productObj.Element.Description != t.resource.Spec.Description {
		params["description"] = t.resource.Spec.Description
	}

	specDeploymentOption := t.resource.Spec.DeploymentOption()
	if specDeploymentOption != nil {
		if productObj.Element.DeploymentOption != *specDeploymentOption {
			params["deployment_option"] = *specDeploymentOption
		}
	} // only update deployment_option when set in the CR

	specAuthMode := t.resource.Spec.AuthenticationMode()
	if specAuthMode != nil {
		if productObj.Element.BackendVersion != *specAuthMode {
			params["backend_version"] = *specAuthMode
		}
	} // only update backend_version when set in the CR

	if len(params) > 0 {
		err = t.entity.Update(params)
		if err != nil {
			return fmt.Errorf("Error sync product [%s;%d]: %w", t.resource.Spec.SystemName, productObj.Element.ID, err)
		}
	}

	return nil
}

func (t *ThreescaleReconciler) syncProxy(_ interface{}) error {
	return nil
}

func (t *ThreescaleReconciler) syncApplicationPlans(_ interface{}) error {
	return nil
}

func (t *ThreescaleReconciler) syncPolicies(_ interface{}) error {
	return nil
}

func (t *ThreescaleReconciler) bumpProxyVersion(_ interface{}) error {
	return nil
}
