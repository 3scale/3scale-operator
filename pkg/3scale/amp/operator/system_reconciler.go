package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemSphinxDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewSystemSphinxDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *SystemSphinxDCReconciler {
	return &SystemSphinxDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemSphinxDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type SystemSidekiqDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewSystemSidekiqDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *SystemSidekiqDCReconciler {
	return &SystemSidekiqDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemSidekiqDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	update := false

	tmpUpdate := DeploymentConfigReconcileReplicas(desired, existing, r.Logger())
	update = update || tmpUpdate

	tmpUpdate = DeploymentConfigReconcileContainerResources(desired, existing, r.Logger())
	update = update || tmpUpdate

	return update
}

type SystemAppDCReconciler struct {
	BaseAPIManagerLogicReconciler
}

func NewSystemAppDCReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) *SystemAppDCReconciler {
	return &SystemAppDCReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemAppDCReconciler) IsUpdateNeeded(desired, existing *appsv1.DeploymentConfig) bool {
	desiredName := ObjectInfo(desired)
	update := false

	tmpUpdate := DeploymentConfigReconcileReplicas(desired, existing, r.Logger())
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

	return update
}

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
	if r.apiManager.Spec.System.FileStorageSpec != nil {
		if r.apiManager.Spec.System.FileStorageSpec.S3 != nil {
			return r.reconcileS3Storage()
		}
		if r.apiManager.Spec.System.FileStorageSpec.DeprecatedS3 != nil {
			r.logger.Info("Warning: deprecated amazonSimpleStorageService field in CR being used. Ignoring it... Please use simpleStorageService")
		}
	}
	return r.reconcileSharedStorage(system.SharedStorage())
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

	err = r.reconcileSMTPSecret(system.SMTPSecret())
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

	err = r.reconcilePodDisruptionBudget(system.AppPodDisruptionBudget())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcilePodDisruptionBudget(system.SidekiqPodDisruptionBudget())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileMonitoringService(component.SystemSidekiqMonitoringService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileServiceMonitor(component.SystemSidekiqServiceMonitor())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileGrafanaDashboard(component.SystemSidekiqGrafanaDashboard(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcilePrometheusRules(component.SystemSidekiqPrometheusRules(r.apiManager.Namespace))
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *SystemReconciler) reconcileS3Storage() error {
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

func (r *SystemReconciler) reconcileSharedStorage(desiredPVC *v1.PersistentVolumeClaim) error {
	reconciler := NewPVCBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyPVCReconciler())
	return reconciler.Reconcile(desiredPVC)
}

func (r *SystemReconciler) reconcileProviderService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *SystemReconciler) reconcileMasterService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *SystemReconciler) reconcileDeveloperService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *SystemReconciler) reconcileSphinxService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *SystemReconciler) reconcileMemcachedService(desiredService *v1.Service) error {
	reconciler := NewServiceBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlySvcReconciler())
	return reconciler.Reconcile(desiredService)
}

func (r *SystemReconciler) reconcileAppDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewSystemAppDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *SystemReconciler) reconcileSidekiqDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewSystemSidekiqDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *SystemReconciler) reconcileSphinxDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	reconciler := NewDeploymentConfigBaseReconciler(r.BaseAPIManagerLogicReconciler, NewSystemSphinxDCReconciler(r.BaseAPIManagerLogicReconciler))
	return reconciler.Reconcile(desiredDeploymentConfig)
}

func (r *SystemReconciler) reconcileSystemConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	return reconciler.Reconcile(desiredConfigMap)
}

func (r *SystemReconciler) reconcileEnvironmentConfigMap(desiredConfigMap *v1.ConfigMap) error {
	reconciler := NewConfigMapBaseReconciler(r.BaseAPIManagerLogicReconciler, NewCreateOnlyConfigMapReconciler())
	return reconciler.Reconcile(desiredConfigMap)
}

func (r *SystemReconciler) reconcileSMTPSecret(desiredSecret *v1.Secret) error {
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemReconciler) reconcileEventsHookSecret(desiredSecret *v1.Secret) error {
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemReconciler) reconcileRedisSecret(desiredSecret *v1.Secret) error {
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemReconciler) reconcileMasterApicastSecret(desiredSecret *v1.Secret) error {
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemReconciler) reconcileSeedSecret(desiredSecret *v1.Secret) error {
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemReconciler) reconcileRecaptchaSecret(desiredSecret *v1.Secret) error {
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemReconciler) reconcileAppSecret(desiredSecret *v1.Secret) error {
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func (r *SystemReconciler) reconcileMemcachedSecret(desiredSecret *v1.Secret) error {
	reconciler := NewSecretBaseReconciler(r.BaseAPIManagerLogicReconciler, NewDefaultsOnlySecretReconciler())
	return reconciler.Reconcile(desiredSecret)
}

func System(cr *appsv1alpha1.APIManager, client client.Client) (*component.System, error) {
	optsProvider := NewSystemOptionsProvider(cr, cr.Namespace, client)
	opts, err := optsProvider.GetSystemOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(opts), nil
}
