package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MemcachedDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewMemcachedDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *MemcachedDCReconciler {
	return &MemcachedDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *MemcachedDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type MemcachedReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &MemcachedReconciler{}

func NewMemcachedReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) MemcachedReconciler {
	return MemcachedReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *MemcachedReconciler) Reconcile() (reconcile.Result, error) {
	memcached, err := r.memcached()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileMemcachedDeploymentConfig(memcached.DeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *MemcachedReconciler) memcached() (*component.Memcached, error) {
	optsProvider := NewMemcachedOptionsProvider(r.apiManager)
	opts, err := optsProvider.GetMemcachedOptions()
	if err != nil {
		return nil, err
	}
	return component.NewMemcached(opts), nil
}

func (r *MemcachedReconciler) reconcileMemcachedDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewMemcachedDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}
