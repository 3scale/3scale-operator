package operator

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
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
	currentNameSpace := r.apiManager.GetNamespace()

	// Used to synchronize the system-app Deployment with the PreHook/PostHook Jobs
	currentAppDeploymentRevision, err := getSystemAppDeploymentRevision(currentNameSpace, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if the system image has changed (indicating an upgrade)
	imageChanged, err := hasSystemImageChanged(currentNameSpace, ampImages.Options.SystemImage, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// SystemApp PreHook Job
	var preHookJob *v1.Job
	if imageChanged {
		// Version upgrades, or image changes, occur before the Deployment is
		// updated, so the revision version of job is increased by 1.
		//   - Job revision == (Deployment revision + 1) <=> We're mid-upgrade, preserve the job
		//   - Job revision == 0 <=> Job does not exist, don't try deleting
		targetRevision := currentAppDeploymentRevision + 1
		currentJobRevision, err := helper.GetAppRevision(component.SystemAppPreHookJobName, currentNameSpace, r.Client())
		if err != nil {
			return reconcile.Result{}, err
		}

		// Delete the old job if it exists and doesn't match the target revision
		// This handles:
		// - Normal upgrades (job exists with same revision)
		// - Edge cases (deployment revision updated externally)
		// - Job deleted (currentJobRevision = 0)
		// - Unknown state (job exists without annotation, currentJobRevision = -1)
		if currentJobRevision != 0 && currentJobRevision != targetRevision {
			err = helper.DeleteJob(r.Context(), component.SystemAppPreHookJobName, currentNameSpace, r.Client())
			if err != nil {
				// If job is still running, requeue and wait for it to complete
				if helper.IsJobStillRunning(err) {
					r.Logger().Info("Cannot delete PreHook job because it is still running, will requeue")
					return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
				}
				return reconcile.Result{}, err
			}
			if currentJobRevision == -1 {
				r.Logger().Info("Deleted PreHook job that was missing revision annotation")
			}
		}

		preHookJob = system.AppPreHookJob(ampImages.Options.SystemImage, currentNameSpace, targetRevision)
	} else {
		// Normal rollout path
		rerunPreHook, err := needAppHookJobRerun(component.SystemAppPreHookJobName, currentAppDeploymentRevision, currentNameSpace, r.Client())
		if err != nil {
			return reconcile.Result{}, err
		}

		if rerunPreHook {
			err = helper.DeleteJob(r.Context(), component.SystemAppPreHookJobName, currentNameSpace, r.Client())
			if err != nil {
				// If job is still running, requeue and wait for it to complete
				if helper.IsJobStillRunning(err) {
					r.Logger().Info("Cannot delete PreHook job because it is still running, will requeue")
					return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
				}
				return reconcile.Result{}, err
			}
		}
		preHookJob = system.AppPreHookJob(ampImages.Options.SystemImage, currentNameSpace, currentAppDeploymentRevision)
	}

	err = r.ReconcileJob(preHookJob, reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Block reconciling system-app Deployment until PreHook Job has completed
	finished := helper.HasJobCompleted(r.Context(), preHookJob, r.Client())
	if !finished {
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
			reconcilers.DeploymentPodInitContainerMutator,
			reconcilers.DeploymentRemoveDuplicateEnvVarMutator,
			reconcilers.DeploymentPodContainerImageMutator,
			r.systemZyncEnvVarMutator,
			r.systemDatabaseTLSEnvVarMutator,
			r.systemRedisTLSEnvVarMutator,
			systemDeploymentVolumesMutator,
			systemDeploymentInitContainerVolumeMountsMutator,
			systemDeploymentContainerVolumeMountsMutator,
		}
		if r.apiManager.Spec.System.AppSpec.Replicas != nil {
			systemAppDeploymentMutators = append(systemAppDeploymentMutators, reconcilers.DeploymentReplicasMutator)
		}

		appDeployment, err := system.AppDeployment(r.Context(), r.Client(), ampImages.Options.SystemImage)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.ReconcileDeployment(appDeployment, reconcilers.DeploymentMutator(systemAppDeploymentMutators...))
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Block reconciling PostHook Job unless BOTH the PreHook Job has completed and the system-app Deployment is ready and not in the process of updating
	deployment := &k8sappsv1.Deployment{}
	err = r.Client().Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: currentNameSpace,
		Name:      component.SystemAppDeploymentName,
	}, deployment)
	if err != nil && !k8serr.IsNotFound(err) {
		return reconcile.Result{}, err
	}
	if k8serr.IsNotFound(err) || !helper.IsDeploymentAvailable(deployment) || helper.IsDeploymentProgressing(deployment) || !finished {
		systemComponentsReady = false
	}

	rerunPostHook, err := needAppHookJobRerun(component.SystemAppPostHookJobName, currentAppDeploymentRevision, currentNameSpace, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	if rerunPostHook {
		err = helper.DeleteJob(r.Context(), component.SystemAppPostHookJobName, currentNameSpace, r.Client())
		if err != nil {
			// If job is still running, requeue and wait for it to complete
			if helper.IsJobStillRunning(err) {
				r.Logger().Info("Cannot delete PostHook job because it is still running, will requeue")
				return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
			}
			return reconcile.Result{}, err
		}
	}

	// SystemApp PostHook Job
	if systemComponentsReady {
		err = r.ReconcileJob(system.AppPostHookJob(ampImages.Options.SystemImage, currentNameSpace, currentAppDeploymentRevision), reconcilers.CreateOnlyMutator)
		if err != nil {
			return reconcile.Result{}, err
		}
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
		reconcilers.DeploymentPodInitContainerMutator,
		r.systemZyncEnvVarMutator,
		r.systemDatabaseTLSEnvVarMutator,
		r.systemRedisTLSEnvVarMutator,
		sidekiqDeploymentVolumesMutator,
		sidekiqDeploymentInitContainerVolumeMountsMutator,
		sidekiqDeploymentContainerVolumeMountsMutator,
	}

	if r.apiManager.Spec.System.SidekiqSpec.Replicas != nil {
		sidekiqDeploymentMutators = append(sidekiqDeploymentMutators, reconcilers.DeploymentReplicasMutator)
	}

	sidekiqDeployment, err := system.SidekiqDeployment(r.Context(), r.Client(), ampImages.Options.SystemImage)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileDeployment(sidekiqDeployment, reconcilers.DeploymentMutator(sidekiqDeploymentMutators...))
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

	return reconcile.Result{}, nil
}

// Revision was copied from k8s.io/kubernetes/pkg/controller/deployment/util/deployment_util.go release-1.17

const (
	// RevisionAnnotation is the revision annotation of a deployment's replica sets which records its rollout sequence
	RevisionAnnotation = "deployment.kubernetes.io/revision"
)

func getSystemAppDeploymentRevision(namespace string, client k8sclient.Client) (int64, error) {
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

	acc, err := meta.Accessor(deployment)
	if err != nil {
		return 0, err
	}

	v, ok := acc.GetAnnotations()[RevisionAnnotation]

	if !ok {
		return 0, nil
	}

	return strconv.ParseInt(v, 10, 64)
}

// hasSystemImageChanged checks if the system image has changed compared to the currently deployed image
// This indicates an upgrade scenario where hooks should be run
func hasSystemImageChanged(namespace string, desiredImage string, client k8sclient.Client) (bool, error) {
	deployment := &k8sappsv1.Deployment{}
	err := client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: namespace,
		Name:      component.SystemAppDeploymentName,
	}, deployment)

	// Return error if can't get Deployment
	if err != nil && !k8serr.IsNotFound(err) {
		return false, fmt.Errorf("error getting deployment %s: %w", component.SystemAppDeploymentName, err)
	}

	// If the Deployment doesn't exist yet, this is a fresh install, not an upgrade
	if k8serr.IsNotFound(err) {
		return false, nil
	}

	// Check if the system-app container image has changed
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == "system-master" || container.Name == "system-provider" || container.Name == "system-developer" {
			if container.Image != desiredImage {
				return true, nil
			}
		}
	}

	return false, nil
}

// needAppHookJobRerun returns true if the system-app Deployment's revision doesn't match the Job's annotation tracking it
func needAppHookJobRerun(jName string, revision int64, namespace string, client k8sclient.Client) (bool, error) {
	trackedRevision, err := helper.GetAppRevision(jName, namespace, client)
	if err != nil {
		return false, err
	}

	// Job doesn't exist - we don't need a rerun (that would try to delete the job)
	if trackedRevision == 0 {
		return false, nil
	}

	// Job exists but has no annotation - rerun
	if trackedRevision == -1 {
		return true, nil
	}

	// Return true if the Deployment's version doesn't match the version tracked in the Job's annotation
	// This allows users to rerun hook jobs on demand
	return trackedRevision != revision, nil
}

func (r *SystemReconciler) systemAppDeploymentResourceMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	desiredName := helper.ObjectInfo(desired)
	update := false

	// Check containers
	if len(desired.Spec.Template.Spec.Containers) != 3 {
		return false, fmt.Errorf("%s desired spec.template.spec.containers length changed to '%d', should be 3", desiredName, len(desired.Spec.Template.Spec.Containers))
	}

	if len(existing.Spec.Template.Spec.Containers) != 3 {
		r.Logger().Info(fmt.Sprintf("%s spec.template.spec.containers length changed to '%d', recreating deployment", desiredName, len(existing.Spec.Template.Spec.Containers)))
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

// systemConfigMapMutator creates facilitates the creation of the ConfigMap on the first reconcile loop
// It also will update the endpoint in case zync is enabled|disabled while preserving all other values in the .data
func systemConfigMapMutator(existingObj, desiredObj k8sclient.Object) (bool, error) {
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

func (r *SystemReconciler) systemDatabaseTLSEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for TLS
	var changed bool

	for _, envVar := range []string{
		"DATABASE_SSL_CA",
		"DATABASE_SSL_CERT",
		"DATABASE_SSL_KEY",
		"DATABASE_SSL_MODE",
		"DB_SSL_CA",
		"DB_SSL_CERT",
		"DB_SSL_KEY",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func (r *SystemReconciler) systemRedisTLSEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	// Reconcile EnvVar only for Redis TLS for Porta - system-app and system-sidekiq
	var changed bool

	for _, envVar := range []string{
		"REDIS_CA_FILE",
		"REDIS_CLIENT_CERT",
		"REDIS_PRIVATE_KEY",
		"REDIS_SSL",
		"BACKEND_REDIS_CA_FILE",
		"BACKEND_REDIS_CLIENT_CERT",
		"BACKEND_REDIS_PRIVATE_KEY",
		"BACKEND_REDIS_SSL",
	} {
		tmpChanged := reconcilers.DeploymentEnvVarReconciler(desired, existing, envVar)
		changed = changed || tmpChanged
	}

	return changed, nil
}

func systemDeploymentVolumesMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNames := []string{
		"tls-secret",
		"writable-tls",
		"system-redis-tls",
		"backend-redis-tls",
		"s3-credentials",
	}

	return reconcilers.WeakDeploymentVolumesMutator(desired, existing, volumeNames)
}

func systemDeploymentInitContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeMountNames := []string{
		"tls-secret",
		"writable-tls",
	}

	return reconcilers.WeakDeploymentInitContainerVolumeMountsMutator(desired, existing, volumeMountNames)
}

func systemDeploymentContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeMountNames := []string{
		"s3-credentials",
		"system-redis-tls",
		"backend-redis-tls",
		"writable-tls",
	}
	return reconcilers.WeakDeploymentContainerVolumeMountsMutator(desired, existing, volumeMountNames)
}

func sidekiqDeploymentVolumesMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNames := []string{
		"s3-credentials",
		"tls-secret",
		"writable-tls",
		"system-redis-tls",
		"backend-redis-tls",
	}

	return reconcilers.WeakDeploymentVolumesMutator(desired, existing, volumeNames)
}

func sidekiqDeploymentInitContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeMountNames := []string{
		"tls-secret",
		"writable-tls",
		"system-redis-tls",
		"backend-redis-tls",
	}

	return reconcilers.WeakDeploymentInitContainerVolumeMountsMutator(desired, existing, volumeMountNames)
}

func sidekiqDeploymentContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeMountNames := []string{
		"s3-credentials",
		"system-redis-tls",
		"backend-redis-tls",
		"writable-tls",
	}
	return reconcilers.WeakDeploymentContainerVolumeMountsMutator(desired, existing, volumeMountNames)
}
