package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
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
	redis, err := Redis(r.apiManager)
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

func Redis(apimanager *appsv1alpha1.APIManager) (*component.Redis, error) {
	optsProvider := NewRedisOptionsProvider(apimanager)
	opts, err := optsProvider.GetRedisOptions()
	if err != nil {
		return nil, err
	}
	return component.NewRedis(opts), nil
}

func (r *RedisReconciler) reconcileBackendDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewRedisBackendDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *RedisReconciler) reconcileBackendService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *RedisReconciler) reconcileBackendConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	return reconciler.Reconcile(desiredConfigMap)
}

func (r *RedisReconciler) reconcileBackendPVC(desiredPVC *v1.PersistentVolumeClaim) error {
	reconciler := NewPVCBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyPVCReconciler())
	return reconciler.Reconcile(desiredPVC)
}

func (r *RedisReconciler) reconcileSystemDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewRedisSystemDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *RedisReconciler) reconcileSystemPVC(desiredPVC *v1.PersistentVolumeClaim) error {
	reconciler := NewPVCBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyPVCReconciler())
	return reconciler.Reconcile(desiredPVC)
}

func (r *RedisReconciler) reconcileSystemService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *RedisReconciler) reconcileBackendImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}

func (r *RedisReconciler) reconcileSystemImageStream(desiredImageStream *imagev1.ImageStream) error {
	reconciler := NewImageStreamBaseReconciler(r.BaseAPIManagerLogicReconciler, NewImageStreamGenericReconciler())
	return reconciler.Reconcile(desiredImageStream)
}
