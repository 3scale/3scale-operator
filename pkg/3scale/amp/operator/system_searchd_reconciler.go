package operator

import (
	"context"

	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
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

	// 3scale 2.14 -> 2.15 (manticore)
	err = r.supportManticore()
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
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.SystemSearchdDeploymentName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	// 3scale 2.14 -> 2.15 (manticore)
	// Create Manticore re-indexing Job only after the system-searchd Deployment is ready
	searchdDeployment := &k8sappsv1.Deployment{}
	err = r.Client().Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: r.apiManager.GetNamespace(),
		Name:      component.SystemSearchdDeploymentName,
	}, searchdDeployment)
	if err != nil {
		return reconcile.Result{}, err
	}
	if helper.IsDeploymentAvailable(searchdDeployment) && !helper.IsDeploymentProgressing(searchdDeployment) {
		system, err := System(r.apiManager, r.Client())
		if err != nil {
			return reconcile.Result{}, err
		}
		reindexingJob := searchd.ReindexingJob(ampImages.Options.SystemImage, system)
		err = r.ReconcileJob(reindexingJob, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		return reconcile.Result{Requeue: true}, nil
	}

	return reconcile.Result{}, nil
}

func (r *SystemSearchdReconciler) supportManticore() error {
	// The upgrade procedure deletes the old PVC called "system-searchd"; it will be removed when the searchd DC is deleted
	// The normal reconcile loop will create a new PVC called "system-searchd-manticore" for the new searchd Deployment
	oldPVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "system-searchd", Namespace: r.apiManager.Namespace},
	}
	common.TagObjectToDelete(oldPVC)
	err := r.ReconcileResource(&corev1.PersistentVolumeClaim{}, oldPVC, reconcilers.CreateOnlyMutator)
	if err != nil {
		return err
	}

	return nil
}

func SystemSearchd(cr *appsv1alpha1.APIManager) (*component.SystemSearchd, error) {
	optsProvider := NewSystemSearchdOptionsProvider(cr)
	opts, err := optsProvider.GetOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemSearchd(opts), nil
}
