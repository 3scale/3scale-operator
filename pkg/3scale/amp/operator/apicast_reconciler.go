package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ApicastEnvCMReconciler struct {
}

func NewApicastEnvCMReconciler() *ApicastEnvCMReconciler {
	return &ApicastEnvCMReconciler{}
}

func (r *ApicastEnvCMReconciler) IsUpdateNeeded(desiredCM, existingCM *v1.ConfigMap) bool {
	update := false

	//	Check APICAST_MANAGEMENT_API
	fieldUpdated := ConfigMapReconcileField(desiredCM, existingCM, "APICAST_MANAGEMENT_API")
	update = update || fieldUpdated

	//	Check OPENSSL_VERIFY
	fieldUpdated = ConfigMapReconcileField(desiredCM, existingCM, "OPENSSL_VERIFY")
	update = update || fieldUpdated

	//	Check APICAST_RESPONSE_CODES
	fieldUpdated = ConfigMapReconcileField(desiredCM, existingCM, "APICAST_RESPONSE_CODES")
	update = update || fieldUpdated

	return update
}

type ApicastStagingDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewApicastDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *ApicastStagingDCReconciler {
	return &ApicastStagingDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ApicastStagingDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileReplicas(desired, existing, r.Logger())
	update = update || tmpUpdate

	tmpUpdate = DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

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
	apicast, err := Apicast(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileStagingDeploymentConfig(apicast.StagingDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileProductionDeploymentConfig(apicast.ProductionDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileStagingService(apicast.StagingService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileProductionService(apicast.ProductionService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileEnvironmentConfigMap(apicast.EnvironmentConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcilePodDisruptionBudget(apicast.StagingPodDisruptionBudget())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcilePodDisruptionBudget(apicast.ProductionPodDisruptionBudget())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ApicastReconciler) reconcileStagingDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewApicastDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *ApicastReconciler) reconcileProductionDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewApicastDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *ApicastReconciler) reconcileStagingService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *ApicastReconciler) reconcileProductionService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *ApicastReconciler) reconcileEnvironmentConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewApicastEnvCMReconciler())
	return reconciler.Reconcile(desiredConfigMap)
}

func Apicast(apimanager *appsv1alpha1.APIManager) (*component.Apicast, error) {
	optsProvider := NewApicastOptionsProvider(apimanager)
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}
	return component.NewApicast(opts), nil
}
