package operator

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"regexp"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
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
	if r.apiManager.IsS3Enabled() {
		return nil
	}

	if r.apiManager.Spec.System.FileStorageSpec != nil &&
		r.apiManager.Spec.System.FileStorageSpec.DeprecatedS3 != nil {
		r.Logger().Info("Warning: deprecated amazonSimpleStorageService field in CR being used. Ignoring it... Please use simpleStorageService")
	}
	// System RWX PVC, i.e. shared storage
	return r.ReconcilePersistentVolumeClaim(system.SharedStorage(), reconcilers.CreateOnlyMutator)
}

func (r *SystemReconciler) Reconcile() (reconcile.Result, error) {
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	system, err := System(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileFileStorage(system)
	if err != nil {
		return reconcile.Result{}, err
	}

	serviceMutators := []reconcilers.MutateFn{
		reconcilers.CreateOnlyMutator,
		reconcilers.ServiceSelectorMutator,
	}

	// Provider Service
	err = r.ReconcileService(system.ProviderService(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Master Service
	err = r.ReconcileService(system.MasterService(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Developer Service
	err = r.ReconcileService(system.DeveloperService(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Memcached Service
	err = r.ReconcileService(system.MemcachedService(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// System CM
	err = r.ReconcileConfigMap(system.SystemConfigMap(), systemConfigMapMutator)
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

	// Used to synchronize rollout of system Deployments
	systemComponentsReady := true

	// If the system-app Deployment generation has changed, delete the PreHook/PostHook Jobs so they can be recreated
	generationChanged, err := helper.HasAppGenerationChanged(component.SystemAppPreHookJobName, component.SystemAppDeploymentName, r.apiManager.GetNamespace(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}
	if generationChanged {
		err = helper.DeleteJob(component.SystemAppPreHookJobName, r.apiManager.GetNamespace(), r.Client())
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	generationChanged, err = helper.HasAppGenerationChanged(component.SystemAppPostHookJobName, component.SystemAppDeploymentName, r.apiManager.GetNamespace(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}
	if generationChanged {
		err = helper.DeleteJob(component.SystemAppPostHookJobName, r.apiManager.GetNamespace(), r.Client())
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Used to synchronize the system-app Deployment with the PreHook/PostHook Jobs
	currentAppDeploymentGeneration, err := getSystemAppDeploymentGeneration(r.apiManager.GetNamespace(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// SystemApp PreHook Job
	preHookJob := system.AppPreHookJob(ampImages.Options.SystemImage, currentAppDeploymentGeneration)
	err = r.ReconcileJob(preHookJob, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Block reconciling system-app Deployment until PreHook Job has completed
	if !helper.HasJobCompleted(preHookJob.Name, preHookJob.Namespace, r.Client()) {
		systemComponentsReady = false
	}

	if systemComponentsReady {
		// SystemApp Deployment
		systemAppDeploymentMutators := []reconcilers.DMutateFn{
			reconcilers.DeploymentAffinityMutator,
			reconcilers.DeploymentTolerationsMutator,
			reconcilers.DeploymentPodTemplateLabelsMutator,
			reconcilers.DeploymentPriorityClassMutator,
			reconcilers.DeploymentTopologySpreadConstraintsMutator,
			reconcilers.DeploymentPodTemplateAnnotationsMutator,
			r.systemAppDeploymentResourceMutator,
			reconcilers.DeploymentRemoveDuplicateEnvVarMutator,
			reconcilers.DeploymentPodContainerImageMutator,
			r.systemZyncEnvVarMutator,
		}
		if r.apiManager.Spec.System.AppSpec.Replicas != nil {
			systemAppDeploymentMutators = append(systemAppDeploymentMutators, reconcilers.DeploymentReplicasMutator)
		}
		err = r.ReconcileDeployment(system.AppDeployment(ampImages.Options.SystemImage), reconcilers.DeploymentMutator(systemAppDeploymentMutators...))
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Block reconciling PostHook Job unless BOTH the PreHook Job has completed and the system-app Deployment is ready and not in the process of updating
	deployment := &k8sappsv1.Deployment{}
	err = r.Client().Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: r.apiManager.GetNamespace(),
		Name:      component.SystemAppDeploymentName,
	}, deployment)
	if err != nil && !k8serr.IsNotFound(err) {
		return reconcile.Result{}, err
	}
	if k8serr.IsNotFound(err) || !helper.IsDeploymentAvailable(deployment) || helper.IsDeploymentProgressing(deployment) || !helper.HasJobCompleted(preHookJob.Name, preHookJob.Namespace, r.Client()) {
		systemComponentsReady = false
	}

	// SystemApp PostHook Job
	if systemComponentsReady {
		err = r.ReconcileJob(system.AppPostHookJob(ampImages.Options.SystemImage, currentAppDeploymentGeneration), reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.SystemAppDeploymentName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		systemComponentsReady = false
	}

	// Sidekiq Deployment
	sidekiqDeploymentMutators := []reconcilers.DMutateFn{
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		reconcilers.DeploymentRemoveDuplicateEnvVarMutator,
		reconcilers.DeploymentArgsMutator,
		reconcilers.DeploymentProbesMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
		reconcilers.DeploymentPodInitContainerImageMutator,
		r.systemZyncEnvVarMutator,
	}
	if r.apiManager.Spec.System.SidekiqSpec.Replicas != nil {
		sidekiqDeploymentMutators = append(sidekiqDeploymentMutators, reconcilers.DeploymentReplicasMutator)
	}

	err = r.ReconcileDeployment(system.SidekiqDeployment(ampImages.Options.SystemImage), reconcilers.DeploymentMutator(sidekiqDeploymentMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err = upgrade.MigrateDeploymentConfigToDeployment(component.SystemSidekiqName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		systemComponentsReady = false
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

	err = r.ReconcilePodMonitor(system.SystemAppPodMonitor(), reconcilers.GenericPodMonitorMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	sumRate, err := helper.SumRateForOpenshiftVersion(r.Context(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboards(system.SystemGrafanaV5Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileGrafanaDashboards(system.SystemGrafanaV4Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(system.SystemAppPrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(system.SystemSidekiqPrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Requeue if any of the system-app Deployment's components aren't ready
	if !systemComponentsReady {
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Check if enough routes exist
	routesExist, err := helper.DefaultRoutesReady(r.apiManager, r.Client(), r.logger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if zync deployments are ready
	zyncReady, err := r.isZyncReady()
	if err != nil {
		return reconcile.Result{}, err
	}

	// If zync is enabled and zync is ready but the routes are missing then create the ResyncRoutes Job
	if system.Options.ZyncEnabled && zyncReady && !routesExist {
		resyncRoutesJob := system.ResyncRoutesJob(ampImages.Options.SystemImage)
		err = r.ReconcileJob(resyncRoutesJob, reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		// Zync is either disabled or not ready or the routes already exist
		// In this scenario, delete the ResyncRoutes Job (if it exists) in case it needs to be created again later
		err = helper.DeleteJob(component.ResyncRoutesJobName, r.apiManager.GetNamespace(), r.Client())
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func getSystemAppDeploymentGeneration(namespace string, client k8sclient.Client) (int64, error) {
	deployment := &k8sappsv1.Deployment{}
	err := client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: namespace,
		Name:      component.SystemAppDeploymentName,
	}, deployment)

	// Return error if can't get Deployment
	if err != nil && !k8serr.IsNotFound(err) {
		return 0, fmt.Errorf("error getting deployment %s: %w", deployment.Name, err)
	}

	// Return 1 if the Deployment doesn't exist yet
	if k8serr.IsNotFound(err) {
		return 1, nil
	}

	return deployment.Generation, nil
}

func (r *SystemReconciler) systemAppDeploymentResourceMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	desiredName := common.ObjectInfo(desired)
	update := false

	// Check containers
	if len(desired.Spec.Template.Spec.Containers) != 3 {
		return false, fmt.Errorf(fmt.Sprintf("%s desired spec.template.spec.containers length changed to '%d', should be 3", desiredName, len(desired.Spec.Template.Spec.Containers)))
	}

	if len(existing.Spec.Template.Spec.Containers) != 3 {
		r.Logger().Info(fmt.Sprintf("%s spec.template.spec.containers length changed to '%d', recreating dc", desiredName, len(existing.Spec.Template.Spec.Containers)))
		existing.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		update = true
	}

	// Check containers resource requirements
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

func (r *SystemReconciler) systemZyncEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	update := false

	// Reconcile ZYNC_AUTHENTICATION_TOKEN env var
	for idx := range existing.Spec.Template.Spec.Containers {
		tmpChanged := helper.EnvVarReconciler(
			desired.Spec.Template.Spec.Containers[idx].Env,
			&existing.Spec.Template.Spec.Containers[idx].Env,
			"ZYNC_AUTHENTICATION_TOKEN")
		update = update || tmpChanged
	}

	return update, nil
}

func (r *SystemReconciler) isZyncReady() (bool, error) {
	zyncDeploymentNames := []string{"zync", "zync-database", "zync-que"}

	for _, deploymentName := range zyncDeploymentNames {
		deployment := &k8sappsv1.Deployment{}
		err := r.Client().Get(r.Context(), k8sclient.ObjectKey{
			Namespace: r.apiManager.GetNamespace(),
			Name:      deploymentName,
		}, deployment)

		// Return error if can't get deployment
		if err != nil && !k8serr.IsNotFound(err) {
			return false, fmt.Errorf("error getting deployment %s: %w", deployment.Name, err)
		}

		// Return without error if deployment doesn't exist to prevent spamming log with errors
		if k8serr.IsNotFound(err) {
			return false, nil
		}

		// Return false because deployment is not yet available
		if !helper.IsDeploymentAvailable(deployment) || helper.IsDeploymentProgressing(deployment) {
			return false, nil
		}
	}

	// All zync deployments exist and are available
	return true, nil
}

// systemConfigMapMutator creates facilitates the creation of the ConfigMap on the first reconcile loop
// It also will update the endpoint in case zync is enabled|disabled while preserving all other values in the .data
func systemConfigMapMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*corev1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", existingObj)
	}
	desired, ok := desiredObj.(*corev1.ConfigMap)
	if !ok {
		return false, fmt.Errorf("%T is not a *v1.ConfigMap", desiredObj)
	}

	zyncFieldKey := "zync.yml"
	desiredZyncConfigString := desired.Data[zyncFieldKey]
	existingZyncConfigString := existing.Data[zyncFieldKey]

	// Define a struct to unmarshal the YAML into
	type ZyncConfig struct {
		Production struct {
			Endpoint       string `yaml:"endpoint"`
			Authentication struct {
				Token string `yaml:"token"`
			} `yaml:"authentication"`
			ConnectTimeout int    `yaml:"connect_timeout"`
			SendTimeout    int    `yaml:"send_timeout"`
			ReceiveTimeout int    `yaml:"receive_timeout"`
			RootURL        string `yaml:"root_url"`
		} `yaml:"production"`
	}

	// Unmarshal the desiredZyncConfig
	var desiredZyncConfig ZyncConfig
	err := yaml.Unmarshal([]byte(desiredZyncConfigString), &desiredZyncConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Extract the desiredEndpoint
	desiredEndpoint := desiredZyncConfig.Production.Endpoint

	// Update the endpoint in existingConfig with the one from desiredConfig
	re := regexp.MustCompile(`(?m)^ *endpoint: '.*'`)
	reconciledZyncConfigString := re.ReplaceAllString(existingZyncConfigString, fmt.Sprintf("  endpoint: '%s'", desiredEndpoint))

	// Assign reconciledZyncConfigString to the ConfigMap's data
	desired.Data[zyncFieldKey] = reconciledZyncConfigString

	// Update the zync.yml field in the ConfigMap
	updated := reconcilers.ConfigMapReconcileField(desired, existing, "zync.yml")

	return updated, nil
}

func System(cr *appsv1alpha1.APIManager, client k8sclient.Client) (*component.System, error) {
	optsProvider := NewSystemOptionsProvider(cr, cr.Namespace, client)
	opts, err := optsProvider.GetSystemOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(opts), nil
}
