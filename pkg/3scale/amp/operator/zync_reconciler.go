package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ZyncDatabaseDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewZyncDatabaseDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *ZyncDatabaseDCReconciler {
	return &ZyncDatabaseDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ZyncDatabaseDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type ZyncDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewZyncDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *ZyncDCReconciler {
	return &ZyncDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ZyncDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	tmpUpdate = DeploymentConfigReconcileReplicas(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type ZyncReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &ZyncReconciler{}

func NewZyncReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) ZyncReconciler {
	return ZyncReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ZyncReconciler) Reconcile() (reconcile.Result, error) {
	zync, err := r.zync()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileQueRole(zync.QueRole())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileQueServiceAccount(zync.QueServiceAccount())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileQueRoleBinding(zync.QueRoleBinding())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileZyncDeploymentConfig(zync.DeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileZyncQueDeploymentConfig(zync.QueDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileZyncDatabaseDeploymentConfig(zync.DatabaseDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileZyncService(zync.Service())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileZyncDatabaseService(zync.DatabaseService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileZyncSecret(zync.Secret())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ZyncReconciler) zync() (*component.Zync, error) {
	optsProvider := OperatorZyncOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetZyncOptions()
	if err != nil {
		return nil, err
	}
	return component.NewZync(opts), nil
}

func (r *ZyncReconciler) reconcileQueRole(desiredRole *rbacv1.Role) error {
	reconciler := NewRoleBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyRoleReconciler())
	err := reconciler.Reconcile(desiredRole)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredRole)))
	return nil
}

func (r *ZyncReconciler) reconcileQueServiceAccount(desiredServiceAccount *v1.ServiceAccount) error {
	reconciler := NewServiceAccountBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyServiceAccountReconciler())
	err := reconciler.Reconcile(desiredServiceAccount)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredServiceAccount)))
	return nil
}

func (r *ZyncReconciler) reconcileQueRoleBinding(desiredRoleBinding *rbacv1.RoleBinding) error {
	reconciler := NewRoleBindingBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyRoleBindingReconciler())
	err := reconciler.Reconcile(desiredRoleBinding)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredRoleBinding)))
	return nil
}

func (r *ZyncReconciler) reconcileZyncDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewZyncDCReconciler(r.BaseAPIManagerLogicReconciler))
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *ZyncReconciler) reconcileZyncQueDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	// Zync deployment config reconciler works for ZyncQue
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewZyncDCReconciler(r.BaseAPIManagerLogicReconciler))
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *ZyncReconciler) reconcileZyncDatabaseDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewZyncDatabaseDCReconciler(r.BaseAPIManagerLogicReconciler))
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *ZyncReconciler) reconcileZyncService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	err := reconciler.Reconcile(desiredService)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredService)))
	return nil
}

func (r *ZyncReconciler) reconcileZyncDatabaseService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	err := reconciler.Reconcile(desiredService)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredService)))
	return nil
}

func (r *ZyncReconciler) reconcileZyncSecret(desiredSecret *v1.Secret) error {
	// Secret values are not affected by CR field values
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	err := reconciler.Reconcile(desiredSecret)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredSecret)))
	return nil
}
