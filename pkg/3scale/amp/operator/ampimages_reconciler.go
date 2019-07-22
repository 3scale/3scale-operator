package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
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
	ampImages, err := r.ampImages()
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

	err = r.reconcileBackendRedisImageStream(ampImages.BackendRedisImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemRedisImageStream(ampImages.SystemRedisImageStream())
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

// TODO should this be performed in another place
func (r *AMPImagesReconciler) ampImages() (*component.AmpImages, error) {
	optsProvider := OperatorAmpImagesOptionsProvider{APIManagerSpec: &r.apiManager.Spec}
	opts, err := optsProvider.GetAmpImagesOptions()
	if err != nil {
		return nil, err
	}
	return component.NewAmpImages(opts), nil
}

func (r *AMPImagesReconciler) reconcileImageStream(desiredImageStream *imagev1.ImageStream) error {
	err := r.InitializeAsAPIManagerObject(desiredImageStream)
	if err != nil {
		return err
	}

	return r.imagestreamReconciler.Reconcile(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileServiceAccount(desiredServiceAccount *v1.ServiceAccount) error {
	err := r.InitializeAsAPIManagerObject(desiredServiceAccount)
	if err != nil {
		return err
	}

	return r.serviceAccountReconciler.Reconcile(desiredServiceAccount)
}

func (r *AMPImagesReconciler) reconcileBackendImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileZyncImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileApicastImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileSystemImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileZyncDatabasePostgreSQLImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileBackendRedisImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileSystemRedisImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileSystemMemcachedImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileDeploymentsServiceAccount(desiredServiceAccount *v1.ServiceAccount) error {
	return r.reconcileServiceAccount(desiredServiceAccount)
}
