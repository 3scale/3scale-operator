package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemPostgreSQLImageReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemPostgreSQLImageReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &SystemPostgreSQLImageReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemPostgreSQLImageReconciler) Reconcile() (reconcile.Result, error) {
	systemPostgreSQLImage, err := SystemPostgreSQLImage(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileImagestream(systemPostgreSQLImage.ImageStream(), reconcilers.GenericImageStreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func SystemPostgreSQLImage(apimanager *appsv1alpha1.APIManager) (*component.SystemPostgreSQLImage, error) {
	optsProvider := NewSystemPostgreSQLImageOptionsProvider(apimanager)
	opts, err := optsProvider.GetSystemPostgreSQLImageOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemPostgreSQLImage(opts), nil
}
