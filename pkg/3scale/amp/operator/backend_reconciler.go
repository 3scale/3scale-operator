package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BackendReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewBackendReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *BackendReconciler {
	return &BackendReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *BackendReconciler) Reconcile() (reconcile.Result, error) {
	backend, err := Backend(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// Cron DC
	err = r.ReconcileDeploymentConfig(backend.CronDeploymentConfig(), reconcilers.GenericDeploymentConfigMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Listerner DC
	err = r.ReconcileDeploymentConfig(backend.ListenerDeploymentConfig(), reconcilers.GenericDeploymentConfigMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Listener Service
	err = r.ReconcileService(backend.ListenerService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Listener Route
	err = r.ReconcileRoute(backend.ListenerRoute(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Worker DC
	err = r.ReconcileDeploymentConfig(backend.WorkerDeploymentConfig(), reconcilers.GenericDeploymentConfigMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Environment ConfigMap
	err = r.ReconcileConfigMap(backend.EnvironmentConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Internal API Secret
	err = r.ReconcileSecret(backend.InternalAPISecretForSystem(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Redis Secret
	err = r.ReconcileSecret(backend.RedisSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Listener Secret
	err = r.ReconcileSecret(backend.ListenerSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Worker PDB
	err = r.ReconcilePodDisruptionBudget(backend.WorkerPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Cron PDB
	err = r.ReconcilePodDisruptionBudget(backend.CronPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Listener PDB
	err = r.ReconcilePodDisruptionBudget(backend.ListenerPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(backend.BackendWorkerPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(backend.BackendListenerPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(component.BackendGrafanaDashboard(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(component.BackendWorkerPrometheusRules(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func Backend(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.Backend, error) {
	optsProvider := NewOperatorBackendOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetBackendOptions()
	if err != nil {
		return nil, err
	}
	return component.NewBackend(opts), nil
}
