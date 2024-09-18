package operator

import (
	"context"
	"fmt"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
)

// RedisDependencyReconciler is a generic DependencyReconciler that reconciles
// an internal Redis instance using the Redis options
type RedisReconciler struct {
	*BaseAPIManagerLogicReconciler

	DeploymentConfig      func(redis *component.Redis) *appsv1.DeploymentConfig
	Service               func(redis *component.Redis) *corev1.Service
	ConfigMap             func(redis *component.RedisConfigMap) *corev1.ConfigMap
	PersistentVolumeClaim func(redis *component.Redis) *corev1.PersistentVolumeClaim
	ImageStream           func(redis *component.Redis) *imagev1.ImageStream
	Secret                func(redis *component.Redis) *corev1.Secret
}

var _ DependencyReconciler = &RedisReconciler{}

func NewSystemRedisDependencyReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &RedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,

		DeploymentConfig:      (*component.Redis).SystemDeploymentConfig,
		Service:               (*component.Redis).SystemService,
		ConfigMap:             (*component.RedisConfigMap).ConfigMap,
		PersistentVolumeClaim: (*component.Redis).SystemPVC,
		ImageStream:           (*component.Redis).SystemImageStream,
		Secret:                (*component.Redis).SystemRedisSecret,
	}
}

func NewBackendRedisDependencyReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &RedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,

		DeploymentConfig:      (*component.Redis).BackendDeploymentConfig,
		Service:               (*component.Redis).BackendService,
		ConfigMap:             (*component.RedisConfigMap).ConfigMap,
		PersistentVolumeClaim: (*component.Redis).BackendPVC,
		ImageStream:           (*component.Redis).BackendImageStream,
		Secret:                (*component.Redis).BackendRedisSecret,
	}
}

func (r *RedisReconciler) Reconcile() (reconcile.Result, error) {
	redisConfigMapGenerator, err := RedisConfigMap(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// We want to reconcile redis-conf ConfigMap before Deployment
	// to avoid restart redis pods twice in case of User change ConfigMap.
	// Annotation "redisConfigMapResourceVersion" added to Pod Template to support Upgrade scenario.
	// this "redisConfigMapResourceVersion" is taken from ConfigMap resourceVersion.
	// If User changes Config Map, Operator will revert it to original one,
	// but resourceVersion could be changed twice (after user change and after operator).
	// To avoid this scenario by placing CM reconciliation before Deployment
	err = r.ReconcileConfigMap(r.ConfigMap(redisConfigMapGenerator), r.redisConfigMapMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// check config map exists, otherwise requeue.
	cmKey := client.ObjectKeyFromObject(r.ConfigMap(redisConfigMapGenerator))
	err = r.Client().Get(context.Background(), cmKey, &corev1.ConfigMap{})
	if apierrors.IsNotFound(err) {
		r.logger.Info("waiting for redis config map to be available")
		return reconcile.Result{Requeue: true}, nil
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	redis, err := Redis(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	dcMutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigImageChangeTriggerMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		reconcilers.DeploymentConfigPodTemplateLabelsMutator,
		reconcilers.DeploymentConfigPriorityClassMutator,
		reconcilers.DeploymentConfigTopologySpreadConstraintsMutator,
		reconcilers.DeploymentConfigPodTemplateAnnotationsMutator,
		// 3scale 2.13 -> 2.14
		upgrade.Redis6CommandArgsEnv,
	)
	err = r.ReconcileDeploymentConfig(r.DeploymentConfig(redis), dcMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// redis Service
	err = r.ReconcileService(r.Service(redis), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
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

func Redis(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.Redis, error) {
	optsProvider := NewRedisOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetRedisOptions()
	if err != nil {
		return nil, err
	}
	return component.NewRedis(opts), nil
}

func (r *RedisReconciler) redisConfigMapMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*corev1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", existingObj)
	}
	desired, ok := desiredObj.(*corev1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", desiredObj)
	}

	update := false
	fieldUpdated := reconcilers.RedisConfigMapReconcileField(desired, existing, "redis.conf")
	update = update || fieldUpdated

	return update, nil
}

func RedisConfigMap(apimanager *appsv1alpha1.APIManager) (*component.RedisConfigMap, error) {
	optsProvider := NewRedisConfigMapOptionsProvider(apimanager)
	opts, err := optsProvider.GetOptions()
	if err != nil {
		return nil, err
	}
	return component.NewRedisConfigMap(opts), nil
}
