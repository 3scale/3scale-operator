package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BackendReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &BackendReconciler{}

func NewBackendReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) BackendReconciler {
	return BackendReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *BackendReconciler) Reconcile() (reconcile.Result, error) {
	backend, err := r.backend()
	if err != nil {
		return reconcile.Result{}, err
	}

	// // TODO finish reconciliations
	r.reconcileCronDeploymentConfig(backend.CronDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileListenerDeploymentConfig(backend.ListenerDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileListenerService(backend.ListenerService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileListenerRoute(backend.ListenerRoute())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileWorkerDeploymentConfig(backend.WorkerDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileEnvironmentConfigMap(backend.EnvironmentConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileInternalAPISecretForSystem(backend.InternalAPISecretForSystem())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileRedisSecret(backend.RedisSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileListenerSecret(backend.ListenerSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *BackendReconciler) backend() (*component.Backend, error) {
	optsProvider := OperatorBackendOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetBackendOptions()
	if err != nil {
		return nil, err
	}
	return component.NewBackend(opts), nil
}

func (r *BackendReconciler) reconcileDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	err := r.InitializeAsAPIManagerObject(desiredDeploymentConfig)
	if err != nil {
		return err
	}

	return r.deploymentConfigReconciler.Reconcile(desiredDeploymentConfig)
}

func (r *BackendReconciler) reconcileRoute(desiredRoute *routev1.Route) error {
	err := r.InitializeAsAPIManagerObject(desiredRoute)
	if err != nil {
		return err
	}

	return r.routeReconciler.Reconcile(desiredRoute)
}

func (r *BackendReconciler) reconcileSecret(desiredSecret *v1.Secret) error {
	err := r.InitializeAsAPIManagerObject(desiredSecret)
	if err != nil {
		return err
	}
	return r.secretReconciler.Reconcile(desiredSecret)
}

func (r *BackendReconciler) reconcileConfigMap(desiredConfigMap *v1.ConfigMap) error {
	err := r.InitializeAsAPIManagerObject(desiredConfigMap)
	if err != nil {
		return err
	}

	return r.configMapReconciler.Reconcile(desiredConfigMap)
}

func (r *BackendReconciler) reconcileService(desiredService *v1.Service) error {
	err := r.InitializeAsAPIManagerObject(desiredService)
	if err != nil {
		return err
	}
	return r.serviceReconciler.Reconcile(desiredService)
}

func (r *BackendReconciler) reconcileCronDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *BackendReconciler) reconcileListenerDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *BackendReconciler) reconcileListenerService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *BackendReconciler) reconcileListenerRoute(desiredRoute *routev1.Route) error {
	return r.reconcileRoute(desiredRoute)
}

func (r *BackendReconciler) reconcileWorkerDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *BackendReconciler) reconcileEnvironmentConfigMap(desiredConfigMap *v1.ConfigMap) error {
	return r.reconcileConfigMap(desiredConfigMap)
}

func (r *BackendReconciler) reconcileInternalAPISecretForSystem(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *BackendReconciler) reconcileRedisSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *BackendReconciler) reconcileListenerSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}
