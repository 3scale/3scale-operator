package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *SystemReconciler {
	return &SystemReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemReconciler) reconcileFileStorage(system *component.System) error {
	if r.apiManager.Spec.System.FileStorageSpec != nil {
		if r.apiManager.Spec.System.FileStorageSpec.S3 != nil {
			return r.validateS3StorageProvidedConfiguration()
		}
		if r.apiManager.Spec.System.FileStorageSpec.DeprecatedS3 != nil {
			r.Logger().Info("Warning: deprecated amazonSimpleStorageService field in CR being used. Ignoring it... Please use simpleStorageService")
		}
	}
	// System RWX PVC, i.e. shared storage
	return r.ReconcilePersistentVolumeClaim(system.SharedStorage(), reconcilers.CreateOnlyMutator)
}

func (r *SystemReconciler) Reconcile() (reconcile.Result, error) {
	system, err := System(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileFileStorage(system)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Provider Service
	err = r.ReconcileService(system.ProviderService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Master Service
	err = r.ReconcileService(system.MasterService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Developer Service
	err = r.ReconcileService(system.DeveloperService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Sphinx Service
	err = r.ReconcileService(system.SphinxService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Memcached Service
	err = r.ReconcileService(system.MemcachedService(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// SystemApp DC
	err = r.ReconcileDeploymentConfig(system.AppDeploymentConfig(), r.systemAppDCMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Sidekiq DC
	err = r.ReconcileDeploymentConfig(system.SidekiqDeploymentConfig(), reconcilers.GenericDeploymentConfigMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Sphinx DC
	err = r.ReconcileDeploymentConfig(system.SphinxDeploymentConfig(), reconcilers.DeploymentConfigResourcesAndAffinityAndTolerationsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System CM
	err = r.ReconcileConfigMap(system.SystemConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// System CM
	err = r.ReconcileConfigMap(system.EnvironmentConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// SMTP Secret
	err = r.ReconcileSecret(system.SMTPSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// EventsHook Secret
	err = r.ReconcileSecret(system.EventsHookSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Redis Secret
	err = r.ReconcileSecret(system.RedisSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// MasterApicast  Secret
	err = r.ReconcileSecret(system.MasterApicastSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// SystemSeed Secret
	err = r.ReconcileSecret(system.SeedSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Recaptcha Secret
	err = r.ReconcileSecret(system.RecaptchaSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// SystemApp Secret
	err = r.ReconcileSecret(system.AppSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Memcached Secret
	err = r.ReconcileSecret(system.MemcachedSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// SystemApp PDB
	err = r.ReconcilePodDisruptionBudget(system.AppPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Sidekiq PDB
	err = r.ReconcilePodDisruptionBudget(system.SidekiqPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(system.SystemSidekiqPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboard(component.SystemGrafanaDashboard(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(component.SystemSidekiqPrometheusRules(r.apiManager.Namespace), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SystemReconciler) validateS3StorageProvidedConfiguration() error {
	// Nothing for reconcile.
	// Check all required fields exist
	awsCredentialsSecretName := r.apiManager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name
	if awsCredentialsSecretName == "" {
		return fmt.Errorf("no aws credentials provided")
	}

	awsSecret, err := helper.GetSecret(awsCredentialsSecretName, r.apiManager.Namespace, r.Client())
	if err != nil {
		return err
	}

	secretData := awsSecret.Data
	var result *string
	result = helper.GetSecretDataValue(secretData, component.AwsAccessKeyID)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.AwsAccessKeyID, awsCredentialsSecretName)
	}

	result = helper.GetSecretDataValue(secretData, component.AwsSecretAccessKey)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.AwsSecretAccessKey, awsCredentialsSecretName)
	}

	result = helper.GetSecretDataValue(secretData, component.AwsBucket)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.AwsBucket, awsCredentialsSecretName)
	}

	result = helper.GetSecretDataValue(secretData, component.AwsRegion)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.AwsRegion, awsCredentialsSecretName)
	}

	return nil
}

func (r *SystemReconciler) systemAppDCMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*appsv1.DeploymentConfig)
	if !ok {
		return false, fmt.Errorf("%T is not a *appsv1.DeploymentConfig", existingObj)
	}
	desired, ok := desiredObj.(*appsv1.DeploymentConfig)
	if !ok {
		return false, fmt.Errorf("%T is not a *appsv1.DeploymentConfig", desiredObj)
	}

	desiredName := common.ObjectInfo(desired)
	update := false

	// Check node affinity and tolerations

	tmpUpdate := reconcilers.DeploymentConfigAffinityReconciler(desired, existing)
	update = update || tmpUpdate

	tmpUpdate = reconcilers.DeploymentConfigTolerationsReconciler(desired, existing)
	update = update || tmpUpdate

	tmpUpdate = reconcilers.DeploymentConfigReplicasReconciler(desired, existing)
	update = update || tmpUpdate

	//
	// Check containers
	//
	if len(desired.Spec.Template.Spec.Containers) != 3 {
		panic(fmt.Sprintf("%s desired spec.template.spec.containers length changed to '%d', should be 3", desiredName, len(desired.Spec.Template.Spec.Containers)))
	}

	if len(existing.Spec.Template.Spec.Containers) != 3 {
		r.Logger().Info(fmt.Sprintf("%s spec.template.spec.containers length changed to '%d', recreating dc", desiredName, len(existing.Spec.Template.Spec.Containers)))
		existing.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		update = true
	}

	//
	// Check containers resource requirements
	//

	for idx := 0; idx < 3; idx++ {
		if !helper.CmpResources(&existing.Spec.Template.Spec.Containers[idx].Resources, &desired.Spec.Template.Spec.Containers[idx].Resources) {
			diff := cmp.Diff(existing.Spec.Template.Spec.Containers[idx].Resources, desired.Spec.Template.Spec.Containers[idx].Resources, cmpopts.IgnoreUnexported(resource.Quantity{}))
			r.Logger().Info(fmt.Sprintf("%s spec.template.spec.containers[%d].resources have changed: %s", desiredName, idx, diff))
			existing.Spec.Template.Spec.Containers[idx].Resources = desired.Spec.Template.Spec.Containers[idx].Resources
			update = true
		}
	}

	return update, nil
}

func System(cr *appsv1alpha1.APIManager, client client.Client) (*component.System, error) {
	optsProvider := NewSystemOptionsProvider(cr, cr.Namespace, client)
	opts, err := optsProvider.GetSystemOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(opts), nil
}
