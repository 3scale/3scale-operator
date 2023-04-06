package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/upgrade"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemSphinxReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemSphinxReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *SystemSphinxReconciler {
	return &SystemSphinxReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemSphinxReconciler) Reconcile() (reconcile.Result, error) {
	sphinx, err := SystemSphinx(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Service
	err = r.ReconcileService(sphinx.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	sphinxDC := sphinx.DeploymentConfig()
	dcKey := client.ObjectKey{Name: sphinxDC.Name, Namespace: r.apiManager.GetNamespace()}
	res, err := upgrade.SphinxFromAMPSystemImage(r.Context(), r.Client(), dcKey, r.Logger())
	if err != nil {
		return reconcile.Result{}, err
	}
	if res.Requeue {
		r.Logger().Info("Upgrading Sphinx DC: requeue")
		return res, nil
	}

	// PVC
	err = r.ReconcilePersistentVolumeClaim(sphinx.PVC(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// DC
	sphinxDCmutator := reconcilers.DeploymentConfigMutator(
		reconcilers.DeploymentConfigImageChangeTriggerMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		reconcilers.DeploymentConfigPodTemplateLabelsMutator,
	)
	err = r.ReconcileDeploymentConfig(sphinxDC, sphinxDCmutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func SystemSphinx(cr *appsv1alpha1.APIManager) (*component.SystemSphinx, error) {
	optsProvider := NewSystemSphinxOptionsProvider(cr)
	opts, err := optsProvider.GetOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemSphinx(opts), nil
}
