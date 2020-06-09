package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BackendRedisReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewBackendRedisReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *BackendRedisReconciler {
	return &BackendRedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *BackendRedisReconciler) Reconcile() (reconcile.Result, error) {
	redis, err := BackendRedis(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backend redis DC
	err = r.ReconcileDeploymentConfig(redis.DeploymentConfig(), reconcilers.DeploymentConfigResourcesMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// backend redis Service
	err = r.ReconcileService(redis.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backenb PVC
	err = r.ReconcilePersistentVolumeClaim(redis.PersistentVolumeClaim(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backend IS
	err = r.ReconcileImagestream(redis.ImageStream(), reconcilers.GenericImageStreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func BackendRedis(apimanager *appsv1alpha1.APIManager) (*component.BackendRedis, error) {
	optsProvider := NewBackendRedisOptionsProvider(apimanager)
	opts, err := optsProvider.GetBackendRedisOptions()
	if err != nil {
		return nil, err
	}
	return component.NewBackendRedis(opts), nil
}
