package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	imagev1 "github.com/openshift/api/image/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "k8s.io/api/core/v1"
)

type AMPImagesReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &AMPImagesReconciler{}

func NewAMPImagesReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) AMPImagesReconciler {
	return AMPImagesReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *AMPImagesReconciler) Reconcile() (reconcile.Result, error) {
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileBackendImageStream(ampImages.BackendImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileZyncImageStream(ampImages.ZyncImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileApicastImageStream(ampImages.APICastImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemImageStream(ampImages.SystemImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileZyncDatabasePostgreSQLImageStream(ampImages.ZyncDatabasePostgreSQLImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemMemcachedImageStream(ampImages.SystemMemcachedImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileDeploymentsServiceAccount(ampImages.DeploymentsServiceAccount())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *AMPImagesReconciler) reconcileBackendImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileZyncImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileApicastImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileSystemImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileZyncDatabasePostgreSQLImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileSystemMemcachedImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileDeploymentsServiceAccount(desiredServiceAccount *v1.ServiceAccount) error {
	reconciler := NewServiceAccountBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyServiceAccountReconciler())
	return reconciler.Reconcile(desiredServiceAccount)
}

func AmpImages(apimanager *appsv1alpha1.APIManager) (*component.AmpImages, error) {
	optsProvider := NewAmpImagesOptionsProvider(apimanager)
	opts, err := optsProvider.GetAmpImagesOptions()
	if err != nil {
		return nil, err
	}
	return component.NewAmpImages(opts), nil
}
