package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	k8sappsv1 "k8s.io/api/apps/v1"
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

	err = r.ReconcilePodMonitor(system.SystemAppPodMonitor(), reconcilers.CreateOnlyMutator)
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

func System(cr *appsv1alpha1.APIManager, client k8sclient.Client) (*component.System, error) {
	optsProvider := NewSystemOptionsProvider(cr, cr.Namespace, client)
	opts, err := optsProvider.GetSystemOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystem(opts), nil
}
