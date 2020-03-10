package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	imagev1 "github.com/openshift/api/image/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemPostgreSQLImageReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &SystemPostgreSQLImageReconciler{}

func NewSystemPostgreSQLImageReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) SystemPostgreSQLImageReconciler {
	return SystemPostgreSQLImageReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemPostgreSQLImageReconciler) Reconcile() (reconcile.Result, error) {
	systemPostgreSQLImage, err := r.systemPostgreSQLImage()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemPostgreSQLImageStream(systemPostgreSQLImage.ImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SystemPostgreSQLImageReconciler) systemPostgreSQLImage() (*component.SystemPostgreSQLImage, error) {
	optsProvider := NewSystemPostgreSQLImageOptionsProvider(r.apiManager)
	opts, err := optsProvider.GetSystemPostgreSQLImageOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemPostgreSQLImage(opts), nil
}

func (r *SystemPostgreSQLImageReconciler) reconcileSystemPostgreSQLImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}
