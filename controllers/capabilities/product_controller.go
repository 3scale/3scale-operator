/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
)

// ProductReconciler reconciles a Product object
type ProductReconciler struct {
	*reconcilers.BaseReconciler
}

// blank assignment to verify that ProductReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &ProductReconciler{}

// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=products,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=products/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=capabilities.3scale.net,namespace=placeholder,resources=products/finalizers,verbs=get;list;watch;create;update;patch;delete

func (r *ProductReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	reqLogger := r.Logger().WithValues("product", req.NamespacedName)
	reqLogger.Info("Reconcile Product", "Operator version", version.Version)

	// Fetch the Product instance
	product := &capabilitiesv1beta1.Product{}
	err := r.Client().Get(r.Context(), req.NamespacedName, product)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(product, "", "  ")
		if err != nil {
			return ctrl.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted Products, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if product.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if product.SetDefaults(reqLogger) {
		err := r.Client().Update(r.Context(), product)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("Failed setting product defaults: %w", err)
		}

		reqLogger.Info("resource defaults updated. Requeueing.")
		return ctrl.Result{Requeue: true}, nil
	}

	statusReconciler, reconcileErr := r.reconcile(product)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return ctrl.Result{}, fmt.Errorf("Failed to sync product: %v. Failed to update product status: %w", reconcileErr, statusUpdateErr)
		}

		return ctrl.Result{}, fmt.Errorf("Failed to update product status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(product, corev1.EventTypeWarning, "Invalid Product Spec", "%v", reconcileErr)
			return ctrl.Result{}, nil
		}

		if helper.IsOrphanSpecError(reconcileErr) {
			// On Orphan spec error, retry
			reqLogger.Info("ERROR", "spec orphan error", reconcileErr)
			return ctrl.Result{Requeue: true}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(product, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
	}

	reqLogger.Info("END", "error", reconcileErr)
	return ctrl.Result{}, reconcileErr
}

func (r *ProductReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capabilitiesv1beta1.Product{}).
		Complete(r)
}

func (r *ProductReconciler) reconcile(productResource *capabilitiesv1beta1.Product) (*ProductStatusReconciler, error) {
	logger := r.Logger().WithValues("product", productResource.Name)

	err := r.validateSpec(productResource)
	if err != nil {
		statusReconciler := NewProductStatusReconciler(r.BaseReconciler, productResource, nil, "", err)
		return statusReconciler, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), productResource.Namespace, productResource.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewProductStatusReconciler(r.BaseReconciler, productResource, nil, "", err)
		return statusReconciler, err
	}

	err = r.checkExternalRefs(productResource, providerAccount)
	logger.Info("checkExternalRefs", "err", err)
	if err != nil {
		statusReconciler := NewProductStatusReconciler(r.BaseReconciler, productResource, nil, providerAccount.AdminURLStr, err)
		return statusReconciler, err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		statusReconciler := NewProductStatusReconciler(r.BaseReconciler, productResource, nil, providerAccount.AdminURLStr, err)
		return statusReconciler, err
	}

	backendRemoteIndex, err := controllerhelper.NewBackendAPIRemoteIndex(threescaleAPIClient, logger)
	if err != nil {
		statusReconciler := NewProductStatusReconciler(r.BaseReconciler, productResource, nil, providerAccount.AdminURLStr, err)
		return statusReconciler, err
	}

	reconciler := NewProductThreescaleReconciler(r.BaseReconciler, productResource, threescaleAPIClient, backendRemoteIndex)
	productEntity, err := reconciler.Reconcile()
	statusReconciler := NewProductStatusReconciler(r.BaseReconciler, productResource, productEntity, providerAccount.AdminURLStr, err)
	return statusReconciler, err
}

func (r *ProductReconciler) validateSpec(resource *capabilitiesv1beta1.Product) error {
	errors := field.ErrorList{}
	errors = append(errors, resource.Validate()...)

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.InvalidError,
		FieldErrorList: errors,
	}
}

func (r *ProductReconciler) checkExternalRefs(resource *capabilitiesv1beta1.Product, providerAccount *controllerhelper.ProviderAccount) error {
	logger := r.Logger().WithValues("product", resource.Name)
	errors := field.ErrorList{}

	backendList, err := controllerhelper.BackendList(resource.Namespace, r.Client(), providerAccount.AdminURLStr, logger)
	if err != nil {
		return fmt.Errorf("checking backend usage references: %w", err)
	}

	backendUsageErrors := r.checkBackendUsages(resource, backendList)
	errors = append(errors, backendUsageErrors...)

	backendUsageList := computeBackendUsageList(backendList, resource.Spec.BackendUsages)

	limitBackendMetricRefErrors := checkAppLimitsExternalRefs(resource, backendUsageList)
	errors = append(errors, limitBackendMetricRefErrors...)

	pricingRulesBackendMetricRefErrors := checkAppPricingRulesExternalRefs(resource, backendUsageList)
	errors = append(errors, pricingRulesBackendMetricRefErrors...)

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.OrphanError,
		FieldErrorList: errors,
	}
}

func (r *ProductReconciler) checkBackendUsages(resource *capabilitiesv1beta1.Product, backendList []capabilitiesv1beta1.Backend) field.ErrorList {
	errors := field.ErrorList{}

	specFldPath := field.NewPath("spec")
	backendUsageFldPath := specFldPath.Child("backendUsages")
	for systemName := range resource.Spec.BackendUsages {
		idx := findBackendBySystemName(backendList, systemName)
		if idx < 0 {
			keyFldPath := backendUsageFldPath.Key(systemName)
			errors = append(errors, field.Invalid(keyFldPath, resource.Spec.BackendUsages[systemName], "backend usage does not have valid backend reference."))
		}
	}

	return errors
}

func checkAppLimitsExternalRefs(resource *capabilitiesv1beta1.Product, backendList []capabilitiesv1beta1.Backend) field.ErrorList {
	// backendList param is expected to be valid product's backendUsageList
	errors := field.ErrorList{}

	specFldPath := field.NewPath("spec")
	applicationPlansFldPath := specFldPath.Child("applicationPlans")
	for planSystemName, planSpec := range resource.Spec.ApplicationPlans {
		planFldPath := applicationPlansFldPath.Key(planSystemName)
		limitsFldPath := planFldPath.Child("limits")
		for idx, limitSpec := range planSpec.Limits {
			if limitSpec.MetricMethodRef.BackendSystemName == nil {
				continue
			}

			limitFldPath := limitsFldPath.Index(idx)
			metricRefFldPath := limitFldPath.Child("metricMethodRef")
			backendIdx := findBackendBySystemName(backendList, *limitSpec.MetricMethodRef.BackendSystemName)
			// Check backend reference is one of the backend usage list
			if backendIdx < 0 {
				backendRefFldPath := metricRefFldPath.Child("backend")
				errors = append(errors, field.Invalid(backendRefFldPath, limitSpec.MetricMethodRef.BackendSystemName, "plan limit has invalid backend reference."))
				continue
			}

			// check backend metric reference
			backendResource := backendList[backendIdx]
			if !backendResource.FindMetricOrMethod(limitSpec.MetricMethodRef.SystemName) {
				metricRefSystemNameFldPath := metricRefFldPath.Child("systemName")
				errors = append(errors, field.Invalid(metricRefSystemNameFldPath, limitSpec.MetricMethodRef.SystemName, "plan limit has invalid backend metric or method reference."))
			}
		}
	}

	return errors
}

func checkAppPricingRulesExternalRefs(resource *capabilitiesv1beta1.Product, backendList []capabilitiesv1beta1.Backend) field.ErrorList {
	// backendList param is expected to be valid product's backendUsageList
	errors := field.ErrorList{}

	specFldPath := field.NewPath("spec")
	applicationPlansFldPath := specFldPath.Child("applicationPlans")
	for planSystemName, planSpec := range resource.Spec.ApplicationPlans {
		planFldPath := applicationPlansFldPath.Key(planSystemName)
		rulesFldPath := planFldPath.Child("pricingRules")
		for idx, ruleSpec := range planSpec.PricingRules {
			if ruleSpec.MetricMethodRef.BackendSystemName == nil {
				continue
			}

			ruleFldPath := rulesFldPath.Index(idx)
			metricRefFldPath := ruleFldPath.Child("metricMethodRef")
			backendIdx := findBackendBySystemName(backendList, *ruleSpec.MetricMethodRef.BackendSystemName)
			// Check backend reference is one of the backend usage list
			if backendIdx < 0 {
				backendRefFldPath := metricRefFldPath.Child("backend")
				errors = append(errors, field.Invalid(backendRefFldPath, ruleSpec.MetricMethodRef.BackendSystemName, "plan pricing rule has invalid backend reference."))
				continue
			}

			// check backend metric reference
			backendResource := backendList[backendIdx]
			if !backendResource.FindMetricOrMethod(ruleSpec.MetricMethodRef.SystemName) {
				metricRefSystemNameFldPath := metricRefFldPath.Child("systemName")
				errors = append(errors, field.Invalid(metricRefSystemNameFldPath, ruleSpec.MetricMethodRef.SystemName, "plan pricing rule has invalid backend metric or method reference."))
			}
		}
	}

	return errors
}

func findBackendBySystemName(list []capabilitiesv1beta1.Backend, systemName string) int {
	for idx := range list {
		if list[idx].Spec.SystemName == systemName {
			return idx
		}
	}
	return -1
}

func computeBackendUsageList(list []capabilitiesv1beta1.Backend, backendUsageMap map[string]capabilitiesv1beta1.BackendUsageSpec) []capabilitiesv1beta1.Backend {
	target := map[string]bool{}
	for systemName := range backendUsageMap {
		target[systemName] = true
	}

	result := make([]capabilitiesv1beta1.Backend, 0)
	for _, backend := range list {
		if _, ok := target[backend.Spec.SystemName]; ok {
			result = append(result, backend)
		}
	}

	return result
}
