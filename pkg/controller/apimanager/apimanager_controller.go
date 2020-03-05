package apimanager

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"k8s.io/api/policy/v1beta1"

	"github.com/3scale/3scale-operator/version"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"

	appsv1 "github.com/openshift/api/apps/v1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/RHsyseng/operator-utils/pkg/olm"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_apimanager")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new APIManager Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, reconciler)
}

// We create an Client Reader that directly queries the API server
// without going to the Cache provided by the Manager's Client because
// there are some resources that do not implement Watch (like ImageStreamTag)
// and the Manager's Client always tries to use the Cache when reading
func newAPIClientReader(mgr manager.Manager) (client.Client, error) {
	return client.New(mgr.GetConfig(), client.Options{Mapper: mgr.GetRESTMapper(), Scheme: mgr.GetScheme()})
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	apiClientReader, err := newAPIClientReader(mgr)
	if err != nil {
		return nil, err
	}

	BaseReconciler := operator.NewBaseReconciler(mgr.GetClient(), apiClientReader, mgr.GetScheme(), log, mgr.GetConfig())
	return &ReconcileAPIManager{
		BaseControllerReconciler: operator.NewBaseControllerReconciler(BaseReconciler),
	}, nil

}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("apimanager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource APIManager
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.APIManager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to DeploymentConfigs to update deployment status
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.APIManager{},
	}
	err = c.Watch(&source.Kind{Type: &appsv1.DeploymentConfig{}}, ownerHandler)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1beta1.PodDisruptionBudget{}}, ownerHandler)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAPIManager implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAPIManager{}

// ReconcileAPIManager reconciles a APIManager object
type ReconcileAPIManager struct {
	operator.BaseControllerReconciler
}

func (r *ReconcileAPIManager) updateVersionAnnotations(cr *appsv1alpha1.APIManager) error {
	if cr.Annotations == nil {
		cr.Annotations = map[string]string{}
	}
	cr.Annotations[appsv1alpha1.ThreescaleVersionAnnotation] = product.ThreescaleRelease
	cr.Annotations[appsv1alpha1.OperatorVersionAnnotation] = version.Version
	return r.Client().Update(context.TODO(), cr)
}

func (r *ReconcileAPIManager) upgradeAPIManager(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	// The object to instantiate would change in every release of the operator
	// that upgrades the threescale version
	upgradeApiManager := &operator.UpgradeApiManager{
		Client:          r.Client(),
		ApiClientReader: r.APIClientReader(),
		Scheme:          r.Scheme(),
		Cr:              cr,
		Logger:          r.Logger(),
		Cfg:             r.Config(),
	}
	return upgradeApiManager.Upgrade()
}

// Reconcile reads that state of the cluster for a APIManager object and makes changes based on the state read
// and what is in the APIManager.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAPIManager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger().WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("ReconcileAPIManager", "Operator version", version.Version, "3scale release", product.ThreescaleRelease)

	instance, err := r.apiManagerInstance(request.NamespacedName)
	if err != nil {
		logger.Error(err, "Error fetching apimanager instance")
		return reconcile.Result{}, err
	}
	if instance == nil {
		logger.Info("resource not found. Ignoring since object must have been deleted")
		return reconcile.Result{}, nil
	}

	res, err := r.setAPIManagerDefaults(instance)
	if err != nil {
		logger.Error(err, "Error")
		return reconcile.Result{}, err
	}
	if res.Requeue {
		logger.Info("Defaults set for APIManager resource")
		return res, nil
	}

	if instance.Annotations[appsv1alpha1.OperatorVersionAnnotation] != version.Version {
		logger.Info(fmt.Sprintf("Upgrade %s -> %s", instance.Annotations[appsv1alpha1.OperatorVersionAnnotation], version.Version))
		// TODO add logic to check that only immediate consecutive installs
		// are possible?
		res, err := r.upgradeAPIManager(instance)
		if err != nil {
			logger.Error(err, "Error upgrading APIManager")
			return reconcile.Result{}, err
		}
		if res.Requeue {
			logger.Info("Upgrading not finished. Requeueing.")
			return res, nil
		}

		err = r.updateVersionAnnotations(instance)
		if err != nil {
			logger.Error(err, "Error updating annotations")
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	result, err := r.reconcileAPIManagerLogic(instance)
	if err != nil {
		logger.Error(err, "Error during reconciliation")
		return result, err
	}
	if result.Requeue {
		logger.Info("Reconciling not finished. Requeueing.")
		return result, nil
	}

	err = r.reconcileAPIManagerStatus(instance)
	if err != nil {
		logger.Error(err, "Error updating status")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileAPIManager) apiManagerInstance(namespacedName types.NamespacedName) (*appsv1alpha1.APIManager, error) {
	// Fetch the APIManager instance
	instance := &appsv1alpha1.APIManager{}

	err := r.Client().Get(context.TODO(), namespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil, nil
		}
		return nil, err
	}
	return instance, nil
}

func (r *ReconcileAPIManager) setAPIManagerDefaults(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	changed, err := cr.SetDefaults() // TODO check where to put this
	if err != nil {
		return reconcile.Result{}, err
	}

	if changed {
		err = r.Client().Update(context.TODO(), cr)
	}

	return reconcile.Result{Requeue: changed}, err
}

func (r *ReconcileAPIManager) reconcileAPIManagerLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	result, err := r.reconcileAMPImagesLogic(cr)
	if err != nil || result.Requeue {
		return result, err
	}

	if !cr.IsExternalDatabaseEnabled() {
		result, err = r.reconcileRedisLogic(cr)
		if err != nil || result.Requeue {
			return result, err
		}

		result, err = r.reconcileSystemDatabaseLogic(cr)
		if err != nil || result.Requeue {
			return result, err
		}
	} else {
		// External databases
		// validate required secrets exist
		err := r.externalDatabasesCheck(cr)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	result, err = r.reconcileBackendLogic(cr)
	if err != nil || result.Requeue {
		return result, err
	}

	result, err = r.reconcileMemcached(cr)
	if err != nil || result.Requeue {
		return result, err
	}

	result, err = r.reconcileSystem(cr)
	if err != nil || result.Requeue {
		return result, err
	}

	result, err = r.reconcileZync(cr)
	if err != nil || result.Requeue {
		return result, err
	}

	result, err = r.reconcileApicast(cr)
	if err != nil || result.Requeue {
		return result, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileAPIManager) reconcileAMPImagesLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewAMPImagesReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	return reconciler.Reconcile()
}

func (r *ReconcileAPIManager) reconcileRedisLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewRedisReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	return reconciler.Reconcile()
}

func (r *ReconcileAPIManager) reconcileBackendLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewBackendReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	return reconciler.Reconcile()
}

func (r *ReconcileAPIManager) reconcileSystemDatabaseLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	if cr.Spec.System.DatabaseSpec != nil && cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return r.reconcileSystemPostgreSQLLogic(cr)
	}

	// Defaults to MySQL
	return r.reconcileSystemMySQLLogic(cr)
}

func (r *ReconcileAPIManager) reconcileSystemPostgreSQLLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewSystemPostgreSQLReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	result, err := reconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	imageReconciler := operator.NewSystemPostgreSQLImageReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	result, err = imageReconciler.Reconcile()
	return result, err
}

func (r *ReconcileAPIManager) reconcileSystemMySQLLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewSystemMySQLReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	result, err := reconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	imageReconciler := operator.NewSystemMySQLImageReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	result, err = imageReconciler.Reconcile()
	return result, err
}

func (r *ReconcileAPIManager) reconcileMemcached(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewMemcachedReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	return reconciler.Reconcile()
}

func (r *ReconcileAPIManager) reconcileSystem(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewSystemReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	return reconciler.Reconcile()
}

func (r *ReconcileAPIManager) reconcileZync(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewZyncReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	return reconciler.Reconcile()
}

func (r *ReconcileAPIManager) reconcileApicast(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseLogicReconciler := operator.NewBaseLogicReconciler(r.BaseReconciler)
	reconciler := operator.NewApicastReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, cr))
	return reconciler.Reconcile()
}

func (r *ReconcileAPIManager) reconcileAPIManagerStatus(cr *appsv1alpha1.APIManager) error {
	return r.setDeploymentStatus(cr)
}

func (r *ReconcileAPIManager) setDeploymentStatus(instance *appsv1alpha1.APIManager) error {
	listOps := []client.ListOption{
		client.InNamespace(instance.Namespace),
	}
	dcList := &appsv1.DeploymentConfigList{}
	err := r.Client().List(context.TODO(), dcList, listOps...)
	if err != nil {
		r.Logger().Error(err, "Failed to list deployment configs")
		return err
	}
	var dcs []appsv1.DeploymentConfig
	for _, dc := range dcList.Items {
		for _, ownerRef := range dc.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				dcs = append(dcs, dc)
				break
			}
		}
	}
	sort.Slice(dcs, func(i, j int) bool { return dcs[i].Name < dcs[j].Name })

	deploymentStatus := olm.GetDeploymentConfigStatus(dcs)
	if !reflect.DeepEqual(instance.Status.Deployments, deploymentStatus) {
		r.Logger().Info("Deployment status will be updated")
		instance.Status.Deployments = deploymentStatus
		err = r.Client().Status().Update(context.TODO(), instance)
		if err != nil {
			r.Logger().Error(err, "Failed to update API Manager deployment status")
			return err
		}
	}
	return nil
}

func (r *ReconcileAPIManager) externalDatabasesCheck(cr *appsv1alpha1.APIManager) error {
	optsProvider := operator.NewHighAvailabilityOptionsProvider(cr.Namespace, r.Client())
	_, err := optsProvider.GetHighAvailabilityOptions()
	return err
}
