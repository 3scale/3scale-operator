package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	zync, err := Zync(r.apiManager, r.Client())
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

	err = r.reconcilePodDisruptionBudget(zync.ZyncPodDisruptionBudget())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcilePodDisruptionBudget(zync.QuePodDisruptionBudget())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileMonitoringService(component.ZyncMonitoringService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileMonitoringService(component.ZyncQueMonitoringService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileServiceMonitor(component.ZyncServiceMonitor())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileServiceMonitor(component.ZyncQueServiceMonitor())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileGrafanaDashboard(component.ZyncGrafanaDashboard(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileGrafanaDashboard(component.ZyncQueGrafanaDashboard(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcilePrometheusRules(component.ZyncPrometheusRules(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcilePrometheusRules(component.ZyncQuePrometheusRules(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ZyncReconciler) reconcileQueRole(desiredRole *rbacv1.Role) error {
	reconciler := NewRoleBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyRoleReconciler())
	return reconciler.Reconcile(desiredRole)
}

func (r *ZyncReconciler) reconcileQueServiceAccount(desiredServiceAccount *v1.ServiceAccount) error {
	reconciler := NewServiceAccountBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyServiceAccountReconciler())
	return reconciler.Reconcile(desiredServiceAccount)
}

func (r *ZyncReconciler) reconcileQueRoleBinding(desiredRoleBinding *rbacv1.RoleBinding) error {
	reconciler := NewRoleBindingBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyRoleBindingReconciler())
	return reconciler.Reconcile(desiredRoleBinding)
}

func (r *ZyncReconciler) reconcileZyncDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewZyncDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *ZyncReconciler) reconcileZyncQueDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	// Zync deployment config reconciler works for ZyncQue
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewZyncDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *ZyncReconciler) reconcileZyncDatabaseDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewZyncDatabaseDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *ZyncReconciler) reconcileZyncService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *ZyncReconciler) reconcileZyncDatabaseService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *ZyncReconciler) reconcileZyncSecret(desiredSecret *v1.Secret) error {
	// Secret values are not affected by CR field values
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func Zync(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.Zync, error) {
	optsProvider := NewZyncOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetZyncOptions()
	if err != nil {
		return nil, err
	}
	return component.NewZync(opts), nil
}
