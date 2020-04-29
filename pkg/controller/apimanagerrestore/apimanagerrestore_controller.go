package apimanagerrestore

import (
	"context"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/restore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_apimanagerrestore")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new APIManagerRestore Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	apiClientReader, err := newAPIClientReader(mgr)
	if err != nil {
		return nil, err
	}

	client := mgr.GetClient()
	scheme := mgr.GetScheme()
	ctx := context.TODO()
	return &ReconcileAPIManagerRestore{
		BaseReconciler: reconcilers.NewBaseReconciler(client, scheme, apiClientReader, ctx, log),
	}, nil
}

// We create an Client Reader that directly queries the API server
// without going to the Cache provided by the Manager's Client because
// there are some resources that do not implement Watch (like ImageStreamTag)
// and the Manager's Client always tries to use the Cache when reading
func newAPIClientReader(mgr manager.Manager) (client.Client, error) {
	return client.New(mgr.GetConfig(), client.Options{Mapper: mgr.GetRESTMapper(), Scheme: mgr.GetScheme()})
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("apimanagerrestore-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource APIManagerRestore
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.APIManagerRestore{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner APIManagerRestore
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.APIManagerRestore{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAPIManagerRestore implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAPIManagerRestore{}

// ReconcileAPIManagerRestore reconciles a APIManagerRestore object
type ReconcileAPIManagerRestore struct {
	*reconcilers.BaseReconciler
}

// Reconcile reads that state of the cluster for a APIManagerRestore object and makes changes based on the state read
// and what is in the APIManagerRestore.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAPIManagerRestore) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling APIManagerRestore")

	// Fetch the APIManagerRestore instance
	instance, err := r.getAPIManagerRestoreCR(request)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Logger().Info("APIManagerRestore not found")
			return reconcile.Result{}, nil
		}
		r.Logger().Error(err, "Error getting APIManagerRestore")
		return reconcile.Result{}, err
	}

	res, err := r.setAPIManagerRestoreDefaults(instance)
	if err != nil {
		logger.Error(err, "Error")
		return reconcile.Result{}, err
	}
	if res.Requeue {
		logger.Info("Defaults set for APIManagerRestore resource")
		return res, nil
	}

	// TODO prepare / implement something related to version annotations or upgrade?

	apiManagerRestoreLogicReconciler, err := r.apiManagerRestoreLogicReconciler(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err = apiManagerRestoreLogicReconciler.Reconcile()
	if err != nil {
		logger.Error(err, "Error during reconciliation")
		return res, err
	}
	if res.Requeue {
		logger.Info("Reconciling not finished. Requeueing.")
		return res, nil
	}

	logger.Info("Reconciliation finished")
	return reconcile.Result{}, nil
}

func (r *ReconcileAPIManagerRestore) getAPIManagerRestoreCR(request reconcile.Request) (*appsv1alpha1.APIManagerRestore, error) {
	instance := appsv1alpha1.APIManagerRestore{}
	err := r.Client().Get(context.TODO(), request.NamespacedName, &instance)
	return &instance, err
}

func (r *ReconcileAPIManagerRestore) setAPIManagerRestoreDefaults(cr *appsv1alpha1.APIManagerRestore) (reconcile.Result, error) {
	changed, err := cr.SetDefaults() // TODO check where to put this
	if err != nil {
		return reconcile.Result{}, err
	}

	if changed {
		err = r.Client().Update(context.TODO(), cr)
	}

	return reconcile.Result{Requeue: changed}, err
}

func (r *ReconcileAPIManagerRestore) apiManagerRestoreLogicReconciler(cr *appsv1alpha1.APIManagerRestore) (*APIManagerRestoreLogicReconciler, error) {
	apiManagerRestoreOptionsProvider := restore.NewAPIManagerRestoreOptionsProvider(cr, r.BaseReconciler.Client())
	options, err := apiManagerRestoreOptionsProvider.Options()
	if err != nil {
		return nil, err
	}

	apiManagerRestore := restore.NewAPIManagerRestore(options)
	return NewAPIManagerRestoreLogicReconciler(r.BaseReconciler, cr, apiManagerRestore), nil

}
