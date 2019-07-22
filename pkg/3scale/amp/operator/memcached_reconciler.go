package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

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
	optsProvider := OperatorMemcachedOptionsProvider{APIManagerSpec: &r.apiManager.Spec}
	opts, err := optsProvider.GetMemcachedOptions()
	if err != nil {
		return nil, err
	}
	return component.NewMemcached(opts), nil
}

func (r *MemcachedReconciler) reconcileDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	err := r.InitializeAsAPIManagerObject(desiredDeploymentConfig)
	if err != nil {
		return err
	}

	return r.deploymentConfigReconciler.Reconcile(desiredDeploymentConfig)
}

func (r *MemcachedReconciler) reconcileMemcachedDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}
