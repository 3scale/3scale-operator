package product

import (
	"context"
	"fmt"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/pkg/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileProduct) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := r.Logger().WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("ReconcileProduct", "Operator version", version.Version)

	// Fetch the Product instance
	product := &capabilitiesv1beta1.Product{}
	err := r.Client().Get(r.Context(), request.NamespacedName, product)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Ignore deleted Products, this can happen when foregroundDeletion is enabled
	// https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#foreground-cascading-deletion
	if product.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	result, err := r.reconcile(product)
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile Product")
		r.EventRecorder().Eventf(product, corev1.EventTypeWarning, "ReconcileError", "%v", err)
	}
	reqLogger.Info("ReconcileProduct END", "result", result, "error", err)
	return result, err
}

func (r *ReconcileProduct) reconcile(product *capabilitiesv1beta1.Product) (reconcile.Result, error) {
	if product.SetDefaults() {
		return reconcile.Result{Requeue: true}, r.Client().Update(r.Context(), product)
	}

	productLogicReconciler := NewProductLogicReconciler(r.BaseReconciler, product)

	res, syncErr := productLogicReconciler.Reconcile()

	if syncErr == nil && res.Requeue {
		return res, nil
	}

	res, statusErr := productLogicReconciler.UpdateStatus()
	if statusErr != nil {
		if syncErr != nil {
			return reconcile.Result{}, fmt.Errorf("Failed to sync product: %v. Failed to update product status: %w", syncErr, statusErr)
		}

		return reconcile.Result{}, fmt.Errorf("Failed to update product status: %w", statusErr)
	}

	if syncErr != nil {
		return reconcile.Result{}, fmt.Errorf("Failed to sync product: %w", syncErr)
	}

	return reconcile.Result{}, nil
}
