package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &SystemReconciler{}

func NewSystemReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) SystemReconciler {
	return SystemReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemReconciler) reconcileFileStorage(system *component.System) error {
	if r.apiManager.Spec.System != nil && r.apiManager.Spec.System.FileStorageSpec != nil {
		if r.apiManager.Spec.System.FileStorageSpec.PVC != nil {
			return r.reconcileSharedStorage(system.SharedStorage())
		} else if r.apiManager.Spec.System.FileStorageSpec.S3 != nil {
			return r.reconcileS3AWSSecret(system.S3AWSSecret())
		} else {
			return fmt.Errorf("No FileStorage spec specified. FileStorage is mandatory")
		}
	}
	return nil
}

func (r *SystemReconciler) reconcileS3AWSSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemReconciler) Reconcile() (reconcile.Result, error) {
	system, err := r.system()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileFileStorage(system)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileProviderService(system.ProviderService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileMasterService(system.MasterService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileDeveloperService(system.DeveloperService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSphinxService(system.SphinxService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileMemcachedService(system.MemcachedService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileAppDeploymentConfig(system.AppDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSidekiqDeploymentConfig(system.SidekiqDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSphinxDeploymentConfig(system.SphinxDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSystemConfigMap(system.SystemConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileEnvironmentConfigMap(system.EnvironmentConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSMTPConfigMap(system.SMTPConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileEventsHookSecret(system.EventsHookSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileRedisSecret(system.RedisSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileMasterApicastSecret(system.MasterApicastSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileSeedSecret(system.SeedSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileRecaptchaSecret(system.RecaptchaSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileAppSecret(system.AppSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileMemcachedSecret(system.MemcachedSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	// TODO rethink where to create the system-database secret
	if r.apiManager.Spec.HighAvailability != nil && r.apiManager.Spec.HighAvailability.Enabled {
		ha, err := r.highAvailability()
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.reconcileDatabaseHASecret(ha.SystemDatabaseSecret())
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *SystemReconciler) system() (*component.System, error) {
	optsProvider := OperatorSystemOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetSystemOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(opts), nil
}

func (r *SystemReconciler) highAvailability() (*component.HighAvailability, error) {
	optsProvider := OperatorHighAvailabilityOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetHighAvailabilityOptions()
	if err != nil {
		return nil, err
	}
	return component.NewHighAvailability(opts), nil
}

func (r *SystemReconciler) reconcileDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	err := r.InitializeAsAPIManagerObject(desiredDeploymentConfig)
	if err != nil {
		return err
	}

	return r.deploymentConfigReconciler.Reconcile(desiredDeploymentConfig)
}

func (r *SystemReconciler) reconcileSecret(desiredSecret *v1.Secret) error {
	err := r.InitializeAsAPIManagerObject(desiredSecret)
	if err != nil {
		return err
	}
	return r.secretReconciler.Reconcile(desiredSecret)
}

func (r *SystemReconciler) reconcileConfigMap(desiredConfigMap *v1.ConfigMap) error {
	err := r.InitializeAsAPIManagerObject(desiredConfigMap)
	if err != nil {
		return err
	}

	return r.configMapReconciler.Reconcile(desiredConfigMap)
}

func (r *SystemReconciler) reconcilePersistentVolumeClaim(desiredPVC *v1.PersistentVolumeClaim) error {
	err := r.InitializeAsAPIManagerObject(desiredPVC)
	if err != nil {
		return err
	}

	return r.persistentVolumeClaimReconciler.Reconcile(desiredPVC)
}

func (r *SystemReconciler) reconcileService(desiredService *v1.Service) error {
	err := r.InitializeAsAPIManagerObject(desiredService)
	if err != nil {
		return err
	}
	return r.serviceReconciler.Reconcile(desiredService)
}

func (r *SystemReconciler) reconcileSharedStorage(desiredPVC *v1.PersistentVolumeClaim) error {
	return r.reconcilePersistentVolumeClaim(desiredPVC)
}

func (r *SystemReconciler) reconcileProviderService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *SystemReconciler) reconcileMasterService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *SystemReconciler) reconcileDeveloperService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *SystemReconciler) reconcileRedisService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *SystemReconciler) reconcileSphinxService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *SystemReconciler) reconcileMemcachedService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *SystemReconciler) reconcileAppDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *SystemReconciler) reconcileSidekiqDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *SystemReconciler) reconcileSphinxDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *SystemReconciler) reconcileSystemConfigMap(desiredConfigMap *v1.ConfigMap) error {
	return r.reconcileConfigMap(desiredConfigMap)
}

func (r *SystemReconciler) reconcileEnvironmentConfigMap(desiredConfigMap *v1.ConfigMap) error {
	return r.reconcileConfigMap(desiredConfigMap)
}

func (r *SystemReconciler) reconcileSMTPConfigMap(desiredConfigMap *v1.ConfigMap) error {
	return r.reconcileConfigMap(desiredConfigMap)
}

func (r *SystemReconciler) reconcileEventsHookSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemReconciler) reconcileRedisSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemReconciler) reconcileMasterApicastSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemReconciler) reconcileSeedSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemReconciler) reconcileRecaptchaSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemReconciler) reconcileAppSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemReconciler) reconcileMemcachedSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *SystemReconciler) reconcileDatabaseHASecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}
