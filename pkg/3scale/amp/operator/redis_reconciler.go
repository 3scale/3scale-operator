package operator

import (
	"github.com/3scale/3scale-operator/pkg/upgrade"
	imagev1 "github.com/openshift/api/image/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

// RedisReconciler is a generic DependencyReconciler that reconciles
// an internal Redis instance using the Redis options
type RedisReconciler struct {
	*BaseAPIManagerLogicReconciler

	Deployment            func(redis *component.Redis) *k8sappsv1.Deployment
	Service               func(redis *component.Redis) *corev1.Service
	ConfigMap             func(redis *component.Redis) *corev1.ConfigMap
	PersistentVolumeClaim func(redis *component.Redis) *corev1.PersistentVolumeClaim
	ImageStream           func(redis *component.Redis) *imagev1.ImageStream
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
		ImageStream:           (*component.Redis).SystemImageStream,
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
		ImageStream:           (*component.Redis).BackendImageStream,
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
