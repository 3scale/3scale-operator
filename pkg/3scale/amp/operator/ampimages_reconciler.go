package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AMPImagesReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewAMPImagesReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *AMPImagesReconciler {
	return &AMPImagesReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *AMPImagesReconciler) Reconcile() (reconcile.Result, error) {
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// backend IS
	err = r.ReconcileImagestream(ampImages.BackendImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// zync IS
	err = r.ReconcileImagestream(ampImages.ZyncImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// apicast IS
	err = r.ReconcileImagestream(ampImages.APICastImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// system IS
	err = r.ReconcileImagestream(ampImages.SystemImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// zync db postresql IS
	err = r.ReconcileImagestream(ampImages.ZyncDatabasePostgreSQLImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// system memcached IS
	err = r.ReconcileImagestream(ampImages.SystemMemcachedImageStream(), reconcilers.GenericImagestreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileServiceAccount(ampImages.DeploymentsServiceAccount(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func AmpImages(apimanager *appsv1alpha1.APIManager) (*component.AmpImages, error) {
	optsProvider := NewAmpImagesOptionsProvider(apimanager)
	opts, err := optsProvider.GetAmpImagesOptions()
	if err != nil {
		return nil, err
	}
	return component.NewAmpImages(opts), nil
}
