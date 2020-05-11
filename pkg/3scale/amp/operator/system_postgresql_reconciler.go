package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemPostgreSQLReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemPostgreSQLReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *SystemPostgreSQLReconciler {
	return &SystemPostgreSQLReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemPostgreSQLReconciler) Reconcile() (reconcile.Result, error) {
	systemPostgreSQL, err := SystemPostgreSQL(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// DC
	err = r.ReconcileDeploymentConfig(systemPostgreSQL.DeploymentConfig(), reconcilers.DeploymentConfigResourcesMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Service
	err = r.ReconcileService(systemPostgreSQL.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// PVC
	err = r.ReconcilePersistentVolumeClaim(systemPostgreSQL.DataPersistentVolumeClaim(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// DB secret
	err = r.ReconcileSecret(systemPostgreSQL.SystemDatabaseSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func SystemPostgreSQL(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.SystemPostgreSQL, error) {
	optsProvider := NewSystemPostgresqlOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetSystemPostgreSQLOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemPostgreSQL(opts), nil
}
