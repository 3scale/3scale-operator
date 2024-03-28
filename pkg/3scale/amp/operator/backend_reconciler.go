package operator

import (
	"context"
	"github.com/go-logr/logr"
	"strings"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

	err = r.ReconcileDeployment(backend.CronDeployment(ampImages.Options.BackendImage), reconcilers.DeploymentMutator(cronDeploymentMutator...))
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

	// Listener Deployment
	RedisQueuesUrl, RedisStorageUrl, RedisQueuesSentinelHost, RedisStorageSentinelHost := GetBackendRedisSecret(r.apiManager.Namespace, r.Context(), r.Client(), r.logger)

	listenerDeploymentMutator := reconcilers.GenericBackendDeploymentMutators()
	// this checks for logical redis exists
	// this checks if SentinelHost are configured with passwords
	if RedisStorageUrl != RedisQueuesUrl && !RedisQueuesSentinelHost && !RedisStorageSentinelHost {
		// this checks if SentinelHost are configured with passwords
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentListenerEnvMutator)
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentListenerArgsMutator)
	}
	if r.apiManager.Spec.Backend.ListenerSpec.Replicas != nil {
		listenerDeploymentMutator = append(listenerDeploymentMutator, reconcilers.DeploymentReplicasMutator)
	}

	err = r.ReconcileDeployment(backend.ListenerDeployment(ampImages.Options.BackendImage), reconcilers.DeploymentMutator(listenerDeploymentMutator...))
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
	if RedisStorageUrl != RedisQueuesUrl && !RedisQueuesSentinelHost && !RedisStorageSentinelHost {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentWorkerEnvMutator)
	}

	if r.apiManager.Spec.Backend.WorkerSpec.Replicas != nil {
		workerDeploymentMutator = append(workerDeploymentMutator, reconcilers.DeploymentReplicasMutator)
	}

	err = r.ReconcileDeployment(backend.WorkerDeployment(ampImages.Options.BackendImage), reconcilers.DeploymentMutator(workerDeploymentMutator...))
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

	err = r.ReconcileGrafanaDashboard(backend.BackendGrafanaDashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
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
	if RedisStorageUrl != RedisQueuesUrl && !RedisQueuesSentinelHost && !RedisStorageSentinelHost {
		err = r.ReconcileHpa(component.DefaultHpa(component.BackendListenerName, r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.ReconcileHpa(component.DefaultHpa(component.BackendWorkerName, r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		// set log message if logical redis db are detected in the backend
		if r.apiManager.Spec.Backend.ListenerSpec.Hpa || r.apiManager.Spec.Backend.WorkerSpec.Hpa {
			message := "logical redis instances or SentinelHost with authentication found in the backend, which is blocking redis async mode, horizontal pod autoscaling for backend cannot be enabled without async mode"
			r.logger.Info(message)
		}
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

func GetBackendRedisSecret(apimanagerNs string, ctx context.Context, client client.Client, logger logr.Logger) (string, string, bool, bool) {
	backendRedisSecret := &v1.Secret{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      "backend-redis",
		Namespace: apimanagerNs,
	}, backendRedisSecret)
	if err != nil {
		logger.Error(err, "Failed to get system-redis secret, can't check logical databases or authenticated redis sentinels, check the backend-redis secret exists")
		return "", "", false, false
	}
	RedisQueuesUrl := strings.TrimSuffix(string(backendRedisSecret.Data["REDIS_QUEUES_URL"]), "1")
	RedisStorageUrl := strings.TrimSuffix(string(backendRedisSecret.Data["REDIS_STORAGE_URL"]), "0")
	RedisQueuesSentinelHost := strings.Contains(string(backendRedisSecret.Data["REDIS_QUEUES_SENTINEL_HOSTS"]), "@")
	RedisStorageSentinelHost := strings.Contains(string(backendRedisSecret.Data["REDIS_STORAGE_SENTINEL_HOSTS"]), "@")

	return RedisQueuesUrl, RedisStorageUrl, RedisQueuesSentinelHost, RedisStorageSentinelHost
}
