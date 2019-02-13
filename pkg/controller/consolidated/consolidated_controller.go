package consolidated

import (
	"context"
	apiv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

var log = logf.Log.WithName("consolidated_controller")

// Add creates a new Consolidated Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileConsolidated{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("consolidated-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Consolidated
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Consolidated{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	//// TODO(user): Modify this to be the types you create that are owned by the primary resource
	//// Watch for changes to secondary resource Pods and requeue the owner Consolidated
	//err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &apiv1alpha1.Consolidated{},
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

var _ reconcile.Reconciler = &ReconcileConsolidated{}

// ReconcileConsolidated reconciles a Consolidated object
type ReconcileConsolidated struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// ReconcileWith3scale reads that state of the cluster for a Consolidated object and makes changes based on the state read
// and what is in the Consolidated.Spec
// TODO(user): Modify this ReconcileWith3scale function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileConsolidated) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Consolidated Object")

	// Fetch the Consolidated instance
	consolidated := &apiv1alpha1.Consolidated{}
	err := r.client.Get(context.TODO(), request.NamespacedName, consolidated)
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
	reqLogger.Info("Found consolidated object", request.Name, request.Namespace)

	if consolidated.DeletionTimestamp != nil && consolidated.GetFinalizers() != nil {

		reqLogger.Info("Terminating consolidated object", request.Name, request.Namespace)
		for _, api := range consolidated.Spec.APIs {
			_ = apiv1alpha1.DeleteInternalAPIFrom3scale(consolidated.Spec.Credentials, api)
		}

		//Remove finalizer
		consolidated.SetFinalizers(nil)
		err := r.client.Update(context.TODO(),consolidated)
		if err != nil {
			reqLogger.Error(err, "Error removing finalizer from consolidated object")
		}
		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil

	}

	existingState, err := apiv1alpha1.NewConsolidatedFrom3scale(consolidated.Spec.Credentials, consolidated.Spec.APIs)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	if apiv1alpha1.CompareConsolidated(*consolidated, *existingState) {
		reqLogger.Info("State between CRs and 3scale is consistent.", "namespace", request.Namespace, "name", request.Name)
		return reconcile.Result{Requeue: true, RequeueAfter: 2 * time.Minute}, nil
	}

	reqLogger.Info("State is not consistent, reconciling", "namespace", request.Namespace, "name", request.Name)
	apisDiff := apiv1alpha1.DiffAPIs(consolidated.Spec.APIs, existingState.Spec.APIs)
	err = apisDiff.ReconcileWith3scale(consolidated.Spec.Credentials)
	if err != nil {
		reqLogger.Error(err, "Error Reconciling APIs")
	}

	return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
}
