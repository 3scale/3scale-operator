package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func HighAvailability(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.HighAvailability, error) {
	optsProvider := NewHighAvailabilityOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetHighAvailabilityOptions()
	if err != nil {
		return nil, err
	}
	return component.NewHighAvailability(opts), nil
}

func NewBackendExternalRedisReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &ExternalDependencyReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		GetSecret:                     (*component.HighAvailability).BackendRedisSecret,
	}
}

func NewSystemExternalRedisReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &ExternalDependencyReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		GetSecret:                     (*component.HighAvailability).SystemRedisSecret,
	}
}

func NewSystemExternalDatabaseReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &ExternalDependencyReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		GetSecret:                     (*component.HighAvailability).SystemDatabaseSecret,
	}
}

// ExternalDependencyReconciler is a DependencyReconciler that uses the
// HighAvailability options to reconcile an external dependency
type ExternalDependencyReconciler struct {
	*BaseAPIManagerLogicReconciler

	GetSecret func(*component.HighAvailability) *corev1.Secret
}

var _ DependencyReconciler = &ExternalDependencyReconciler{}

func (r *ExternalDependencyReconciler) Reconcile() (reconcile.Result, error) {
	ha, err := HighAvailability(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileSecret(r.GetSecret(ha), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
