package apimanagerbackup

import (
	"context"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	// controllerName is the name of this controller
	controllerName = "controller_apimanagerbackup"
	log            = logf.Log.WithName(controllerName)
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new APIManagerBackup Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	recorder := mgr.GetEventRecorderFor(controllerName)
	return &ReconcileAPIManagerBackup{
		BaseReconciler: reconcilers.NewBaseReconciler(client, scheme, apiClientReader, ctx, log, recorder),
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
	c, err := controller.New("controllerName", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource APIManagerBackup
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.APIManagerBackup{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAPIManagerBackup implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAPIManagerBackup{}

// ReconcileAPIManagerBackup reconciles a APIManagerBackup object
type ReconcileAPIManagerBackup struct {
	*reconcilers.BaseReconciler
}

// Reconcile reads that state of the cluster for a APIManagerBackup object and makes changes based on the state read
// and what is in the APIManagerBackup.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic. This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAPIManagerBackup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling APIManagerBackup")

	// Fetch the APIManagerBackup instance
	instance, err := r.getAPIManagerBackupCR(request)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Logger().Info("APIManagerBackup not found")
			return reconcile.Result{}, nil
		}
		r.Logger().Error(err, "Error getting APIManagerBackup")
		return reconcile.Result{}, err
	}

	res, err := r.setAPIManagerBackupDefaults(instance)
	if err != nil {
		logger.Error(err, "Error")
		return reconcile.Result{}, err
	}
	if res.Requeue {
		logger.Info("Defaults set for APIManagerBackup resource")
		return res, nil
	}

	// TODO prepare / implement something related to version annotations or upgrade?

	apiManagerBackupLogicReconciler, err := r.apiManagerBackupLogicReconciler(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err = apiManagerBackupLogicReconciler.Reconcile()
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

func (r *ReconcileAPIManagerBackup) getAPIManagerBackupCR(request reconcile.Request) (*appsv1alpha1.APIManagerBackup, error) {
	instance := appsv1alpha1.APIManagerBackup{}
	err := r.Client().Get(context.TODO(), request.NamespacedName, &instance)
	return &instance, err
}

func (r *ReconcileAPIManagerBackup) setAPIManagerBackupDefaults(cr *appsv1alpha1.APIManagerBackup) (reconcile.Result, error) {
	changed, err := cr.SetDefaults() // TODO check where to put this
	if err != nil {
		return reconcile.Result{}, err
	}

	if changed {
		err = r.Client().Update(context.TODO(), cr)
	}

	return reconcile.Result{Requeue: changed}, err
}

func (r *ReconcileAPIManagerBackup) apiManagerBackupLogicReconciler(cr *appsv1alpha1.APIManagerBackup) (*APIManagerBackupLogicReconciler, error) {
	return NewAPIManagerBackupLogicReconciler(r.BaseReconciler, cr)
}
