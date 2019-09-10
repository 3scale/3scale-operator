package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	imagev1 "github.com/openshift/api/image/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemMySQLImageReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &SystemMySQLImageReconciler{}

func NewSystemMySQLImageReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) SystemMySQLImageReconciler {
	return SystemMySQLImageReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemMySQLImageReconciler) Reconcile() (reconcile.Result, error) {
	if r.apiManager.Spec.HighAvailability != nil && r.apiManager.Spec.HighAvailability.Enabled {
		return reconcile.Result{}, nil
	}

	systemMySQLImage, err := r.systemMySQLImage()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemMySQLImageStream(systemMySQLImage.ImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SystemMySQLImageReconciler) systemMySQLImage() (*component.SystemMySQLImage, error) {
	optsProvider := OperatorSystemMySQLImageOptionsProvider{APIManagerSpec: &r.apiManager.Spec}
	opts, err := optsProvider.GetSystemMySQLImageOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemMySQLImage(opts), nil
}

func (r *SystemMySQLImageReconciler) reconcileSystemMySQLImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	err := reconciler.Reconcile(desiredImageStream)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredImageStream)))
	return nil
}
