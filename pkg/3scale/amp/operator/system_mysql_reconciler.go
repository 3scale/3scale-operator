package operator

import (
	"fmt"

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

func (r *SystemMySQLReconciler) reconcileSystemMySQLDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyDCReconciler())
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	err := reconciler.Reconcile(desiredService)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredService)))
	return nil
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLMainConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	err := reconciler.Reconcile(desiredConfigMap)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredConfigMap)))
	return nil
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLExtraConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	err := reconciler.Reconcile(desiredConfigMap)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredConfigMap)))
	return nil
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLSystemDatabaseSecret(desiredSecret *v1.Secret) error {
	// Secret values are not affected by CR field values
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	err := reconciler.Reconcile(desiredSecret)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredSecret)))
	return nil
}

func (r *SystemMySQLReconciler) reconcileSystemMySQLPersistentVolumeClaim(desiredPVC *v1.PersistentVolumeClaim) error {
	reconciler := NewPVCBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyPVCReconciler())
	err := reconciler.Reconcile(desiredPVC)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredPVC)))
	return nil
}
