package operator

import (
	appsv1beta1 "github.com/3scale/3scale-operator/apis/apps/v1beta1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MemcachedReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewMemcachedReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *MemcachedReconciler {
	return &MemcachedReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *MemcachedReconciler) Reconcile() (reconcile.Result, error) {
	memcached, err := Memcached(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// DC
	mutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
	)
	err = r.ReconcileDeploymentConfig(memcached.DeploymentConfig(), mutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func Memcached(apimanager *appsv1beta1.APIManager) (*component.Memcached, error) {
	optsProvider := NewMemcachedOptionsProvider(apimanager)
	opts, err := optsProvider.GetMemcachedOptions()
	if err != nil {
		return nil, err
	}
	return component.NewMemcached(opts), nil
}
