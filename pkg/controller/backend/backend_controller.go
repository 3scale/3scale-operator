package backend

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
	controllerName = "controller_backend"

	// package level logger
	log = logf.Log.WithName(controllerName)
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Backend Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	return &ReconcileBackend{
		BaseReconciler: reconcilers.NewBaseReconciler(client, scheme, apiClientReader, ctx, log, discoveryClient, recorder),
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("backend-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Backend
	err = c.Watch(&source.Kind{Type: &capabilitiesv1beta1.Backend{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBackend implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBackend{}

// ReconcileBackend reconciles a Backend object
type ReconcileBackend struct {
	*reconcilers.BaseReconciler
}

// Reconcile reads that state of the cluster for a Backend object and makes changes based on the state read
// and what is in the Backend.Spec
func (r *ReconcileBackend) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := r.Logger().WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconcile Backend", "Operator version", version.Version)

	// Fetch the Backend instance
	backend := &capabilitiesv1beta1.Backend{}
	err := r.Client().Get(context.TODO(), request.NamespacedName, backend)
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
		jsonData, err := json.MarshalIndent(backend, "", "  ")
		if err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	// Ignore deleted Backends, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if backend.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	result, err := r.reconcile(backend)
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile")
		r.EventRecorder().Eventf(backend, corev1.EventTypeWarning, "ReconcileError", "%v", err)
	}
	reqLogger.Info("END", "result", result, "error", err)
	return result, err
}

func (r *ReconcileBackend) reconcile(backendResource *capabilitiesv1beta1.Backend) (reconcile.Result, error) {
	logger := r.Logger().WithValues("backend", backendResource.Name)

	if backendResource.SetDefaults(logger) {
		err := r.Client().Update(r.Context(), backendResource)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("Failed setting backend defaults: %w", err)
		}

		logger.Info("resource defaults updated. Requeueing.")
		return reconcile.Result{Requeue: true}, nil
	}

	err := r.validateSpec(backendResource)
	if err != nil {
		if helper.IsInvalidSpecError(err) {
			// On Validation error, no need to retry as spec is not valid and needs to be changed
			logger.Info("ERROR", "spec validation error", err)
			r.EventRecorder().Eventf(backendResource, corev1.EventTypeWarning, "Invalid Backend Spec", "%v", err)
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	providerAccount, err := controllerhelper.LookupProviderAccount(r.Client(), backendResource.Namespace, backendResource.Spec.ProviderAccountRef, r.Logger())
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("reconcile backend spec: %w", err)
	}

	threescaleAPIClient, err := controllerhelper.PortaClient(providerAccount)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("reconcile backend spec: %w", err)
	}

	backendRemoteIndex, err := controllerhelper.NewBackendAPIRemoteIndex(threescaleAPIClient, r.Logger())
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("reconcile backend spec: %w", err)
	}

	reconciler := NewThreescaleReconciler(r.BaseReconciler, backendResource, threescaleAPIClient, backendRemoteIndex)
	backendAPIEntity, syncErr := reconciler.Reconcile()

	statusReconciler := NewStatusReconciler(r.BaseReconciler, backendResource, backendAPIEntity, providerAccount.AdminURLStr, syncErr)
	statusResult, statusErr := statusReconciler.Reconcile()
	if statusErr != nil {
		if syncErr != nil {
			return reconcile.Result{}, fmt.Errorf("Failed to sync backend: %v. Failed to update backend status: %w", syncErr, statusErr)
		}

		return reconcile.Result{}, fmt.Errorf("Failed to update backend status: %w", statusErr)
	}

	if statusResult.Requeue {
		return statusResult, nil
	}

	if syncErr != nil {
		return reconcile.Result{}, fmt.Errorf("Failed to sync backend: %w", syncErr)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBackend) validateSpec(backendResource *capabilitiesv1beta1.Backend) error {
	errors := field.ErrorList{}
	// internal validation
	errors = append(errors, backendResource.Validate()...)

	if len(errors) == 0 {
		return nil
	}

	return &helper.SpecFieldError{
		ErrorType:      helper.InvalidError,
		FieldErrorList: errors,
	}
}
