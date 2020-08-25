package product

import (
	"context"
	"encoding/json"
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	controllerhelper "github.com/3scale/3scale-operator/pkg/controller/helper"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/discovery"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	// controllerName is the name of this controller
	controllerName = "controller_product"

	// package level logger
	log = logf.Log.WithName(controllerName)
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Product Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}

	return add(mgr, reconciler)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	apiClientReader, err := common.NewAPIClientReader(mgr)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}

	client := mgr.GetClient()
	scheme := mgr.GetScheme()
	ctx := context.TODO()
	recorder := mgr.GetEventRecorderFor(controllerName)
	return &ReconcileProduct{
		BaseReconciler: reconcilers.NewBaseReconciler(client, scheme, apiClientReader, ctx, log, discoveryClient, recorder),
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Product
	err = c.Watch(&source.Kind{Type: &capabilitiesv1beta1.Product{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileProduct implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileProduct{}

// ReconcileProduct reconciles a Product object
type ReconcileProduct struct {
	*reconcilers.BaseReconciler
}

// Reconcile reads that state of the cluster for a Product object and makes changes based on the state read
// and what is in the Product.Spec
func (r *ReconcileProduct) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := r.Logger().WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconcile Product", "Operator version", version.Version)

	// Fetch the Product instance
	product := &capabilitiesv1beta1.Product{}
	err := r.Client().Get(r.Context(), request.NamespacedName, product)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("resource not found. Ignoring since object must have been deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(product, "", "  ")
		if err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted Products, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if product.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	result, err := r.reconcile(product)
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile")
		r.EventRecorder().Eventf(product, corev1.EventTypeWarning, "ReconcileError", "%v", err)
	}
	reqLogger.Info("END", "result", result, "error", err)
	return result, err
}

func (r *ReconcileProduct) reconcile(productResource *capabilitiesv1beta1.Product) (reconcile.Result, error) {
	logger := r.Logger().WithValues("product", productResource.Name)

	if productResource.SetDefaults(logger) {
		err := r.Client().Update(r.Context(), productResource)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("Failed setting product defaults: %w", err)
		}

		logger.Info("resource defaults updated. Requeueing.")
		return reconcile.Result{Requeue: true}, nil
	}

	err := r.validateSpec(productResource)
	if err != nil {
		if helper.IsInvalidSpecError(err) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			logger.Info("ERROR", "spec validation error", err)
			r.EventRecorder().Eventf(productResource, corev1.EventTypeWarning, "Invalid Product Spec", "%v", err)
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), productResource.Namespace, productResource.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("reconcile product spec: %w", err)
	}

	err = r.checkExternalRefs(productResource, providerAccount)
	if err != nil {
		if helper.IsOrphanSpecError(err) {
			// On Orphan spec error, retry
			logger.Info("ERROR", "spec orphan error", err)
			return reconcile.Result{Requeue: true}, nil
		}

		return reconcile.Result{}, err
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("reconcile product spec: %w", err)
	}

	backendRemoteIndex, err := controllerhelper.NewBackendAPIRemoteIndex(threescaleAPIClient, r.Logger())
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("reconcile product spec: %w", err)
	}

	reconciler := NewThreescaleReconciler(r.BaseReconciler, productResource, threescaleAPIClient, backendRemoteIndex)
	productEntity, syncErr := reconciler.Reconcile()

	statusReconciler := NewStatusReconciler(r.BaseReconciler, productResource, productEntity, providerAccount.AdminURLStr, syncErr)
	statusResult, statusErr := statusReconciler.Reconcile()
	if statusErr != nil {
		if syncErr != nil {
			return reconcile.Result{}, fmt.Errorf("Failed to sync product: %v. Failed to update product status: %w", syncErr, statusErr)
		}

		return reconcile.Result{}, fmt.Errorf("Failed to update product status: %w", statusErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if syncErr != nil {
		return reconcile.Result{}, fmt.Errorf("Failed to sync product: %w", syncErr)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileProduct) validateSpec(resource *capabilitiesv1beta1.Product) error {
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

func (r *ReconcileProduct) checkExternalRefs(resource *capabilitiesv1beta1.Product, providerAccount *controllerhelper.ProviderAccount) error {
	errors := field.ErrorList{}

	backendList, err := r.backendList(resource, providerAccount)
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

func (r *ReconcileProduct) checkBackendUsages(resource *capabilitiesv1beta1.Product, backendList []capabilitiesv1beta1.Backend) field.ErrorList {
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

// Returns a list of k8s backend list where all elements meet the following conditions:
// - Sync state (ensure remote backend exist and in sync)
// - Same 3scale provider Account as the product
func (r *ReconcileProduct) backendList(resource *capabilitiesv1beta1.Product, productProviderAccount *controllerhelper.ProviderAccount) ([]capabilitiesv1beta1.Backend, error) {
	logger := r.Logger().WithValues("reconcile", resource.Name)
	backendList := &capabilitiesv1beta1.BackendList{}
	opts := []controllerclient.ListOption{
		controllerclient.InNamespace(resource.Namespace),
	}
	err := r.Client().List(r.Context(), backendList, opts...)
	logger.V(1).Info("Get list of Backend resources.", "Err", err)
	if err != nil {
		return nil, fmt.Errorf("backendList: %w", err)
	}
	logger.V(1).Info("Backend resources", "total", len(backendList.Items))

	validBackends := make([]capabilitiesv1beta1.Backend, 0)
	for idx := range backendList.Items {
		// Only synchronized
		if !backendList.Items[idx].IsSynced() {
			continue
		}

		backendProviderAccount, err := controllerhelper.LookupProviderAccount(r.Client(), resource.Namespace, backendList.Items[idx].Spec.ProviderAccountRef, r.Logger())
		if err != nil {
			return nil, fmt.Errorf("backendList: %w", err)
		}

		// Only same provider Account
		if productProviderAccount.AdminURLStr != backendProviderAccount.AdminURLStr {
			continue
		}
		validBackends = append(validBackends, backendList.Items[idx])
	}

	logger.V(1).Info("Backend valid resources", "total", len(validBackends))
	return validBackends, nil
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
