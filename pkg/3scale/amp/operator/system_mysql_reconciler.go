package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemMySQLReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemMySQLReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *SystemMySQLReconciler {
	return &SystemMySQLReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemMySQLReconciler) Reconcile() (reconcile.Result, error) {
	systemMySQL, err := SystemMySQL(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// DC
	err = r.ReconcileDeploymentConfig(systemMySQL.DeploymentConfig(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Service
	err = r.ReconcileService(systemMySQL.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Main CM
	err = r.ReconcileConfigMap(systemMySQL.MainConfigConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Extra CM
	err = r.ReconcileConfigMap(systemMySQL.ExtraConfigConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// PCV
	err = r.ReconcilePersistentVolumeClaim(systemMySQL.PersistentVolumeClaim(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Secret
	err = r.ReconcileSecret(systemMySQL.SystemDatabaseSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func SystemMySQL(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.SystemMysql, error) {
	optsProvider := NewSystemMysqlOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetMysqlOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemMysql(opts), nil
}
