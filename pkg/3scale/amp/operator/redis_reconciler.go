package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
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

func (r *RedisReconciler) Reconcile() (reconcile.Result, error) {
	redis, err := Redis(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backend redis DC
	err = r.ReconcileDeploymentConfig(redis.BackendDeploymentConfig(), reconcilers.DeploymentConfigResourcesMutator)
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
	err = r.ReconcileImagestream(redis.BackendImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis DC
	err = r.ReconcileDeploymentConfig(redis.SystemDeploymentConfig(), reconcilers.DeploymentConfigResourcesMutator)
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
	err = r.ReconcileImagestream(redis.SystemImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func Redis(apimanager *appsv1alpha1.APIManager) (*component.Redis, error) {
	optsProvider := NewRedisOptionsProvider(apimanager)
	opts, err := optsProvider.GetRedisOptions()
	if err != nil {
		return nil, err
	}
	return component.NewRedis(opts), nil
}
