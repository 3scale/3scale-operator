package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CommonEmbeddedRedisReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewCommonEmbeddedRedisReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *CommonEmbeddedRedisReconciler {
	return &CommonEmbeddedRedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *CommonEmbeddedRedisReconciler) Reconcile() (reconcile.Result, error) {
	redis, err := CommonEmbeddedRedis(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// backend CM
	err = r.ReconcileConfigMap(redis.ConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func CommonEmbeddedRedis(apimanager *appsv1alpha1.APIManager) (*component.CommonEmbeddedRedis, error) {
	optsProvider := NewCommonEmbeddedRedisOptionProvider(apimanager)
	opts, err := optsProvider.GetCommonEmbeddedRedisOptions()
	if err != nil {
		return nil, err
	}
	return component.NewCommonEmbeddedRedis(opts), nil
}
