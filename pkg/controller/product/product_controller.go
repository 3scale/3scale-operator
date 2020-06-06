package product

import (
	"context"
	"encoding/json"
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
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

	client := mgr.GetClient()
	scheme := mgr.GetScheme()
	ctx := context.TODO()
	recorder := mgr.GetEventRecorderFor(controllerName)
	return &ReconcileProduct{
		BaseReconciler: reconcilers.NewBaseReconciler(client, scheme, apiClientReader, ctx, log, recorder),
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
	logger := r.Logger().WithValues("reconcile", productResource.Name)

	if productResource.SetDefaults() {
		err := r.Client().Update(r.Context(), productResource)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("Failed setting product defaults: %w", err)
		}

		return reconcile.Result{Requeue: true}, nil
	}

	productEntity, specErr := r.reconcileSpec(productResource)

	statusReconciler := NewStatusReconciler(r.BaseReconciler, productResource, productEntity, specErr)
	statusErr := statusReconciler.Reconcile()
	if statusErr != nil {
		if specErr != nil {
			return reconcile.Result{}, fmt.Errorf("Failed to sync product: %v. Failed to update product status: %w", specErr, statusErr)
		}

		return reconcile.Result{}, fmt.Errorf("Failed to update product status: %w", statusErr)
	}

	if helper.IsInvalidSpecError(specErr) {
		// On Validation error, no need to retry as spec is not valid and needs to be changed
		logger.Info("ERROR", "spec validation error", specErr)
		return reconcile.Result{}, nil
	}

	if specErr != nil {
		return reconcile.Result{}, fmt.Errorf("Failed to sync product: %w", specErr)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileProduct) reconcileSpec(resource *capabilitiesv1beta1.Product) (*helper.ProductEntity, error) {
	err := r.validateSpec(resource)
	if err != nil {
		return nil, fmt.Errorf("reconcile product spec: %w", err)
	}

	providerAccount, err := helper.LookupProviderAccount(r.Client(), resource.Namespace, resource.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return nil, fmt.Errorf("reconcile product spec: %w", err)
	}

	err = r.checkExternalRefs(resource, providerAccount)
	if err != nil {
		return nil, fmt.Errorf("reconcile product spec: %w", err)
	}

	threescaleAPIClient, err := helper.PortaClient(providerAccount)
	if err != nil {
		return nil, fmt.Errorf("reconcile product spec: %w", err)
	}

	reconciler := NewThreescaleReconciler(r.BaseReconciler, resource, threescaleAPIClient)
	entity, err := reconciler.Reconcile()
	if err != nil {
		return nil, fmt.Errorf("reconcile product spec: %w", err)
	}

	return entity, nil
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

func (r *ReconcileProduct) checkExternalRefs(resource *capabilitiesv1beta1.Product, providerAccount *helper.ProviderAccount) error {
	errors := field.ErrorList{}
	// external validation
	backendUsageErrors, err := r.checkBackendUsages(resource, providerAccount)
	if err != nil {
		return fmt.Errorf("check product external refs: %w", err)
	}

	errors = append(errors, backendUsageErrors...)

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.OrphanError,
		FieldErrorList: errors,
	}
}

func (r *ReconcileProduct) checkBackendUsages(resource *capabilitiesv1beta1.Product, providerAccount *helper.ProviderAccount) (field.ErrorList, error) {
	errors := field.ErrorList{}
	backendList, err := r.backendList(resource, providerAccount)
	if err != nil {
		return nil, fmt.Errorf("checking backend usage references: %w", err)
	}

	specFldPath := field.NewPath("spec")
	backendUsageFldPath := specFldPath.Child("backendUsages")
	for systemName := range resource.Spec.BackendUsages {
		if !findBackendBySystemName(backendList, systemName) {
			keyFldPath := backendUsageFldPath.Key(systemName)
			errors = append(errors, field.Invalid(keyFldPath, resource.Spec.BackendUsages[systemName], "backend usage does not have valid backend reference."))
		}
	}

	return errors, nil
}

func (r *ReconcileProduct) backendList(resource *capabilitiesv1beta1.Product, productProviderAccount *helper.ProviderAccount) ([]capabilitiesv1beta1.Backend, error) {
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

		backendProviderAccount, err := helper.LookupProviderAccount(r.Client(), resource.Namespace, backendList.Items[idx].Spec.ProviderAccountRef, r.Logger())
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

func findBackendBySystemName(list []capabilitiesv1beta1.Backend, systemName string) bool {
	for idx := range list {
		if list[idx].Spec.SystemName == systemName {
			return true
		}
	}
	return false
}
