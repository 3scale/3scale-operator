package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type HighAvailabilityReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewHighAvailabilityReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *HighAvailabilityReconciler {
	return &HighAvailabilityReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *HighAvailabilityReconciler) Reconcile() (reconcile.Result, error) {
	ha, err := HighAvailability(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// Backend Redis Secret
	err = r.ReconcileSecret(ha.BackendRedisSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System Redis Secret
	err = r.ReconcileSecret(ha.SystemRedisSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System database Secret
	err = r.ReconcileSecret(ha.SystemDatabaseSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func HighAvailability(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.HighAvailability, error) {
	optsProvider := NewHighAvailabilityOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetHighAvailabilityOptions()
	if err != nil {
		return nil, err
	}
	return component.NewHighAvailability(opts), nil
}
