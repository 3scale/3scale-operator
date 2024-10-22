package operator

import (
	"context"
	"fmt"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RedisReconciler is a generic DependencyReconciler that reconciles
// an internal Redis instance using the Redis options
type RedisReconciler struct {
	*BaseAPIManagerLogicReconciler

	Deployment            func(redis *component.Redis) *k8sappsv1.Deployment
	Service               func(redis *component.Redis) *corev1.Service
	ConfigMap             func(redis *component.RedisConfigMap) *corev1.ConfigMap
	PersistentVolumeClaim func(redis *component.Redis) *corev1.PersistentVolumeClaim
	Secret                func(redis *component.Redis) *corev1.Secret
}

var _ DependencyReconciler = &RedisReconciler{}

func NewSystemRedisDependencyReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &RedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,

		Deployment:            (*component.Redis).SystemDeployment,
		Service:               (*component.Redis).SystemService,
		ConfigMap:             (*component.RedisConfigMap).ConfigMap,
		PersistentVolumeClaim: (*component.Redis).SystemPVC,
		Secret:                (*component.Redis).SystemRedisSecret,
	}
}

func NewBackendRedisDependencyReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &RedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,

		Deployment:            (*component.Redis).BackendDeployment,
		Service:               (*component.Redis).BackendService,
		ConfigMap:             (*component.RedisConfigMap).ConfigMap,
		PersistentVolumeClaim: (*component.Redis).BackendPVC,
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

	deploymentMutator := reconcilers.DeploymentMutator(
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
	)
	redisDeployment := r.Deployment(redis)
	err = r.ReconcileDeployment(redisDeployment, deploymentMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	// Overriding the Deployment health check because the redis PVCs are ReadWriteOnce and so they can't be assigned across multiple nodes (pods)
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(redisDeployment.Name, r.apiManager.GetNamespace(), true, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	// 3scale 2.14 -> 2.15 Upgrade
	// delete NAMESPACE key from secret system-redis
	err = r.deleteSystemRedisSecretNamespaceKey()
	if err != nil {
		return reconcile.Result{}, err
	}

	serviceMutators := []reconcilers.MutateFn{
		reconcilers.CreateOnlyMutator,
		reconcilers.ServiceSelectorMutator,
	}

	// Service
	err = r.ReconcileService(r.Service(redis), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// PVC
	err = r.ReconcilePersistentVolumeClaim(r.PersistentVolumeClaim(redis), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Secret
	err = r.ReconcileSecret(r.Secret(redis), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.updateSystemRedisSecretSSL()
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.updateBackendRedisSecretSSL()
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

func (r *RedisReconciler) deleteSystemRedisSecretNamespaceKey() error {
	secret := &corev1.Secret{}
	err := r.Client().Get(context.TODO(), client.ObjectKey{Namespace: r.apiManager.Namespace, Name: "system-redis"}, secret)
	if k8serr.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	namespaceKey := "NAMESPACE"
	if _, ok := secret.Data[namespaceKey]; ok {
		delete(secret.Data, namespaceKey)
		err = r.UpdateResource(secret)
		if err != nil {
			return err
		}
	}

	return nil
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

func (r *RedisReconciler) updateSystemRedisSecretSSL() error {
	secret := &corev1.Secret{}
	err := r.Client().Get(context.TODO(), client.ObjectKey{
		Namespace: r.apiManager.Namespace, Name: "system-redis"}, secret)
	if k8serr.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if secret.Data == nil { // tmp, added for test that need to be improved?. In runtime not happens
		return nil
	}

	if len(secret.Data[component.SystemSecretSystemRedisCAFile]) > 0 ||
		len(secret.Data[component.SystemSecretSystemRedisClientCertificate]) > 0 ||
		len(secret.Data[component.SystemSecretSystemRedisPrivateKey]) > 0 {
		secret.Data[component.SystemSecretSystemRedisSSL] = []byte("true")
	} else {
		secret.Data[component.SystemSecretSystemRedisSSL] = []byte("false")
	}
	err = r.UpdateResource(secret)
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisReconciler) updateBackendRedisSecretSSL() error {
	secret := &corev1.Secret{}
	err := r.Client().Get(context.TODO(), client.ObjectKey{
		Namespace: r.apiManager.Namespace, Name: "backend-redis"}, secret)
	if k8serr.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if secret.Data == nil { // tmp, added for test that need to be improved?. In runtime not happens
		return nil
	}

	if len(secret.Data[component.BackendSecretBackendRedisConfigCAFile]) > 0 ||
		len(secret.Data[component.BackendSecretBackendRedisConfigClientCertificate]) > 0 ||
		len(secret.Data[component.BackendSecretBackendRedisConfigPrivateKey]) > 0 {
		secret.Data[component.BackendSecretBackendRedisConfigSSL] = []byte("true")
	} else {
		secret.Data[component.BackendSecretBackendRedisConfigSSL] = []byte("false")
	}

	if len(secret.Data[component.BackendSecretBackendRedisConfigQueuesCAFile]) > 0 ||
		len(secret.Data[component.BackendSecretBackendRedisConfigQueuesClientCertificate]) > 0 ||
		len(secret.Data[component.BackendSecretBackendRedisConfigQueuesPrivateKey]) > 0 {
		secret.Data[component.BackendSecretBackendRedisConfigQueuesSSL] = []byte("true")
	} else {
		secret.Data[component.BackendSecretBackendRedisConfigQueuesSSL] = []byte("false")
	}

	err = r.UpdateResource(secret)
	if err != nil {
		return err
	}
	return nil
}
