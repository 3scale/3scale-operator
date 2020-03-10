package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemPostgresqlDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewSystemPostgresqlDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *SystemPostgresqlDCReconciler {
	return &SystemPostgresqlDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemPostgresqlDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type SystemPostgreSQLReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &SystemPostgreSQLReconciler{}

func NewSystemPostgreSQLReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) SystemPostgreSQLReconciler {
	return SystemPostgreSQLReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemPostgreSQLReconciler) Reconcile() (reconcile.Result, error) {
	systemPostgreSQL, err := r.systemPostgreSQL()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemPostgreSQLDeploymentConfig(systemPostgreSQL.DeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemPostgreSQLService(systemPostgreSQL.Service())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemPostgreSQLDataPersistentVolumeClaim(systemPostgreSQL.DataPersistentVolumeClaim())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemPostgreSQLSystemDatabaseSecret(systemPostgreSQL.SystemDatabaseSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SystemPostgreSQLReconciler) systemPostgreSQL() (*component.SystemPostgreSQL, error) {
	optsProvider := NewSystemPostgresqlOptionsProvider(r.apiManager, r.apiManager.Namespace, r.Client())
	opts, err := optsProvider.GetSystemPostgreSQLOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemPostgreSQL(opts), nil
}

func (r *SystemPostgreSQLReconciler) reconcileSystemPostgreSQLDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewSystemPostgresqlDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *SystemPostgreSQLReconciler) reconcileSystemPostgreSQLService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *SystemPostgreSQLReconciler) reconcileSystemPostgreSQLSystemDatabaseSecret(desiredSecret *v1.Secret) error {
	// Secret values are not affected by CR field values
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemPostgreSQLReconciler) reconcileSystemPostgreSQLDataPersistentVolumeClaim(desiredPVC *v1.PersistentVolumeClaim) error {
	reconciler := NewPVCBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyPVCReconciler())
	return reconciler.Reconcile(desiredPVC)
}
