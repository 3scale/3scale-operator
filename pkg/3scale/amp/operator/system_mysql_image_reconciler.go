package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemMySQLImageReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemMySQLImageReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *SystemMySQLImageReconciler {
	return &SystemMySQLImageReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemMySQLImageReconciler) Reconcile() (reconcile.Result, error) {
	systemMySQLImage, err := SystemMySQLImage(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileImagestream(systemMySQLImage.ImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func SystemMySQLImage(apimanager *appsv1alpha1.APIManager) (*component.SystemMySQLImage, error) {
	optsProvider := NewSystemMysqlImageOptionsProvider(apimanager)
	opts, err := optsProvider.GetSystemMySQLImageOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemMySQLImage(opts), nil
}
