package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ApicastReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &ApicastReconciler{}

func NewApicastReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) ApicastReconciler {
	return ApicastReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ApicastReconciler) Reconcile() (reconcile.Result, error) {
	apicast, err := r.apicast()
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileStagingDeploymentConfig(apicast.StagingDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileProductionDeploymentConfig(apicast.ProductionDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileStagingService(apicast.StagingService())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileProductionService(apicast.ProductionService())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileEnvironmentConfigMap(apicast.EnvironmentConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ApicastReconciler) apicast() (*component.Apicast, error) {
	optsProvider := OperatorApicastOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}
	return component.NewApicast(opts), nil
}

func (r *ApicastReconciler) reconcileDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	err := r.InitializeAsAPIManagerObject(desiredDeploymentConfig)
	if err != nil {
		return err
	}

	return r.deploymentConfigReconciler.Reconcile(desiredDeploymentConfig)
}

func (r *ApicastReconciler) reconcileConfigMap(desiredConfigMap *v1.ConfigMap) error {
	err := r.InitializeAsAPIManagerObject(desiredConfigMap)
	if err != nil {
		return err
	}

	return r.configMapReconciler.Reconcile(desiredConfigMap)
}

func (r *ApicastReconciler) reconcileService(desiredService *v1.Service) error {
	err := r.InitializeAsAPIManagerObject(desiredService)
	if err != nil {
		return err
	}
	return r.serviceReconciler.Reconcile(desiredService)
}

func (r *ApicastReconciler) reconcileStagingDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *ApicastReconciler) reconcileProductionDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *ApicastReconciler) reconcileStagingService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *ApicastReconciler) reconcileProductionService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *ApicastReconciler) reconcileEnvironmentConfigMap(desiredConfigMap *v1.ConfigMap) error {
	return r.reconcileConfigMap(desiredConfigMap)
}
