package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

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

	r.reconcileQueRole(zync.QueRole())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileQueServiceAccount(zync.QueServiceAccount())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileQueRoleBinding(zync.QueRoleBinding())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileZyncDeploymentConfig(zync.DeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileZyncQueDeploymentConfig(zync.QueDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileZyncDatabaseDeploymentConfig(zync.DatabaseDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileZyncService(zync.Service())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileZyncDatabaseService(zync.DatabaseService())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileZyncSecret(zync.Secret())
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

func (r *ZyncReconciler) reconcileDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	err := r.InitializeAsAPIManagerObject(desiredDeploymentConfig)
	if err != nil {
		return err
	}

	return r.deploymentConfigReconciler.Reconcile(desiredDeploymentConfig)
}

func (r *ZyncReconciler) reconcileSecret(desiredSecret *v1.Secret) error {
	err := r.InitializeAsAPIManagerObject(desiredSecret)
	if err != nil {
		return err
	}
	return r.secretReconciler.Reconcile(desiredSecret)
}

func (r *ZyncReconciler) reconcileService(desiredService *v1.Service) error {
	err := r.InitializeAsAPIManagerObject(desiredService)
	if err != nil {
		return err
	}
	return r.serviceReconciler.Reconcile(desiredService)
}

func (r *ZyncReconciler) reconcileServiceAccount(desiredServiceAccount *v1.ServiceAccount) error {
	err := r.InitializeAsAPIManagerObject(desiredServiceAccount)
	if err != nil {
		return err
	}

	return r.serviceAccountReconciler.Reconcile(desiredServiceAccount)
}

func (r *ZyncReconciler) reconcileRole(desiredRole *rbacv1.Role) error {
	err := r.InitializeAsAPIManagerObject(desiredRole)
	if err != nil {
		return err
	}
	return r.roleReconciler.Reconcile(desiredRole)
}

func (r *ZyncReconciler) reconcileRoleBinding(desiredRoleBinding *rbacv1.RoleBinding) error {
	err := r.InitializeAsAPIManagerObject(desiredRoleBinding)
	if err != nil {
		return err
	}
	return r.roleBindingReconciler.Reconcile(desiredRoleBinding)
}

func (r *ZyncReconciler) reconcileQueRole(desiredRole *rbacv1.Role) error {
	return r.reconcileRole(desiredRole)
}

func (r *ZyncReconciler) reconcileQueServiceAccount(desiredServiceAccount *v1.ServiceAccount) error {
	return r.reconcileServiceAccount(desiredServiceAccount)
}

func (r *ZyncReconciler) reconcileQueRoleBinding(desiredRoleBinding *rbacv1.RoleBinding) error {
	return r.reconcileRoleBinding(desiredRoleBinding)
}

func (r *ZyncReconciler) reconcileZyncDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *ZyncReconciler) reconcileZyncQueDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *ZyncReconciler) reconcileZyncDatabaseDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *ZyncReconciler) reconcileZyncService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *ZyncReconciler) reconcileZyncDatabaseService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *ZyncReconciler) reconcileZyncSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}
