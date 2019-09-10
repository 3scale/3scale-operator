package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BackendWorkerDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewBackendWorkerDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *BackendWorkerDCReconciler {
	return &BackendWorkerDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *BackendWorkerDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileReplicas(desired, existing, r.Logger())
	update = update || tmpUpdate

	tmpUpdate = DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type BackendListenerDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewBackendListenerDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *BackendListenerDCReconciler {
	return &BackendListenerDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *BackendListenerDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileReplicas(desired, existing, r.Logger())
	update = update || tmpUpdate

	tmpUpdate = DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type BackendCronDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewBackendCronDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *BackendCronDCReconciler {
	return &BackendCronDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *BackendCronDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileReplicas(desired, existing, r.Logger())
	update = update || tmpUpdate

	tmpUpdate = DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type BackendReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &BackendReconciler{}

func NewBackendReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) BackendReconciler {
	return BackendReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *BackendReconciler) Reconcile() (reconcile.Result, error) {
	backend, err := r.backend()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileCronDeploymentConfig(backend.CronDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileListenerDeploymentConfig(backend.ListenerDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileListenerService(backend.ListenerService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileListenerRoute(backend.ListenerRoute())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileWorkerDeploymentConfig(backend.WorkerDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileEnvironmentConfigMap(backend.EnvironmentConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileInternalAPISecret(backend.InternalAPISecretForSystem())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileRedisSecret(backend.RedisSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileListenerSecret(backend.ListenerSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *BackendReconciler) backend() (*component.Backend, error) {
	optsProvider := OperatorBackendOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetBackendOptions()
	if err != nil {
		return nil, err
	}
	return component.NewBackend(opts), nil
}

func (r *BackendReconciler) reconcileCronDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewBackendCronDCReconciler(r.BaseAPIManagerLogicReconciler))
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *BackendReconciler) reconcileListenerDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewBackendListenerDCReconciler(r.BaseAPIManagerLogicReconciler))
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *BackendReconciler) reconcileListenerService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	err := reconciler.Reconcile(desiredService)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredService)))
	return nil
}

func (r *BackendReconciler) reconcileListenerRoute(desiredRoute *routev1.Route) error {
	reconciler := NewRouteBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyRouteReconciler())
	err := reconciler.Reconcile(desiredRoute)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredRoute)))
	return nil
}

func (r *BackendReconciler) reconcileWorkerDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewBackendWorkerDCReconciler(r.BaseAPIManagerLogicReconciler))
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *BackendReconciler) reconcileEnvironmentConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	err := reconciler.Reconcile(desiredConfigMap)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredConfigMap)))
	return nil
}

func (r *BackendReconciler) reconcileInternalAPISecret(desiredSecret *v1.Secret) error {
	// Secret values are not affected by CR field values
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	err := reconciler.Reconcile(desiredSecret)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredSecret)))
	return nil
}

func (r *BackendReconciler) reconcileRedisSecret(desiredSecret *v1.Secret) error {
	// Secret values are not affected by CR field values
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	err := reconciler.Reconcile(desiredSecret)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredSecret)))
	return nil
}

func (r *BackendReconciler) reconcileListenerSecret(desiredSecret *v1.Secret) error {
	// Secret values are not affected by CR field values
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	err := reconciler.Reconcile(desiredSecret)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredSecret)))
	return nil
}
