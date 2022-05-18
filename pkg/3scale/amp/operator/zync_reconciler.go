package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ZyncReconciler struct {
	*BaseAPIManagerLogicReconciler
}

const (
	disableZyncQueInstancesSyncing = "apps.3scale.net/zync-que-replica-field"
	disableZyncInstancesSyncing    = "apps.3scale.net/zync-replica-field"
)

func NewZyncReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *ZyncReconciler {
	return &ZyncReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ZyncReconciler) Reconcile() (reconcile.Result, error) {
	zync, err := Zync(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que Role
	err = r.ReconcileRole(zync.QueRole(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que SA
	err = r.ReconcileServiceAccount(zync.QueServiceAccount(), reconcilers.ServiceAccountImagePullPolicyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que RoleBinding
	err = r.ReconcileRoleBinding(zync.QueRoleBinding(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync DC
	zyncMutator := reconcilers.GetConfigMutators(r.apiManager.Annotations, disableZyncInstancesSyncing)
	err = r.ReconcileDeploymentConfig(zync.DeploymentConfig(), reconcilers.DeploymentConfigMutator(zyncMutator...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que DC
	zyncQueMutator := reconcilers.GetConfigMutators(r.apiManager.Annotations, disableZyncQueInstancesSyncing)
	err = r.ReconcileDeploymentConfig(zync.QueDeploymentConfig(), reconcilers.DeploymentConfigMutator(zyncQueMutator...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Service
	err = r.ReconcileService(zync.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !r.apiManager.IsExternal(appsv1alpha1.ZyncDatabase) {
		// Zync DB DC
		err = r.ReconcileDeploymentConfig(zync.DatabaseDeploymentConfig(), reconcilers.DeploymentConfigMutator(reconcilers.GenericOpts()...))
		if err != nil {
			return reconcile.Result{}, err
		}

		// Zync DB Service
		err = r.ReconcileService(zync.DatabaseService(), reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Zync Secret
	err = r.ReconcileSecret(zync.Secret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync PDB
	err = r.ReconcilePodDisruptionBudget(zync.ZyncPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que PDB
	err = r.ReconcilePodDisruptionBudget(zync.QuePodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(zync.ZyncPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(zync.ZyncQuePodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	sumRate, err := helper.SumRateForOpenshiftVersion(r.Context(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(zync.ZyncGrafanaDashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(zync.ZyncPrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(zync.ZyncQuePrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func Zync(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.Zync, error) {
	optsProvider := NewZyncOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetZyncOptions()
	if err != nil {
		return nil, err
	}
	return component.NewZync(opts), nil
}
