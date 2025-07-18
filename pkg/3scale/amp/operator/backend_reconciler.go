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
	cronDeploymentMutator = append(cronDeploymentMutator,
		r.backendRedisTLSEnvVarMutator,
		r.backendQueuesRedisTLSEnvVarMutator,
		reconcilers.DeploymentVolumesMutator,
		reconcilers.DeploymentInitContainerVolumeMountsMutator,
		reconcilers.DeploymentContainerVolumeMountsMutator,
	)

	if r.apiManager.Spec.Backend.CronSpec.Replicas != nil {
		cronDeploymentMutator = append(cronDeploymentMutator, reconcilers.DeploymentReplicasMutator)
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
	listenerDeploymentMutator = append(listenerDeploymentMutator,
		r.backendRedisAsyncReconciler,
		r.backendRedisTLSEnvVarMutator,
		r.backendQueuesRedisTLSEnvVarMutator,
		reconcilers.DeploymentVolumesMutator,
		reconcilers.DeploymentInitContainerVolumeMountsMutator,
		reconcilers.DeploymentContainerVolumeMountsMutator,
	)

	if r.apiManager.Spec.Backend.ListenerSpec.Replicas != nil {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentReplicasMutator)
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
	workerDeploymentMutator = append(workerDeploymentMutator,
		r.workerRedisAsyncReconciler,
		r.backendRedisTLSEnvVarMutator,
		r.backendQueuesRedisTLSEnvVarMutator,
		backendDeploymentVolumesMutator,
		backendDeploymentInitContainerVolumeMountsMutator,
		backendDeploymentContainerVolumeMountsMutator,
	)

	if r.apiManager.Spec.Backend.WorkerSpec.Replicas != nil {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentReplicasMutator)
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

func (r *BackendReconciler) backendRedisTLSEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for Redis TLS
	var changed bool

	for _, envVar := range []string{
		"CONFIG_REDIS_CA_FILE",
		"CONFIG_REDIS_CERT",
		"CONFIG_REDIS_PRIVATE_KEY",
		"CONFIG_REDIS_SSL",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func (r *BackendReconciler) backendQueuesRedisTLSEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for Redis QUEUES TLS
	var changed bool

	for _, envVar := range []string{
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

func (r *BackendReconciler) backendRedisAsyncReconciler(desired, existing *k8sappsv1.Deployment) (bool, error) {
	var changed bool

	for _, envVar := range []string{
		"CONFIG_REDIS_ASYNC",
		"LISTENER_WORKERS",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func (r *BackendReconciler) workerRedisAsyncReconciler(desired, existing *k8sappsv1.Deployment) (bool, error) {
	var changed bool

	for _, envVar := range []string{
		"CONFIG_REDIS_ASYNC",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func backendDeploymentVolumesMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNames := []string{
		"backend-redis-tls",
		"queues-redis-tls",
	}
	return reconcilers.WeakDeploymentVolumesMutator(desired, existing, volumeNames)
}

func backendDeploymentInitContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeMountNames := []string{
		"backend-redis-tls",
		"queues-redis-tls",
	}
	return reconcilers.WeakDeploymentInitContainerVolumeMountsMutator(desired, existing, volumeMountNames)
}

func backendDeploymentContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeMountNames := []string{
		"backend-redis-tls",
		"queues-redis-tls",
	}
	return reconcilers.WeakDeploymentContainerVolumeMountsMutator(desired, existing, volumeMountNames)
}
