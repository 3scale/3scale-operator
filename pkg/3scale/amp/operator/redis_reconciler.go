package operator

import (
	"context"
	"fmt"
	"github.com/3scale/3scale-operator/pkg/common"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
)

// RedisReconciler is a generic DependencyReconciler that reconciles
// an internal Redis instance using the Redis options
type RedisReconciler struct {
	*BaseAPIManagerLogicReconciler

	Deployment            func(redis *component.Redis) *k8sappsv1.Deployment
	Service               func(redis *component.Redis) *corev1.Service
	ConfigMap             func(redis *component.Redis) *corev1.ConfigMap
	PersistentVolumeClaim func(redis *component.Redis) *corev1.PersistentVolumeClaim
	Secret                func(redis *component.Redis) *corev1.Secret
}

var _ DependencyReconciler = &RedisReconciler{}

func NewSystemRedisDependencyReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &RedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,

		Deployment:            (*component.Redis).SystemDeployment,
		Service:               (*component.Redis).SystemService,
		ConfigMap:             (*component.Redis).ConfigMap,
		PersistentVolumeClaim: (*component.Redis).SystemPVC,
		Secret:                (*component.Redis).SystemRedisSecret,
	}
}

func NewBackendRedisDependencyReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &RedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,

		Deployment:            (*component.Redis).BackendDeployment,
		Service:               (*component.Redis).BackendService,
		ConfigMap:             (*component.Redis).ConfigMap,
		PersistentVolumeClaim: (*component.Redis).BackendPVC,
		Secret:                (*component.Redis).BackendRedisSecret,
	}
}

func (r *RedisReconciler) Reconcile() (reconcile.Result, error) {
	redis, err := Redis(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// We want to reconcile redis-conf ConfigMap before Deployment
	// to avoid restart redis pods twice in case of User change ConfigMap.
	// Annotation GenerationID was added to Pod Template to support Upgrade scenario.
	// this GenerationID is taken from ConfigMap resourceVersion.
	// If User change Config Map, Operator will revert it to original one,
	// but resourceVersion could be changed twice (after user change and after operator).
	// To avoid this scenario by placing CM reconciliation before Deployment
	if r.ConfigMap != nil {
		redisConfigMap := r.ConfigMap(redis)
		isInternalRedis, err := r.isInternalRedis()
		if err != nil {
			return reconcile.Result{}, err
		}
		if isInternalRedis {
			redis_configuration := r.ConfigMap(redis).Data["redis.conf"]
			redisConfigMap.Data["redis.conf"] = redis_configuration + "\n" + "rename-command REPLICAOF \"\"" + "\n" + "rename-command SLAVEOF \"\"" + "\n"
		}
		err = r.ReconcileConfigMap(redisConfigMap, r.redisConfigCmMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
		if isInternalRedis {
			r.RollOutRedisPods(context.TODO())
		}
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

func (r *RedisReconciler) redisConfigCmMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*corev1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", existingObj)
	}
	desired, ok := desiredObj.(*corev1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", desiredObj)
	}

	update := false
	fieldUpdated := reconcilers.ConfigMapReconcileField(desired, existing, "redis.conf")
	update = update || fieldUpdated

	return update, nil
}

func (r *RedisReconciler) isInternalRedis() (bool, error) {
	if r.apiManager.Spec.ExternalComponents != nil &&
		r.apiManager.Spec.ExternalComponents.Backend != nil &&
		r.apiManager.Spec.ExternalComponents.Backend.Redis != nil &&
		*r.apiManager.Spec.ExternalComponents.Backend.Redis == true {
		return false, nil
	}
	if r.apiManager.Spec.ExternalComponents != nil &&
		r.apiManager.Spec.ExternalComponents.System != nil &&
		r.apiManager.Spec.ExternalComponents.System.Redis != nil &&
		*r.apiManager.Spec.ExternalComponents.System.Redis == true {
		return false, nil
	}
	return true, nil
}

func (r *RedisReconciler) RollOutRedisPods(ctx context.Context) error {

	podLabelSelector := labels.NewSelector()
	labelSelector, err := labels.NewRequirement("deployment", selection.In, []string{"backend-redis", "system-redis"})
	if err != nil {
		return fmt.Errorf("podlabelSelector failes: %w", err)
	}
	podLabelSelector = podLabelSelector.Add(*labelSelector)

	redisPods := &corev1.PodList{}
	opts := client.ListOptions{
		Namespace:     r.apiManager.GetNamespace(),
		LabelSelector: podLabelSelector,
	}
	err = r.Client().List(ctx, redisPods, &opts)
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	if len(redisPods.Items) < 2 {
		return fmt.Errorf("redis pods not found: %w", err)
	}
	for _, pod := range redisPods.Items {
		err = r.Client().Update(context.Background(), &pod)
		if err != nil {
			return fmt.Errorf("error update redis pod: %w", err)
		}
	}

	return nil
}
