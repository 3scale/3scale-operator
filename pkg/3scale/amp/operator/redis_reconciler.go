package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type RedisSystemDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewRedisSystemDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *RedisSystemDCReconciler {
	return &RedisSystemDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *RedisSystemDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type RedisBackendDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewRedisBackendDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *RedisBackendDCReconciler {
	return &RedisBackendDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *RedisBackendDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type RedisReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &RedisReconciler{}

func NewRedisReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) RedisReconciler {
	return RedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *RedisReconciler) Reconcile() (reconcile.Result, error) {
	if r.apiManager.Spec.HighAvailability != nil && r.apiManager.Spec.HighAvailability.Enabled {
		return reconcile.Result{}, nil
	}

	redis, err := r.redis()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileBackendDeploymentConfig(redis.BackendDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileBackendService(redis.BackendService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileBackendConfigMap(redis.BackendConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileBackendPVC(redis.BackendPVC())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileBackendImageStream(redis.BackendImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemDeploymentConfig(redis.SystemDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemPVC(redis.SystemPVC())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemImageStream(redis.SystemImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemService(redis.SystemService())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *RedisReconciler) redis() (*component.Redis, error) {
	optsProvider := OperatorRedisOptionsProvider{APIManagerSpec: &r.apiManager.Spec}
	opts, err := optsProvider.GetRedisOptions()
	if err != nil {
		return nil, err
	}
	return component.NewRedis(opts), nil
}

func (r *RedisReconciler) reconcileBackendDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewRedisBackendDCReconciler(r.BaseAPIManagerLogicReconciler))
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *RedisReconciler) reconcileBackendService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	err := reconciler.Reconcile(desiredService)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredService)))
	return nil
}

func (r *RedisReconciler) reconcileBackendConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	err := reconciler.Reconcile(desiredConfigMap)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredConfigMap)))
	return nil
}

func (r *RedisReconciler) reconcileBackendPVC(desiredPVC *v1.PersistentVolumeClaim) error {
	reconciler := NewPVCBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyPVCReconciler())
	err := reconciler.Reconcile(desiredPVC)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredPVC)))
	return nil
}

func (r *RedisReconciler) reconcileSystemDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewRedisSystemDCReconciler(r.BaseAPIManagerLogicReconciler))
	err := reconciler.Reconcile(desiredDeploymentConfig)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredDeploymentConfig)))
	return nil
}

func (r *RedisReconciler) reconcileSystemPVC(desiredPVC *v1.PersistentVolumeClaim) error {
	reconciler := NewPVCBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyPVCReconciler())
	err := reconciler.Reconcile(desiredPVC)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredPVC)))
	return nil
}

func (r *RedisReconciler) reconcileSystemService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	err := reconciler.Reconcile(desiredService)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredService)))
	return nil
}

func (r *RedisReconciler) reconcileBackendImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	err := reconciler.Reconcile(desiredImageStream)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredImageStream)))
	return nil
}

func (r *RedisReconciler) reconcileSystemImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	err := reconciler.Reconcile(desiredImageStream)
	if err != nil {
		return err
	}
	r.Logger().Info(fmt.Sprintf("%s reconciled", ObjectInfo(desiredImageStream)))
	return nil
}
