package product

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
)

type ThreescaleReconciler struct {
	*reconcilers.BaseReconciler
	resource            *capabilitiesv1beta1.Product
	productEntity       *controllerhelper.ProductEntity
	backendRemoteIndex  *controllerhelper.BackendAPIRemoteIndex
	threescaleAPIClient *threescaleapi.ThreeScaleClient
	logger              logr.Logger
}

func NewThreescaleReconciler(b *reconcilers.BaseReconciler, resource *capabilitiesv1beta1.Product, threescaleAPIClient *threescaleapi.ThreeScaleClient, backendRemoteIndex *controllerhelper.BackendAPIRemoteIndex) *ThreescaleReconciler {
	return &ThreescaleReconciler{
		BaseReconciler:      b,
		resource:            resource,
		threescaleAPIClient: threescaleAPIClient,
		backendRemoteIndex:  backendRemoteIndex,
		logger:              b.Logger().WithValues("3scale Reconciler", resource.Name),
	}
}

func (t *ThreescaleReconciler) Reconcile() (*controllerhelper.ProductEntity, error) {
	productEntity, err := t.reconcile3scaleProduct()
	if err != nil {
		return nil, err
	}
	t.productEntity = productEntity

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

	err = taskRunner.Run()
	if err != nil {
		return nil, err
	}

	return t.productEntity, nil
}

func (t *ThreescaleReconciler) reconcile3scaleProduct() (*controllerhelper.ProductEntity, error) {
	productList, err := t.threescaleAPIClient.ListProducts()
	if err != nil {
		return nil, fmt.Errorf("reconcile3scaleProduct product [%s]: %w", t.resource.Spec.SystemName, err)
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
			return nil, fmt.Errorf("reconcile3scaleProduct product [%s]: %w", t.resource.Spec.SystemName, err)
		}

		productObj = product
	}

	return controllerhelper.NewProductEntity(productObj, t.threescaleAPIClient, t.logger), nil
}
