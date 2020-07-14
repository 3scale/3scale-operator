package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ZyncReconciler struct {
	*BaseAPIManagerLogicReconciler
}

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
	err = r.ReconcileServiceAccount(zync.QueServiceAccount(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que RoleBinding
	err = r.ReconcileRoleBinding(zync.QueRoleBinding(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync DC
	err = r.ReconcileDeploymentConfig(zync.DeploymentConfig(), reconcilers.GenericDeploymentConfigMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que DC
	err = r.ReconcileDeploymentConfig(zync.QueDeploymentConfig(), reconcilers.GenericDeploymentConfigMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync DB DC
	err = r.ReconcileDeploymentConfig(zync.DatabaseDeploymentConfig(), reconcilers.DeploymentConfigResourcesAndAffinityAndTolerationsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Service
	err = r.ReconcileService(zync.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Service
	err = r.ReconcileService(zync.DatabaseService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
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

	err = r.ReconcileGrafanaDashboard(component.ZyncGrafanaDashboard(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(component.ZyncPrometheusRules(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(component.ZyncQuePrometheusRules(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
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
