package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

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

	r.reconcileSystemPostgreSQLDeploymentConfig(systemPostgreSQL.DeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileSystemPostgreSQLService(systemPostgreSQL.Service())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileSystemPostgreSQLDataPersistentVolumeClaim(systemPostgreSQL.DataPersistentVolumeClaim())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileSystemPostgreSQLSystemDatabaseSecret(systemPostgreSQL.SystemDatabaseSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SystemPostgreSQLReconciler) systemPostgreSQL() (*component.SystemPostgreSQL, error) {
	optsProvider := OperatorSystemPostgreSQLOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetSystemPostgreSQLOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemPostgreSQL(opts), nil
}

func (r *SystemPostgreSQLReconciler) reconcileDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	err := r.InitializeAsAPIManagerObject(desiredDeploymentConfig)
	if err != nil {
		return err
	}

	return r.deploymentConfigReconciler.Reconcile(desiredDeploymentConfig)
}

func (r *SystemPostgreSQLReconciler) reconcileService(desiredService *v1.Service) error {
	err := r.InitializeAsAPIManagerObject(desiredService)
	if err != nil {
		return err
	}

	return r.serviceReconciler.Reconcile(desiredService)
}

func (r *SystemPostgreSQLReconciler) reconcileSecret(desiredSecret *v1.Secret) error {
	err := r.InitializeAsAPIManagerObject(desiredSecret)
	if err != nil {
		return err
	}

	return r.secretReconciler.Reconcile(desiredSecret)
}

func (r *SystemPostgreSQLReconciler) reconcilePersistentVolumeClaim(desiredPVC *v1.PersistentVolumeClaim) error {
	err := r.InitializeAsAPIManagerObject(desiredPVC)
	if err != nil {
		return err
	}

	return r.persistentVolumeClaimReconciler.Reconcile(desiredPVC)
}

func (r *SystemPostgreSQLReconciler) reconcileSystemPostgreSQLDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *SystemPostgreSQLReconciler) reconcileSystemPostgreSQLService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *SystemPostgreSQLReconciler) reconcileSystemPostgreSQLSystemDatabaseSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemPostgreSQLReconciler) reconcileSystemPostgreSQLDataPersistentVolumeClaim(desiredPVC *v1.PersistentVolumeClaim) error {
	return r.reconcilePersistentVolumeClaim(desiredPVC)
}
