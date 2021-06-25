package operator

import (
	"fmt"
	"reflect"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	apicast, err := Apicast(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// Staging DC
	stagingDCMutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigReplicasMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		r.apicastLogLevelEnvVarMutator,
		r.volumeMountsMutator,
		r.volumesMutator,
	)
	err = r.ReconcileDeploymentConfig(apicast.StagingDeploymentConfig(), stagingDCMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Production DC
	productionDCMutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigReplicasMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		r.apicastProductionWorkersEnvVarMutator,
		r.apicastLogLevelEnvVarMutator,
		r.volumeMountsMutator,
		r.volumesMutator,
	)
	err = r.ReconcileDeploymentConfig(apicast.ProductionDeploymentConfig(), productionDCMutator)
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

	err = r.ReconcileGrafanaDashboard(apicast.ApicastMainAppGrafanaDashboard(), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(apicast.ApicastServicesGrafanaDashboard(), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(apicast.ApicastPrometheusRules(), reconcilers.CreateOnlyMutator)
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

func (r *ApicastReconciler) apicastProductionWorkersEnvVarMutator(desired, existing *appsv1.DeploymentConfig) bool {
	// Reconcile EnvVar only for "APICAST_WORKERS"
	return reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "APICAST_WORKERS")
}

func (r *ApicastReconciler) apicastLogLevelEnvVarMutator(desired, existing *appsv1.DeploymentConfig) bool {
	// Reconcile EnvVar only for "APICAST_LOG_LEVEL"
	return reconcilers.DeploymentConfigEnvVarReconciler(desired, existing, "APICAST_LOG_LEVEL")
}

func (r *ApicastReconciler) volumeMountsMutator(desired, existing *appsv1.DeploymentConfig) bool {
	changed := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].VolumeMounts, desired.Spec.Template.Spec.Containers[0].VolumeMounts) {
		changed = true
		existing.Spec.Template.Spec.Containers[0].VolumeMounts = desired.Spec.Template.Spec.Containers[0].VolumeMounts
	}

	return changed
}

func (r *ApicastReconciler) volumesMutator(desired, existing *appsv1.DeploymentConfig) bool {
	changed := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Volumes, desired.Spec.Template.Spec.Volumes) {
		changed = true
		existing.Spec.Template.Spec.Volumes = desired.Spec.Template.Spec.Volumes
	}

	return changed
}

func Apicast(apimanager *appsv1alpha1.APIManager, cl client.Client) (*component.Apicast, error) {
	optsProvider := NewApicastOptionsProvider(apimanager, cl)
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}
	return component.NewApicast(opts), nil
}
