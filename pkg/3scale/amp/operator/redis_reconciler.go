package operator

import (
	"context"
	"github.com/3scale/3scale-operator/pkg/helper"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

	// 3scale 2.14 -> 2.15 Upgrade
	// delete NAMESPACE attribute from secret system-redis
	// and delete REDIS_NAMESPACE env var from deployments system-app and system-sidekiq
	res, err := r.deleteSystemRedisSecretNamespaceAttribute()
	if err != nil {
		return reconcile.Result{}, err
	}
	if res.Requeue {
		return res, nil
	}

	res, err = r.deleteDeploymentsRedisNamespaceEnvVar()
	if err != nil {
		return reconcile.Result{}, err
	}
	if res.Requeue {
		return res, nil
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

func (r *RedisReconciler) deleteSystemRedisSecretNamespaceAttribute() (reconcile.Result, error) {
	redis, err := Redis(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	desired := redis.SystemRedisSecret()
	existing := &corev1.Secret{}
	err = r.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: r.apiManager.Namespace}, existing)
	if k8serr.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	update := false
	namespaceAttr := "NAMESPACE"
	if _, ok := existing.Data[namespaceAttr]; ok {
		update = true
		delete(existing.Data, namespaceAttr)
	}
	if update {
		err = r.UpdateResource(existing)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *RedisReconciler) deleteDeploymentsRedisNamespaceEnvVar() (reconcile.Result, error) {
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}
	system, err := System(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}
	desired := system.AppDeployment(ampImages.Options.SystemImage)
	res, err := r.deleteDeploymentEnvVars(desired, []string{"REDIS_NAMESPACE"})
	if err != nil {
		return reconcile.Result{}, err
	}
	desired = system.SidekiqDeployment(ampImages.Options.SystemImage)
	res, err = r.deleteDeploymentEnvVars(desired, []string{"REDIS_NAMESPACE"})
	return res, err
}

func (r *RedisReconciler) deleteDeploymentEnvVars(desired *k8sappsv1.Deployment, envVarNames []string) (reconcile.Result, error) {
	existing := &k8sappsv1.Deployment{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: r.apiManager.Namespace}, existing)

	if k8serr.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	update := false

	// containers
	for containerIdx := range existing.Spec.Template.Spec.Containers {
		container := &existing.Spec.Template.Spec.Containers[containerIdx]
		for _, envVarName := range envVarNames {
			if envVarIdx := helper.FindEnvVar(container.Env, envVarName); envVarIdx >= 0 {
				// remove index
				container.Env = append(container.Env[:envVarIdx], container.Env[envVarIdx+1:]...)
				update = true
			}
		}
	}

	// initContainers
	for initContainerIdx := range existing.Spec.Template.Spec.InitContainers {
		initContainer := &existing.Spec.Template.Spec.InitContainers[initContainerIdx]
		for _, envVarName := range envVarNames {
			if envVarIdx := helper.FindEnvVar(initContainer.Env, envVarName); envVarIdx >= 0 {
				// remove index
				initContainer.Env = append(initContainer.Env[:envVarIdx], initContainer.Env[envVarIdx+1:]...)
				update = true
			}
		}
	}

	if update {
		err = r.UpdateResource(existing)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
