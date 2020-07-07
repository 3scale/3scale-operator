package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func ApicastEnvCMMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*v1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", existingObj)
	}
	desired, ok := desiredObj.(*v1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", desiredObj)
	}

	update := false

	//	Check APICAST_MANAGEMENT_API
	fieldUpdated := reconcilers.ConfigMapReconcileField(desired, existing, "APICAST_MANAGEMENT_API")
	update = update || fieldUpdated

	//	Check OPENSSL_VERIFY
	fieldUpdated = reconcilers.ConfigMapReconcileField(desired, existing, "OPENSSL_VERIFY")
	update = update || fieldUpdated

	//	Check APICAST_RESPONSE_CODES
	fieldUpdated = reconcilers.ConfigMapReconcileField(desired, existing, "APICAST_RESPONSE_CODES")
	update = update || fieldUpdated

	return update, nil
}

type ApicastReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewApicastReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *ApicastReconciler {
	return &ApicastReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ApicastReconciler) Reconcile() (reconcile.Result, error) {
	apicast, err := Apicast(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Staging DC
	err = r.ReconcileDeploymentConfig(apicast.StagingDeploymentConfig(), reconcilers.GenericDeploymentConfigMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Production DC
	err = r.ReconcileDeploymentConfig(apicast.ProductionDeploymentConfig(), reconcilers.GenericDeploymentConfigMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Staging Service
	err = r.ReconcileService(apicast.StagingService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Production Service
	err = r.ReconcileService(apicast.ProductionService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Environment ConfigMap
	err = r.ReconcileConfigMap(apicast.EnvironmentConfigMap(), ApicastEnvCMMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Staging PDB
	err = r.ReconcilePodDisruptionBudget(apicast.StagingPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Production PDB
	err = r.ReconcilePodDisruptionBudget(apicast.ProductionPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(component.ApicastMainAppGrafanaDashboard(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(component.ApicastServicesGrafanaDashboard(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(component.ApicastPrometheusRules(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(apicast.ApicastProductionPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(apicast.ApicastStagingPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func Apicast(apimanager *appsv1alpha1.APIManager) (*component.Apicast, error) {
	optsProvider := NewApicastOptionsProvider(apimanager)
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}
	return component.NewApicast(opts), nil
}
