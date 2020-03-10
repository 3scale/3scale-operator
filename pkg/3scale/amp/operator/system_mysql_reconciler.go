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
	optsProvider := NewSystemMysqlOptionsProvider(r.apiManager, r.apiManager.Namespace, r.Client())
	opts, err := optsProvider.GetMysqlOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemMysql(opts), nil
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyDCReconciler())
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLMainConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	return reconciler.Reconcile(desiredConfigMap)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLExtraConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	return reconciler.Reconcile(desiredConfigMap)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLSystemDatabaseSecret(desiredSecret *v1.Secret) error {
	// Secret values are not affected by CR field values
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLPersistentVolumeClaim(desiredPVC *v1.PersistentVolumeClaim) error {
	reconciler := NewPVCBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyPVCReconciler())
	return reconciler.Reconcile(desiredPVC)
}
