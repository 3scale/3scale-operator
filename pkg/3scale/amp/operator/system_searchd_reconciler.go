package operator

import (
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
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
	searchd, err := SystemSearchd(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.13 -> 2.14
	err = r.upgradeFromSphinx()
	if err != nil {
		return reconcile.Result{}, err
	}

	// Service
	err = r.ReconcileService(searchd.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// PVC
	err = r.ReconcilePersistentVolumeClaim(searchd.PVC(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// DC
	searchdDCmutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigImageChangeTriggerMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		reconcilers.DeploymentConfigPodTemplateLabelsMutator,
	)
	err = r.ReconcileDeploymentConfig(searchd.DeploymentConfig(), searchdDCmutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SystemSearchdReconciler) upgradeFromSphinx() error {
	// The upgrade procedure is simply based on:
	// * Delete "old" DC called system-sphinx (if found)
	// * The regular reconciling logic will create a new DC called system-searchd

	oldService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "system-sphinx", Namespace: r.apiManager.Namespace},
	}
	common.TagObjectToDelete(oldService)
	err := r.ReconcileResource(&v1.Service{}, oldService, reconcilers.CreateOnlyMutator)
	if err != nil {
		return err
	}

	oldDC := &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "system-sphinx", Namespace: r.apiManager.Namespace},
	}
	common.TagObjectToDelete(oldDC)
	err = r.ReconcileResource(&appsv1.DeploymentConfig{}, oldDC, reconcilers.CreateOnlyMutator)
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
