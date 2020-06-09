package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemRedisReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemRedisReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *SystemRedisReconciler {
	return &SystemRedisReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemRedisReconciler) Reconcile() (reconcile.Result, error) {
	if r.apiManager.IsSystemRedisDatabaseExternal() {
		err := r.checkSystemRedisSecretFields()
		return reconcile.Result{}, err
	}

	redis, err := SystemRedis(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis DC
	err = r.ReconcileDeploymentConfig(redis.DeploymentConfig(), reconcilers.DeploymentConfigResourcesMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis Service
	err = r.ReconcileService(redis.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis PVC
	err = r.ReconcilePersistentVolumeClaim(redis.PersistentVolumeClaim(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System redis IS
	err = r.ReconcileImagestream(redis.ImageStream(), reconcilers.GenericImageStreamMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func SystemRedis(apimanager *appsv1alpha1.APIManager) (*component.SystemRedis, error) {
	optsProvider := NewSystemRedisOptionsProvider(apimanager)
	opts, err := optsProvider.GetSystemRedisOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemRedis(opts), nil
}

func (r *SystemRedisReconciler) checkSystemRedisSecretFields() error {
	secretSource := helper.NewSecretSource(r.Client(), r.apiManager.Namespace)

	cases := []struct {
		secretName  string
		secretField string
	}{
		{
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisURLFieldName,
		},
		{
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisMessageBusRedisURLFieldName,
		},
	}

	for _, option := range cases {
		_, err := secretSource.RequiredFieldValueFromRequiredSecret(option.secretName, option.secretField)
		if err != nil {
			return err
		}
	}

	return nil
}
