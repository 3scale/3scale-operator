package operator

import (
	"github.com/3scale/3scale-operator/pkg/upgrade"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

type SystemSearchdReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemSearchdReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *SystemSearchdReconciler {
	return &SystemSearchdReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemSearchdReconciler) Reconcile() (reconcile.Result, error) {
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	searchd, err := SystemSearchd(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	serviceMutators := []reconcilers.MutateFn{
		reconcilers.CreateOnlyMutator,
		reconcilers.ServiceSelectorMutator,
	}

	// Service
	err = r.ReconcileService(searchd.Service(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// PVC
	err = r.ReconcilePersistentVolumeClaim(searchd.PVC(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Deployment
	searchdDeploymentMutator := reconcilers.DeploymentMutator(
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentStrategyMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentProbesMutator,
		reconcilers.DeploymentArgsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
	)
	err = r.ReconcileDeployment(searchd.Deployment(ampImages.Options.SystemSearchdImage), searchdDeploymentMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	// Overriding the Deployment health check because the system-searchd PVC is ReadWriteOnce and so it can't be assigned across multiple nodes (pods)
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.SystemSearchdDeploymentName, r.apiManager.GetNamespace(), true, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func SystemSearchd(cr *appsv1alpha1.APIManager) (*component.SystemSearchd, error) {
	optsProvider := NewSystemSearchdOptionsProvider(cr)
	opts, err := optsProvider.GetOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemSearchd(opts), nil
}
