package openapi

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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	// controllerName is the name of this controller
	controllerName = "controller_openapi"

	// package level logger
	log = logf.Log.WithName(controllerName)
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Openapi Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	return &ReconcileOpenapi{
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

	// Watch for changes to primary resource Openapi
	err = c.Watch(&source.Kind{Type: &capabilitiesv1beta1.Openapi{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileOpenapi implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileOpenapi{}

// ReconcileOpenapi reconciles a Openapi object
type ReconcileOpenapi struct {
	*reconcilers.BaseReconciler
}

// Reconcile reads that state of the cluster for a Openapi object and makes changes based on the state read
// and what is in the Openapi.Spec
func (r *ReconcileOpenapi) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := r.Logger().WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconcile Openapi", "Operator version", version.Version)

	// Fetch the Openapi instance
	openapiCR := &capabilitiesv1beta1.Openapi{}
	err := r.Client().Get(context.TODO(), request.NamespacedName, openapiCR)
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
		jsonData, err := json.MarshalIndent(openapiCR, "", "  ")
		if err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted Openapi, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if openapiCR.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	if openapiCR.SetDefaults(reqLogger) {
		err := r.Client().Update(r.Context(), openapiCR)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("Failed setting openapi defaults: %w", err)
		}

		reqLogger.Info("resource defaults updated. Requeueing.")
		return reconcile.Result{Requeue: true}, nil
	}

	statusReconciler, reconcileStatus, reconcileErr := r.reconcileSpec(openapiCR)
	statusResult, statusUpdateErr := statusReconciler.Reconcile()
	if statusUpdateErr != nil {
		if reconcileErr != nil {
			return reconcile.Result{}, fmt.Errorf("Failed to reconcile openapi: %v. Failed to update openapi status: %w", reconcileErr, statusUpdateErr)
		}

		return reconcile.Result{}, fmt.Errorf("Failed to update openapi status: %w", statusUpdateErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if reconcileErr != nil {
		if helper.IsInvalidSpecError(reconcileErr) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			reqLogger.Info("ERROR", "spec validation error", reconcileErr)
			r.EventRecorder().Eventf(openapiCR, corev1.EventTypeWarning, "Invalid Openapi Spec", "%v", reconcileErr)
			return reconcile.Result{}, nil
		}

		reqLogger.Error(reconcileErr, "Failed to reconcile")
		r.EventRecorder().Eventf(openapiCR, corev1.EventTypeWarning, "ReconcileError", "%v", reconcileErr)
		return reconcile.Result{}, reconcileErr
	}

	return reconcileStatus, nil
}

func (r *ReconcileOpenapi) reconcileSpec(openapiCR *capabilitiesv1beta1.Openapi) (*StatusReconciler, reconcile.Result, error) {
	logger := r.Logger().WithValues("openapi", openapiCR.Name)

	err := r.validateSpec(openapiCR)
	if err != nil {
		statusReconciler := NewStatusReconciler(r.BaseReconciler, openapiCR, "", err, false)
		return statusReconciler, reconcile.Result{}, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), openapiCR.Namespace, openapiCR.Spec.ProviderAccountRef, logger)
	if err != nil {
		statusReconciler := NewStatusReconciler(r.BaseReconciler, openapiCR, "", err, false)
		return statusReconciler, reconcile.Result{}, err
	}

	productSynced, err := r.checkProductSynced(openapiCR)
	if err != nil {
		statusReconciler := NewStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, false)
		return statusReconciler, reconcile.Result{}, err
	}

	statusReconciler := NewStatusReconciler(r.BaseReconciler, openapiCR, providerAccount.AdminURLStr, err, productSynced)
	return statusReconciler, reconcile.Result{Requeue: !productSynced}, err
}

func (r *ReconcileOpenapi) validateSpec(resource *capabilitiesv1beta1.Openapi) error {
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

func (r *ReconcileOpenapi) checkProductSynced(resource *capabilitiesv1beta1.Openapi) (bool, error) {
	// TODO check product resource is synced
	return true, nil
}
