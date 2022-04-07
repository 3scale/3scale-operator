package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RedisReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewRedisReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *RedisReconciler {
	return &RedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

// RedisDependencyReconciler is a generic DependencyReconciler that reconciles
// an internal Redis instance using the Redis options
type RedisDependencyReconciler struct {
	*BaseAPIManagerLogicReconciler

	DeploymentConfig      func(redis *component.Redis) *appsv1.DeploymentConfig
	Service               func(redis *component.Redis) *corev1.Service
	ConfigMap             func(redis *component.Redis) *corev1.ConfigMap
	PersistentVolumeClaim func(redis *component.Redis) *corev1.PersistentVolumeClaim
	ImageStream           func(redis *component.Redis) *imagev1.ImageStream
	Secret                func(redis *component.Redis) *corev1.Secret
}

var _ DependencyReconciler = &RedisDependencyReconciler{}

func NewSystemRedisDependencyReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &RedisDependencyReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,

		DeploymentConfig:      (*component.Redis).SystemDeploymentConfig,
		Service:               (*component.Redis).SystemService,
		ConfigMap:             nil,
		PersistentVolumeClaim: (*component.Redis).SystemPVC,
		ImageStream:           (*component.Redis).SystemImageStream,
		Secret:                (*component.Redis).SystemRedisSecret,
	}
}

func NewBackendRedisDependencyReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &RedisDependencyReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,

		DeploymentConfig:      (*component.Redis).BackendDeploymentConfig,
		Service:               (*component.Redis).BackendService,
		ConfigMap:             (*component.Redis).BackendConfigMap,
		PersistentVolumeClaim: (*component.Redis).BackendPVC,
		ImageStream:           (*component.Redis).BackendImageStream,
		Secret:                (*component.Redis).BackendRedisSecret,
	}
}

func (r *RedisDependencyReconciler) Reconcile() (reconcile.Result, error) {
	redis, err := Redis(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	dcMutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
	)
	err = r.ReconcileDeploymentConfig(r.DeploymentConfig(redis), dcMutator)

	// redis Service
	err = r.ReconcileService(r.Service(redis), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// CM
	if r.ConfigMap != nil {
		err = r.ReconcileConfigMap(r.ConfigMap(redis), reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// PVC
	err = r.ReconcilePersistentVolumeClaim(r.PersistentVolumeClaim(redis), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// IS
	err = r.ReconcileImagestream(r.ImageStream(redis), reconcilers.GenericImageStreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Redis Secret
	err = r.ReconcileSecret(r.Secret(redis), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *RedisReconciler) Reconcile() (reconcile.Result, error) {
	redis, err := Redis(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backend redis DC
	backendDCMutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
	)
	err = r.ReconcileDeploymentConfig(redis.BackendDeploymentConfig(), backendDCMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// backend redis Service
	err = r.ReconcileService(redis.BackendService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// backend CM
	err = r.ReconcileConfigMap(redis.BackendConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backenb PVC
	err = r.ReconcilePersistentVolumeClaim(redis.BackendPVC(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backend IS
	err = r.ReconcileImagestream(redis.BackendImageStream(), reconcilers.GenericImageStreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backend Redis Secret
	err = r.ReconcileSecret(redis.BackendRedisSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis DC
	systemDCMutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
	)
	err = r.ReconcileDeploymentConfig(redis.SystemDeploymentConfig(), systemDCMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis Service
	err = r.ReconcileService(redis.SystemService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis PVC
	err = r.ReconcilePersistentVolumeClaim(redis.SystemPVC(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis IS
	err = r.ReconcileImagestream(redis.SystemImageStream(), reconcilers.GenericImageStreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System Redis Secret
	err = r.ReconcileSecret(redis.SystemRedisSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func Redis(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.Redis, error) {
	optsProvider := NewRedisOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetRedisOptions()
	if err != nil {
		return nil, err
	}
	return component.NewRedis(opts), nil
}
