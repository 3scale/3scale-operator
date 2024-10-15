package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
	k8sappsv1 "k8s.io/api/apps/v1"
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
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	backend, err := Backend(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// Cron Deployment
	cronDeploymentMutator := reconcilers.GenericBackendDeploymentMutators()
	if r.apiManager.Spec.Backend.CronSpec.Replicas != nil {
		cronDeploymentMutator = append(cronDeploymentMutator, reconcilers.DeploymentReplicasMutator)
	}
	cronDeploymentMutator = append(cronDeploymentMutator, r.backendRedisTLSEnvVarMutator)
	if r.apiManager.Spec.Backend.BackendRedisTLSEnabled == nil ||
		(r.apiManager.Spec.Backend.BackendRedisTLSEnabled != nil && *r.apiManager.Spec.Backend.BackendRedisTLSEnabled == false) {
		cronDeploymentMutator = append(cronDeploymentMutator, reconcilers.DeploymentBackendRedisTLSRemoveEnvMutator)
		cronDeploymentMutator = append(cronDeploymentMutator, reconcilers.DeploymentBackendRedisTLSRemoveCertsMountMutator)
	} else {
		cronDeploymentMutator = append(cronDeploymentMutator, reconcilers.DeploymentBackendRedisTLSAddCertsMountMutator)
	}
	if r.apiManager.Spec.Backend.QueuesRedisTLSEnabled == nil ||
		(r.apiManager.Spec.Backend.QueuesRedisTLSEnabled != nil && *r.apiManager.Spec.Backend.QueuesRedisTLSEnabled == false) {
		cronDeploymentMutator = append(cronDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSRemoveEnvMutator)
		cronDeploymentMutator = append(cronDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSRemoveCertsMountMutator)
	} else {
		cronDeploymentMutator = append(cronDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSAddCertsMountMutator)
	}

	cronDeployment, err := backend.CronDeployment(r.Context(), r.Client(), ampImages.Options.BackendImage)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileDeployment(cronDeployment, reconcilers.DeploymentMutator(cronDeploymentMutator...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.BackendCronName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	listenerDeploymentMutator := reconcilers.GenericBackendDeploymentMutators()
	if r.apiManager.IsAsyncDisableAnnotationPresent() {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentListenerAsyncDisableArgsMutator)
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentListenerAsyncDisableEnvMutator)
	} else {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentListenerEnvMutator)
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentListenerArgsMutator)
	}
	if r.apiManager.Spec.Backend.ListenerSpec.Replicas != nil {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentReplicasMutator)
	}
	listenerDeploymentMutator = append(listenerDeploymentMutator, r.backendRedisTLSEnvVarMutator)
	if r.apiManager.Spec.Backend.BackendRedisTLSEnabled == nil ||
		(r.apiManager.Spec.Backend.BackendRedisTLSEnabled != nil && *r.apiManager.Spec.Backend.BackendRedisTLSEnabled == false) {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentBackendRedisTLSRemoveEnvMutator)
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentBackendRedisTLSRemoveCertsMountMutator)
	} else {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentBackendRedisTLSAddCertsMountMutator)
	}
	if r.apiManager.Spec.Backend.QueuesRedisTLSEnabled == nil ||
		(r.apiManager.Spec.Backend.QueuesRedisTLSEnabled != nil && *r.apiManager.Spec.Backend.QueuesRedisTLSEnabled == false) {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSRemoveEnvMutator)
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSRemoveCertsMountMutator)
	} else {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSAddCertsMountMutator)
	}

	listenerDeployment, err := backend.ListenerDeployment(r.Context(), r.Client(), ampImages.Options.BackendImage)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileDeployment(listenerDeployment, reconcilers.DeploymentMutator(listenerDeploymentMutator...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err = upgrade.MigrateDeploymentConfigToDeployment(component.BackendListenerName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	serviceMutators := []reconcilers.MutateFn{
		reconcilers.CreateOnlyMutator,
		reconcilers.ServiceSelectorMutator,
	}

	// Listener Service
	err = r.ReconcileService(backend.ListenerService(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Listener Route
	err = r.ReconcileRoute(backend.ListenerRoute(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Worker Deployment
	workerDeploymentMutator := reconcilers.GenericBackendDeploymentMutators()
	if r.apiManager.IsAsyncDisableAnnotationPresent() {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentWorkerDisableAsyncEnvMutator)
	} else {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentWorkerEnvMutator)
	}
	if r.apiManager.Spec.Backend.WorkerSpec.Replicas != nil {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentReplicasMutator)
	}

	workerDeploymentMutator = append(workerDeploymentMutator, r.backendRedisTLSEnvVarMutator)
	if r.apiManager.Spec.Backend.BackendRedisTLSEnabled == nil ||
		(r.apiManager.Spec.Backend.BackendRedisTLSEnabled != nil && *r.apiManager.Spec.Backend.BackendRedisTLSEnabled == false) {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentBackendRedisTLSRemoveEnvMutator)
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentBackendRedisTLSRemoveCertsMountMutator)

	} else {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentBackendRedisTLSAddCertsMountMutator)
	}
	if r.apiManager.Spec.Backend.QueuesRedisTLSEnabled == nil ||
		(r.apiManager.Spec.Backend.QueuesRedisTLSEnabled != nil && *r.apiManager.Spec.Backend.QueuesRedisTLSEnabled == false) {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSRemoveEnvMutator)
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSRemoveCertsMountMutator)
	} else {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentQueuesRedisTLSAddCertsMountMutator)
	}

	workerDeployment, err := backend.WorkerDeployment(r.Context(), r.Client(), ampImages.Options.BackendImage)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileDeployment(workerDeployment, reconcilers.DeploymentMutator(workerDeploymentMutator...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err = upgrade.MigrateDeploymentConfigToDeployment(component.BackendWorkerName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
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

	sumRate, err := helper.SumRateForOpenshiftVersion(r.Context(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboards(backend.BackendGrafanaV5Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboards(backend.BackendGrafanaV4Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(backend.BackendWorkerPrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(backend.BackendListenerPrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileHpa(component.DefaultHpa(component.BackendListenerName, r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileHpa(component.DefaultHpa(component.BackendWorkerName, r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
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

func containsAsyncDisable(m map[string]string, key, value string) bool {
	if v, ok := m[key]; ok {
		return v == value
	}
	return false
}

func (r *BackendReconciler) backendRedisTLSEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for Redis TLS
	var changed bool

	for _, envVar := range []string{
		"CONFIG_REDIS_CA_FILE",
		"CONFIG_REDIS_CERT",
		"CONFIG_REDIS_PRIVATE_KEY",
		"CONFIG_REDIS_SSL",
		"CONFIG_QUEUES_CA_FILE",
		"CONFIG_QUEUES_CERT",
		"CONFIG_QUEUES_PRIVATE_KEY",
		"CONFIG_QUEUES_SSL",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}
