package operator

import (
	"context"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
		reconcilers.DeploymentPodInitContainerMutator,
		systemDatabaseTLSEnvVarMutator,
		reconcilers.DeploymentVolumesMutator,
		reconcilers.DeploymentInitContainerVolumeMountsMutator,
		reconcilers.DeploymentContainerVolumeMountsMutator,
	)

	searchdDep, err := searchd.Deployment(r.Context(), r.Client(), r.apiManager.Namespace, ampImages.Options.SystemSearchdImage)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileDeployment(searchdDep, searchdDeploymentMutator)
	if err != nil {
		return reconcile.Result{}, err
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

func SystemSearchd(cr *appsv1alpha1.APIManager) (*component.SystemSearchd, error) {
	optsProvider := NewSystemSearchdOptionsProvider(cr)
	opts, err := optsProvider.GetOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemSearchd(opts), nil
}

func systemDatabaseTLSEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for TLS
	var changed bool

	for _, envVar := range []string{
		"DATABASE_SSL_CA",
		"DATABASE_SSL_CERT",
		"DATABASE_SSL_KEY",
		"DATABASE_SSL_MODE",
		"DB_SSL_CA",
		"DB_SSL_CERT",
		"DB_SSL_KEY",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}
