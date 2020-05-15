package product

import (
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	threescaleapi "github.com/3scale/3scale-porta-go-client/client"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type LogicReconciler struct {
	*reconcilers.BaseReconciler
	product *capabilitiesv1beta1.Product
	logger  logr.Logger

	threescaleAPIClient *threescaleapi.ThreeScaleClient
	productObj          *threescaleapi.Product
	syncError           error
}

func NewLogicReconciler(b *reconcilers.BaseReconciler, product *capabilitiesv1beta1.Product) *LogicReconciler {
	return &LogicReconciler{
		BaseReconciler: b,
		product:        product,
		logger:         b.Logger().WithValues("Product Controller", product.Name),
	}
}

func (p *LogicReconciler) Reconcile() (reconcile.Result, error) {
	threescaleAPIClient, err := helper.LookupThreescaleClient(p.Client(), p.product.Namespace, p.product.Spec.ProviderAccountRef, p.logger)
	if err != nil {
		//Unkown Error - requeue the request.
		p.syncError = err
		return reconcile.Result{}, err
	}
	p.threescaleAPIClient = threescaleAPIClient

	err = p.validate()
	if err != nil {
		p.syncError = err
		p.logger.Error(err, "validate error")

		if !helper.IsInvalidSpecError(err) {
			return reconcile.Result{}, err
		}

		// On Validation error, no need to retry as spec is not valid and needs to be changed
		return reconcile.Result{}, nil
	}

	err = p.checkExternalReferences()
	if err != nil {
		p.logger.Error(err, "External references")
		p.syncError = err

		// should retry anyway
		return reconcile.Result{}, err
	}

	productObj, err := p.sync3scale()
	if err != nil {
		p.syncError = err
		p.logger.Error(err, "3scale sync error")

		// should retry anyway
		return reconcile.Result{}, err
	}

	p.logger.V(1).Info("3scale synchronized correctly")
	p.productObj = productObj

	return reconcile.Result{}, nil
}

func (p *LogicReconciler) UpdateStatus() (reconcile.Result, error) {
	p.logger.Info("Update Status")
	newStatus := p.calculateStatus()

	equalStatus := p.product.Status.Equals(newStatus)
	p.logger.V(1).Info("Status", "status is different", !equalStatus)
	p.logger.V(1).Info("Status", "generation is different", p.product.Generation != p.product.Status.ObservedGeneration)
	if equalStatus && p.product.Generation == p.product.Status.ObservedGeneration {
		// Steady state
		p.logger.V(1).Info("Status was not updated")
		return reconcile.Result{}, nil
	}

	// Save the generation number we acted on, otherwise we might wrongfully indicate
	// that we've seen a spec update when we retry.
	// TODO: This can clobber an update if we allow multiple agents to write to the
	// same status.
	newStatus.ObservedGeneration = p.product.Generation

	p.logger.V(1).Info("Status", "sequence no:", fmt.Sprintf("sequence No: %v->%v", p.product.Status.ObservedGeneration, newStatus.ObservedGeneration))

	p.product.Status = *newStatus
	updateErr := p.Client().Status().Update(p.Context(), p.product)
	if updateErr != nil {
		return reconcile.Result{}, updateErr
	}

	return reconcile.Result{}, nil
}

func (p *LogicReconciler) validate() error {
	// internal validation

	// spec.deployment is oneOf by CRD openapiV3 validation
	// spec.deployment.apicastSelfManaged.authentication is oneOf by CRD openapiV3 validation
	return nil
}

func (p *LogicReconciler) checkExternalReferences() error {
	// Check backend usages
	// build map[backendResourceName] -> backendID
	return nil
}

func (p *LogicReconciler) sync3scale() (*threescaleapi.Product, error) {
	ctx := &Context{}
	taskRunner := helper.NewTaskRunner(ctx, p.logger)

	taskRunner.AddTask("SyncProduct", p.wrappedContext(p.syncProduct))
	taskRunner.AddTask("SyncBackendUsage", p.wrappedContext(p.syncBackendUsage))
	taskRunner.AddTask("SyncProxy", p.wrappedContext(p.syncProxy))
	taskRunner.AddTask("SyncMethods", p.wrappedContext(p.syncMethods))
	taskRunner.AddTask("SyncMetrics", p.wrappedContext(p.syncMetrics))
	taskRunner.AddTask("SyncMappingRules", p.wrappedContext(p.syncMappingRules))
	taskRunner.AddTask("SyncApplicationPlans", p.wrappedContext(p.syncApplicationPlans))
	taskRunner.AddTask("SyncPolicies", p.wrappedContext(p.syncPolicies))
	// This should be the last step
	taskRunner.AddTask("BumbProxyVersion", p.wrappedContext(p.bumpProxyVersion))

	err := taskRunner.Run()
	if err != nil {
		return nil, err
	}

	return ctx.Product(), nil
}

func (p *LogicReconciler) calculateStatus() *capabilitiesv1beta1.ProductStatus {
	newStatus := &capabilitiesv1beta1.ProductStatus{}
	if p.productObj != nil {
		newStatus.ID = &p.productObj.Element.ID
		newStatus.State = &p.productObj.Element.State
	}

	newStatus.ObservedGeneration = p.product.Status.ObservedGeneration

	newStatus.Conditions = p.product.Status.Conditions
	newStatus.Conditions.SetCondition(p.syncCondition())
	newStatus.Conditions.SetCondition(p.orphanCondition())
	newStatus.Conditions.SetCondition(p.invalidCondition())
	newStatus.Conditions.SetCondition(p.failedCondition())

	// terminal problems
	if helper.IsInvalidSpecError(p.syncError) {
		message := p.syncError.Error()
		newStatus.ErrorMessage = &message
	}

	return newStatus
}

func (p *LogicReconciler) syncCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductSyncedConditionType,
		Status: corev1.ConditionFalse,
	}

	if p.syncError == nil {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (p *LogicReconciler) orphanCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductOrphanConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsOrphanSpecError(p.syncError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = p.syncError.Error()
	}

	return condition
}

func (p *LogicReconciler) invalidCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductInvalidConditionType,
		Status: corev1.ConditionFalse,
	}

	if helper.IsInvalidSpecError(p.syncError) {
		condition.Status = corev1.ConditionTrue
		condition.Message = p.syncError.Error()
	}

	return condition
}

func (p *LogicReconciler) failedCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductFailedConditionType,
		Status: corev1.ConditionFalse,
	}

	// This condition could be activated together with other conditions
	if p.syncError != nil {
		condition.Status = corev1.ConditionTrue
		condition.Message = p.syncError.Error()
	}

	return condition
}

func (p *LogicReconciler) wrappedContext(f func(ctx *Context) error) func(interface{}) error {
	return func(ctxI interface{}) error {
		switch ctx := ctxI.(type) {
		case *Context:
			return f(ctx)
		default:
			return fmt.Errorf("Unknown type %T", ctx)
		}
	}
}

func (p *LogicReconciler) syncProduct(ctx *Context) error {
	productList, err := p.threescaleAPIClient.ListProducts()
	if err != nil {
		return fmt.Errorf("Error reading product list: %w", err)
	}

	// Find product in the list by system name
	idx, productExists := func(pList []threescaleapi.Product) (int, bool) {
		for i, product := range pList {
			if product.Element.SystemName == p.product.Spec.SystemName {
				return i, true
			}
		}
		return -1, false
	}(productList.Products)

	var productObj *threescaleapi.Product
	if productExists {
		productObj = &productList.Products[idx]
	} else {
		// Create product using system_name.
		// it cannot be modified later
		params := threescaleapi.Params{
			"system_name": p.product.Spec.SystemName,
		}
		product, err := p.threescaleAPIClient.CreateProduct(p.product.Spec.Name, params)
		if err != nil {
			return fmt.Errorf("Error creating product %s: %w", p.product.Spec.SystemName, err)
		}

		productObj = product
	}

	updatedParams := threescaleapi.Params{}
	if productObj.Element.Name != p.product.Spec.Name {
		updatedParams["name"] = p.product.Spec.Name
	}

	if productObj.Element.Description != p.product.Spec.Description {
		updatedParams["description"] = p.product.Spec.Description
	}

	specDeploymentOption := p.product.Spec.DeploymentOption()
	if specDeploymentOption != nil {
		if productObj.Element.DeploymentOption != *specDeploymentOption {
			updatedParams["deployment_option"] = *specDeploymentOption
		}
	} // only update deployment_option when set in the CR

	specAuthMode := p.product.Spec.AuthenticationMode()
	if specAuthMode != nil {
		if productObj.Element.BackendVersion != *specAuthMode {
			updatedParams["backend_version"] = *specAuthMode
		}
	} // only update backend_version when set in the CR

	if len(updatedParams) > 0 {
		updatedProduct, err := p.threescaleAPIClient.UpdateProduct(productObj.Element.ID, updatedParams)
		if err != nil {
			return fmt.Errorf("Error updating product %s: %w", productObj.Element.SystemName, err)
		}
		productObj = updatedProduct
	}

	ctx.SetProduct(productObj)

	return nil
}

func (p *LogicReconciler) syncBackendUsage(ctx *Context) error {
	return nil
}

func (p *LogicReconciler) syncProxy(ctx *Context) error {
	return nil
}

func (p *LogicReconciler) syncMethods(ctx *Context) error {
	return nil
}

func (p *LogicReconciler) syncMetrics(ctx *Context) error {
	return nil
}

func (p *LogicReconciler) syncMappingRules(ctx *Context) error {
	return nil
}

func (p *LogicReconciler) syncApplicationPlans(ctx *Context) error {
	return nil
}

func (p *LogicReconciler) syncPolicies(ctx *Context) error {
	return nil
}

func (p *LogicReconciler) bumpProxyVersion(ctx *Context) error {
	return nil
}
