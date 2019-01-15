package consolidated

import (
	"context"
	"log"
	"reflect"

	apiv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

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

// Reconcile reads that state of the cluster for a Consolidated object and makes changes based on the state read
// and what is in the Consolidated.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileConsolidated) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling Consolidated %s/%s\n", request.Namespace, request.Name)

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

	log.Printf("Detected Consolidated: %s, %s ", request.Name, request.Namespace)

	existingState, err := apiv1alpha1.NewConsolidatedFrom3scale()
	if err != nil {
		return reconcile.Result{Requeue:true}, err
	}

	if reflect.DeepEqual(consolidated, existingState) {
		return reconcile.Result{}, nil
	} else {
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}
