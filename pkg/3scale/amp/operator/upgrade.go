package operator

import (
	"context"
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/go-logr/logr"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpgradeApiManager struct {
	Cr              *appsv1alpha1.APIManager
	Client          client.Client
	Logger          logr.Logger
	ApiClientReader client.Reader
	Scheme          *runtime.Scheme
}

func (u *UpgradeApiManager) Upgrade() (reconcile.Result, error) {
	res, err := u.upgradeSystemAppPreHookPodEnv()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeImages()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeSystemSMTP()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeImages() (reconcile.Result, error) {
	res, err := u.upgradeAMPImageStreams()
	if res.Requeue || err != nil {
		return res, err
	}

	if !u.highAvailabilityModeEnabled() {
		res, err = u.upgradeBackendRedisImageStream()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemRedisImageStream()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemDatabaseImageStream()
		if res.Requeue || err != nil {
			return res, err
		}
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeAMPImageStreams() (reconcile.Result, error) {
	// implement upgrade procedure by reconcile procedure
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewAMPImagesReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeSystemSMTP() (reconcile.Result, error) {
	res, err := u.migrateSystemSMTPData()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeSystemSMTPEnvVars()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeSystemSMTPEnvVars() (reconcile.Result, error) {
	res, err := u.upgradeSystemSidekiqEnvVars()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeSystemAppEnvVars()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeSystemSidekiqEnvVars() (reconcile.Result, error) {
	system, err := System(u.Cr, u.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	desiredSidekiqDeploymentConfig := system.SidekiqDeploymentConfig()
	existingSidekiqDeploymentConfig := &appsv1.DeploymentConfig{}
	err = u.Client.Get(context.TODO(), types.NamespacedName{Name: desiredSidekiqDeploymentConfig.Name, Namespace: u.Cr.Namespace}, existingSidekiqDeploymentConfig)
	if err != nil {
		return reconcile.Result{}, err
	}

	changed, err := u.ensureDeploymentConfigPodTemplateEnvVars(desiredSidekiqDeploymentConfig, existingSidekiqDeploymentConfig)
	if err != nil {
		return reconcile.Result{}, err
	}

	if changed {
		u.Logger.Info(fmt.Sprintf("Update object %s", ObjectInfo(existingSidekiqDeploymentConfig)))
		err := u.Client.Update(context.TODO(), existingSidekiqDeploymentConfig)
		return reconcile.Result{Requeue: true}, err
	}
	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeSystemAppEnvVars() (reconcile.Result, error) {
	system, err := System(u.Cr, u.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	desiredSystemAppDeploymentConfig := system.AppDeploymentConfig()
	existingSystemAppDeploymentConfig := &appsv1.DeploymentConfig{}
	err = u.Client.Get(context.TODO(), types.NamespacedName{Name: desiredSystemAppDeploymentConfig.Name, Namespace: u.Cr.Namespace}, existingSystemAppDeploymentConfig)
	if err != nil {
		return reconcile.Result{}, err
	}

	changed, err := u.ensureDeploymentConfigPodTemplateEnvVars(desiredSystemAppDeploymentConfig, existingSystemAppDeploymentConfig)
	if err != nil {
		return reconcile.Result{}, err
	}

	tmpChanged := u.ensureDeploymentConfigPreHookPodEnvVars(desiredSystemAppDeploymentConfig, existingSystemAppDeploymentConfig)
	changed = changed || tmpChanged

	if changed {
		u.Logger.Info(fmt.Sprintf("Update object %s", ObjectInfo(existingSystemAppDeploymentConfig)))
		err := u.Client.Update(context.TODO(), existingSystemAppDeploymentConfig)
		return reconcile.Result{Requeue: true}, err
	}
	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) migrateSystemSMTPData() (reconcile.Result, error) {
	existingConfigMap := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{
		Name:      "smtp",
		Namespace: u.Cr.Namespace,
	}
	err := u.Client.Get(context.TODO(), configMapNamespacedName, existingConfigMap)
	if err != nil {
		return reconcile.Result{}, err
	}
	existingSecret := &v1.Secret{}
	secretNamespacedName := types.NamespacedName{
		Name:      component.SystemSecretSystemSMTPSecretName,
		Namespace: u.Cr.Namespace,
	}
	err = u.Client.Get(context.TODO(), secretNamespacedName, existingSecret)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	if errors.IsNotFound(err) {
		system, err := System(u.Cr, u.Client)
		if err != nil {
			return reconcile.Result{}, err
		}
		// We rely on System options provider not needing the system-smtp secret
		// existing so we obtain the default one and overwrite the data
		// with the existing ConfigMap data and we create the secret
		existingSecret = system.SMTPSecret()
		existingSecret.SetNamespace(u.Cr.Namespace)
		err = controllerutil.SetControllerReference(u.Cr, existingSecret, u.Scheme)
		if err != nil {
			return reconcile.Result{}, err
		}
		// We make sure StringData is nil so it does not get precedence over Data.
		// We use Data to set the secret and not StringData due to at the time
		// of writing this comment when using the Kubernetes FakeClient the
		// mocked client does not convert from StringData to Data, producing a
		// different behavior than with the real code execution
		existingSecret.StringData = nil
		existingSecret.Data = helper.GetSecretDataFromStringData(existingConfigMap.Data)

		u.Logger.Info(fmt.Sprintf("Create object %s", ObjectInfo(existingSecret)))
		err = u.Client.Create(context.TODO(), existingSecret)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, err
	}

	changed := false
	secretStringData := helper.GetSecretStringDataFromData(existingSecret.Data)
	changed = !reflect.DeepEqual(existingConfigMap.Data, secretStringData)
	if changed {
		existingSecret.Data = helper.GetSecretDataFromStringData(existingConfigMap.Data)
		u.Logger.Info(fmt.Sprintf("Update object %s", ObjectInfo(existingSecret)))
		err := u.Client.Update(context.TODO(), existingSecret)
		return reconcile.Result{Requeue: true}, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) ensureDeploymentConfigPreHookPodEnvVars(desired, existing *appsv1.DeploymentConfig) bool {
	changed := false
	desiredPreHookPod := desired.Spec.Strategy.RollingParams.Pre.ExecNewPod
	existingPrehookPod := existing.Spec.Strategy.RollingParams.Pre.ExecNewPod
	if !reflect.DeepEqual(existingPrehookPod.Env, desiredPreHookPod.Env) {
		existingPrehookPod.Env = desiredPreHookPod.Env
		changed = true
	}
	return changed
}

func (u *UpgradeApiManager) ensureDeploymentConfigPodTemplateEnvVars(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	if len(existing.Spec.Template.Spec.Containers) != len(desired.Spec.Template.Spec.Containers) {
		return false, fmt.Errorf("%s desired containers length does not match existing containers length", desired.Name)
	}

	changed := false
	for idx := range existing.Spec.Template.Spec.Containers {
		desiredContainer := &desired.Spec.Template.Spec.Containers[idx]
		existingContainer := &existing.Spec.Template.Spec.Containers[idx]
		if !reflect.DeepEqual(existingContainer.Env, desiredContainer.Env) {
			existingContainer.Env = desiredContainer.Env
			changed = true
		}
	}

	return changed, nil
}

func (u *UpgradeApiManager) highAvailabilityModeEnabled() bool {
	return u.Cr.Spec.HighAvailability != nil && u.Cr.Spec.HighAvailability.Enabled
}

func (u *UpgradeApiManager) upgradeBackendRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.Cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewImageStreamBaseReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr), NewImageStreamGenericReconciler())
	return reconcile.Result{}, reconciler.Reconcile(redis.BackendImageStream())
}

func (u *UpgradeApiManager) upgradeSystemRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.Cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewImageStreamBaseReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr), NewImageStreamGenericReconciler())
	return reconcile.Result{}, reconciler.Reconcile(redis.SystemImageStream())
}

func (u *UpgradeApiManager) upgradeSystemDatabaseImageStream() (reconcile.Result, error) {
	if u.Cr.Spec.System.DatabaseSpec.MySQL != nil {
		return u.upgradeSystemMySQLImageStream()
	}

	if u.Cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.upgradeSystemPostgreSQLImageStream()
	}

	return reconcile.Result{}, fmt.Errorf("System database is not set")
}

func (u *UpgradeApiManager) upgradeSystemMySQLImageStream() (reconcile.Result, error) {
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewSystemMySQLImageReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeSystemPostgreSQLImageStream() (reconcile.Result, error) {
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewSystemPostgreSQLImageReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeSystemAppPreHookPodEnv() (reconcile.Result, error) {
	existingDeploymentConfig := &appsv1.DeploymentConfig{}
	err := u.Client.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: u.Cr.Namespace}, existingDeploymentConfig)
	if err != nil {
		return reconcile.Result{}, err
	}

	system, err := System(u.Cr, u.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	desiredDeploymentConfig := system.AppDeploymentConfig()
	changed := u.ensureDeploymentConfigPreHookPodEnvVars(desiredDeploymentConfig, existingDeploymentConfig)

	if changed {
		u.Logger.Info(fmt.Sprintf("Update object %s", ObjectInfo(existingDeploymentConfig)))
		err := u.Client.Update(context.TODO(), existingDeploymentConfig)
		return reconcile.Result{Requeue: true}, err
	}

	return reconcile.Result{}, nil
}
