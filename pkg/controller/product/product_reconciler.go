package product

import (
	"context"
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

type ProductLogicReconciler struct {
	*reconcilers.BaseReconciler
	product *capabilitiesv1beta1.Product
	logger  logr.Logger

	threescaleAPIClient *threescaleapi.ThreeScaleClient
	productObj          *threescaleapi.Service
	syncError           error
}

func NewProductLogicReconciler(b *reconcilers.BaseReconciler, product *capabilitiesv1beta1.Product) *ProductLogicReconciler {
	return &ProductLogicReconciler{
		BaseReconciler: b,
		product:        product,
		logger:         b.Logger().WithValues("Product Controller", product.Name),
	}
}

func (p *ProductLogicReconciler) Reconcile() (reconcile.Result, error) {
	p.logger.Info("Reconcile Product Logic")
	threescaleAPIClient, err := helper.LookupThreescaleClient(p.Client(), p.product.Namespace, p.product.Spec.ProviderAccountRef)
	if err != nil {
		//Unkown Error - requeue the request.
		p.syncError = err
		return reconcile.Result{}, err
	}
	p.threescaleAPIClient = threescaleAPIClient

	err = p.validate()
	if err != nil {
		p.syncError = err
		// Validation error
		// No need to retry as spec is not valid and needs to be changed
		return reconcile.Result{}, nil
	}

	err = p.checkExternalReferences()
	if err != nil {
		p.syncError = err
		// should retry
		return reconcile.Result{}, err
	}

	productObj, err := p.sync3scale()
	if err != nil {
		p.syncError = err
		// should retry
		return reconcile.Result{}, err
	}
	p.productObj = productObj

	return reconcile.Result{}, nil
}

func (p *ProductLogicReconciler) UpdateStatus() (reconcile.Result, error) {
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
	updateErr := p.Client().Status().Update(context.TODO(), p.product)
	if updateErr != nil {
		return reconcile.Result{}, updateErr
	}

	return reconcile.Result{}, nil
}

func (p *ProductLogicReconciler) validate() error {
	// internal validation
	return nil
}

func (p *ProductLogicReconciler) checkExternalReferences() error {
	return nil
}

func (p *ProductLogicReconciler) sync3scale() (*threescaleapi.Service, error) {
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

func (p *ProductLogicReconciler) calculateStatus() *capabilitiesv1beta1.ProductStatus {
	newStatus := &capabilitiesv1beta1.ProductStatus{}
	if p.productObj != nil {
		newStatus.ID = &p.productObj.ID
		newStatus.State = &p.productObj.State
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

func (p *ProductLogicReconciler) syncCondition() common.Condition {
	condition := common.Condition{
		Type:   capabilitiesv1beta1.ProductSyncedConditionType,
		Status: corev1.ConditionFalse,
	}

	if p.syncError == nil {
		condition.Status = corev1.ConditionTrue
	}

	return condition
}

func (p *ProductLogicReconciler) orphanCondition() common.Condition {
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

func (p *ProductLogicReconciler) invalidCondition() common.Condition {
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

func (p *ProductLogicReconciler) failedCondition() common.Condition {
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

func (p *ProductLogicReconciler) wrappedContext(f func(ctx *Context) error) func(interface{}) error {
	return func(ctxI interface{}) error {
		switch ctx := ctxI.(type) {
		case *Context:
			return f(ctx)
		default:
			return fmt.Errorf("Unknown type %T", ctx)
		}
	}
}

func (p *ProductLogicReconciler) syncProduct(ctx *Context) error {
	serviceList, err := p.threescaleAPIClient.ListServices()
	if err != nil {
		return fmt.Errorf("Error reading product list: %w", err)
	}

	// Find service in the list by system name
	idx, serviceExists := func(sList []threescaleapi.Service) (int, bool) {
		for i, service := range sList {
			if service.SystemName == p.product.Spec.SystemName {
				return i, true
			}
		}
		return -1, false
	}(serviceList.Services)

	var productObj *threescaleapi.Service
	if serviceExists {
		productObj = &serviceList.Services[idx]
	} else {
		// Create service using system_name.
		// 3scale API will use passed attr as system name and it cannot be modified later
		service, err := p.threescaleAPIClient.CreateService(p.product.Spec.SystemName)
		if err != nil {
			return fmt.Errorf("Error creating product %s: %w", p.product.Spec.SystemName, err)
		}

		productObj = &service
	}

	updatedParams := threescaleapi.Params{}
	if productObj.Name != p.product.Spec.Name {
		updatedParams["name"] = p.product.Spec.Name
	}

	if productObj.Description != p.product.Spec.Description {
		updatedParams["description"] = p.product.Spec.Description
	}

	if productObj.DeploymentOption != p.product.DeploymentOption() {
		updatedParams["deployment_option"] = p.product.DeploymentOption()
	}

	if productObj.BackendVersion != p.product.AuthenticationMode() {
		updatedParams["backend_version"] = p.product.AuthenticationMode()
	}

	if len(updatedParams) > 0 {
		updatedService, err := p.threescaleAPIClient.UpdateService(productObj.ID, updatedParams)
		if err != nil {
			return fmt.Errorf("Error updating product %s: %w", productObj.SystemName, err)
		}
		productObj = &updatedService
	}

	ctx.SetProduct(productObj)

	return nil
}

func (p *ProductLogicReconciler) syncBackendUsage(ctx *Context) error {
	return nil
}

func (p *ProductLogicReconciler) syncProxy(ctx *Context) error {
	return nil
}

func (p *ProductLogicReconciler) syncMethods(ctx *Context) error {
	return nil
}

func (p *ProductLogicReconciler) syncMetrics(ctx *Context) error {
	return nil
}

func (p *ProductLogicReconciler) syncMappingRules(ctx *Context) error {
	return nil
}

func (p *ProductLogicReconciler) syncApplicationPlans(ctx *Context) error {
	return nil
}

func (p *ProductLogicReconciler) syncPolicies(ctx *Context) error {
	return nil
}

func (p *ProductLogicReconciler) bumpProxyVersion(ctx *Context) error {
	return nil
}
