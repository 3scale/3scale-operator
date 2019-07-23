package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemMySQLReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &SystemMySQLReconciler{}

func NewSystemMySQLReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) SystemMySQLReconciler {
	return SystemMySQLReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemMySQLReconciler) Reconcile() (reconcile.Result, error) {
	if r.apiManager.Spec.HighAvailability != nil && r.apiManager.Spec.HighAvailability.Enabled {
		return reconcile.Result{}, nil
	}

	systemMySQL, err := r.systemMySQL()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemMySQLDeploymentConfig(systemMySQL.DeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemMySQLService(systemMySQL.Service())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemMySQLMainConfigMap(systemMySQL.MainConfigConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemMySQLExtraConfigMap(systemMySQL.ExtraConfigConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemMySQLPersistentVolumeClaim(systemMySQL.PersistentVolumeClaim())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemMySQLSystemDatabaseSecret(systemMySQL.SystemDatabaseSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SystemMySQLReconciler) systemMySQL() (*component.SystemMysql, error) {
	optsProvider := OperatorMysqlOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetMysqlOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemMysql(opts), nil
}

func (r *SystemMySQLReconciler) reconcileDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	err := r.InitializeAsAPIManagerObject(desiredDeploymentConfig)
	if err != nil {
		return err
	}

	return r.deploymentConfigReconciler.Reconcile(desiredDeploymentConfig)
}

func (r *SystemMySQLReconciler) reconcileService(desiredService *v1.Service) error {
	err := r.InitializeAsAPIManagerObject(desiredService)
	if err != nil {
		return err
	}

	return r.serviceReconciler.Reconcile(desiredService)
}

func (r *SystemMySQLReconciler) reconcileSecret(desiredSecret *v1.Secret) error {
	err := r.InitializeAsAPIManagerObject(desiredSecret)
	if err != nil {
		return err
	}

	return r.secretReconciler.Reconcile(desiredSecret)
}

func (r *SystemMySQLReconciler) reconcilePersistentVolumeClaim(desiredPVC *v1.PersistentVolumeClaim) error {
	err := r.InitializeAsAPIManagerObject(desiredPVC)
	if err != nil {
		return err
	}

	return r.persistentVolumeClaimReconciler.Reconcile(desiredPVC)
}

func (r *SystemMySQLReconciler) reconcileConfigMap(desiredConfigMap *v1.ConfigMap) error {
	err := r.InitializeAsAPIManagerObject(desiredConfigMap)
	if err != nil {
		return err
	}

	return r.configMapReconciler.Reconcile(desiredConfigMap)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLMainConfigMap(desiredConfigMap *v1.ConfigMap) error {
	return r.reconcileConfigMap(desiredConfigMap)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLExtraConfigMap(desiredConfigMap *v1.ConfigMap) error {
	return r.reconcileConfigMap(desiredConfigMap)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLSystemDatabaseSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLPersistentVolumeClaim(desiredPVC *v1.PersistentVolumeClaim) error {
	return r.reconcilePersistentVolumeClaim(desiredPVC)
}
