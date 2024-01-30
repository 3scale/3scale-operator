package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MemcachedReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewMemcachedReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *MemcachedReconciler {
	return &MemcachedReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *MemcachedReconciler) Reconcile() (reconcile.Result, error) {
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	memcached, err := Memcached(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Deployment
	mutator := reconcilers.DeploymentMutator(
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
	)
	err = r.ReconcileDeployment(memcached.Deployment(ampImages.Options.SystemMemcachedImage), mutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.SystemMemcachedDeploymentName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func Memcached(apimanager *appsv1alpha1.APIManager) (*component.Memcached, error) {
	optsProvider := NewMemcachedOptionsProvider(apimanager)
	opts, err := optsProvider.GetMemcachedOptions()
	if err != nil {
		return nil, err
	}
	return component.NewMemcached(opts), nil
}
