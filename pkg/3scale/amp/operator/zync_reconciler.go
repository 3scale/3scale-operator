package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
	k8sappsv1 "k8s.io/api/apps/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ZyncReconciler struct {
	*BaseAPIManagerLogicReconciler

	ZyncEnabled bool
}

func NewZyncReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler, zyncEnabled bool) *ZyncReconciler {
	return &ZyncReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		ZyncEnabled:                   zyncEnabled,
	}
}

func (r *ZyncReconciler) Reconcile() (reconcile.Result, error) {
	ampImages, err := AmpImages(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	zync, err := Zync(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	if !r.ZyncEnabled {
		err := r.deleteZyncComponents(zync, ampImages)
		return reconcile.Result{}, err
	}

	// Zync Que Role
	err = r.ReconcileRole(zync.QueRole(), reconcilers.RoleRuleMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que SA
	err = r.ReconcileServiceAccount(zync.QueServiceAccount(), reconcilers.ServiceAccountImagePullPolicyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que RoleBinding
	err = r.ReconcileRoleBinding(zync.QueRoleBinding(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Deployment
	zyncMutators := []reconcilers.DMutateFn{
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
		reconcilers.DeploymentPodInitContainerImageMutator,
		reconcilers.DeploymentPodInitContainerMutator,
		zyncDatabaseTLSEnvVarMutator,
		zyncDeploymentVolumesMutator,
		zyncDeploymentInitContainerVolumeMountsMutator,
		zyncDeploymentContainerVolumeMountsMutator,
	}

	if r.apiManager.Spec.Zync.AppSpec.Replicas != nil {
		zyncMutators = append(zyncMutators, reconcilers.DeploymentReplicasMutator)
	}

	zyncDep, err := zync.Deployment(r.Context(), r.Client(), ampImages.Options.ZyncImage)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileDeployment(zyncDep, reconcilers.DeploymentMutator(zyncMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.ZyncName, r.apiManager.GetNamespace(), false, r.Client(), nil)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	// Zync Que Deployment
	zyncQueMutators := []reconcilers.DMutateFn{
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
		reconcilers.DeploymentPodInitContainerMutator,
		zyncDatabaseTLSEnvVarMutator,
		zyncDeploymentVolumesMutator,
		zyncDeploymentInitContainerVolumeMountsMutator,
		zyncDeploymentContainerVolumeMountsMutator,
	}
	if r.apiManager.Spec.Zync.QueSpec.Replicas != nil {
		zyncQueMutators = append(zyncQueMutators, reconcilers.DeploymentReplicasMutator)
	}

	zyncQueDep, err := zync.QueDeployment(r.Context(), r.Client(), ampImages.Options.ZyncImage)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileDeployment(zyncQueDep, reconcilers.DeploymentMutator(zyncQueMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	isMigrated, err = upgrade.MigrateDeploymentConfigToDeployment(component.ZyncQueDeploymentName, r.apiManager.GetNamespace(), false, r.Client(), r.BaseReconciler.Scheme())
	if err != nil {
		return reconcile.Result{}, err
	}
	if !isMigrated {
		return reconcile.Result{Requeue: true}, nil
	}

	serviceMutators := []reconcilers.MutateFn{
		reconcilers.CreateOnlyMutator,
		reconcilers.ServiceSelectorMutator,
	}

	// Zync Service
	err = r.ReconcileService(zync.Service(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	if !r.apiManager.IsExternal(appsv1alpha1.ZyncDatabase) {
		// Zync DB Deployment
		zyncDBDMutator := reconcilers.DeploymentMutator(
			reconcilers.DeploymentContainerResourcesMutator,
			reconcilers.DeploymentAffinityMutator,
			reconcilers.DeploymentTolerationsMutator,
			reconcilers.DeploymentPodTemplateLabelsMutator,
			reconcilers.DeploymentPriorityClassMutator,
			reconcilers.DeploymentTopologySpreadConstraintsMutator,
			reconcilers.DeploymentPodTemplateAnnotationsMutator,
			reconcilers.DeploymentPodContainerImageMutator,
			zyncDatabaseTLSEnvVarMutator,
		)
		err = r.ReconcileDeployment(zync.DatabaseDeployment(ampImages.Options.ZyncDatabasePostgreSQLImage), zyncDBDMutator)
		if err != nil {
			return reconcile.Result{}, err
		}

		// 3scale 2.14 -> 2.15
		isMigrated, err = upgrade.MigrateDeploymentConfigToDeployment(component.ZyncDatabaseDeploymentName, r.apiManager.GetNamespace(), false, r.Client(), nil)
		if err != nil {
			return reconcile.Result{}, err
		}
		if !isMigrated {
			return reconcile.Result{Requeue: true}, nil
		}

		// Zync DB Service
		err = r.ReconcileService(zync.DatabaseService(), reconcilers.ServiceMutator(serviceMutators...))
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Zync Secret
	err = r.ReconcileSecret(zync.Secret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync PDB
	err = r.ReconcilePodDisruptionBudget(zync.ZyncPodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que PDB
	err = r.ReconcilePodDisruptionBudget(zync.QuePodDisruptionBudget(), reconcilers.GenericPDBMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(zync.ZyncPodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePodMonitor(zync.ZyncQuePodMonitor(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	sumRate, err := helper.SumRateForOpenshiftVersion(r.Context(), r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcileGrafanaDashboards(zync.ZyncGrafanaV5Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ReconcileGrafanaDashboards(zync.ZyncGrafanaV4Dashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(zync.ZyncPrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ReconcilePrometheusRules(zync.ZyncQuePrometheusRules(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteZyncComponents handles the removal of all zync components
// The operator only enters this function when the APIManager .spec.zync.enabled was initially set to true but then changed to false
// When this happens, the operator will call this function on each reconcile loop to ensure the resources are cleaned up
func (r *ZyncReconciler) deleteZyncComponents(zync *component.Zync, ampImages *component.AmpImages) error {
	// ZyncQue PrometheusRules
	zyncQuePrometheusRules := zync.ZyncPrometheusRules()
	common.TagObjectToDelete(zyncQuePrometheusRules)
	err := r.ReconcilePrometheusRules(zyncQuePrometheusRules, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// Zync PrometheusRules
	zyncPrometheusRules := zync.ZyncPrometheusRules()
	common.TagObjectToDelete(zyncPrometheusRules)
	err = r.ReconcilePrometheusRules(zyncPrometheusRules, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// Zync GrafanaDashboards
	sumRate, err := helper.SumRateForOpenshiftVersion(r.Context(), r.Client())
	if err != nil {
		return err
	}
	zyncGrafanaV5Dashboard := zync.ZyncGrafanaV5Dashboard(sumRate)
	common.TagObjectToDelete(zyncGrafanaV5Dashboard)
	err = r.ReconcileGrafanaDashboards(zyncGrafanaV5Dashboard, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}
	zyncGrafanaV4Dashboard := zync.ZyncGrafanaV4Dashboard(sumRate)
	common.TagObjectToDelete(zyncGrafanaV4Dashboard)
	err = r.ReconcileGrafanaDashboards(zyncGrafanaV4Dashboard, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// Zync PodDisruptionBudget
	zyncPodDisruptionBudget := zync.ZyncPodDisruptionBudget()
	common.TagObjectToDelete(zyncPodDisruptionBudget)
	err = r.ReconcilePodDisruptionBudget(zyncPodDisruptionBudget, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// ZyncQue PodDisruptionBudget
	zyncQuePodDisruptionBudget := zync.QuePodDisruptionBudget()
	common.TagObjectToDelete(zyncQuePodDisruptionBudget)
	err = r.ReconcilePodDisruptionBudget(zyncQuePodDisruptionBudget, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// Zync PodMonitor
	zyncPodMonitor := zync.ZyncPodMonitor()
	common.TagObjectToDelete(zyncPodMonitor)
	err = r.ReconcilePodMonitor(zyncPodMonitor, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// ZyncQue PodMonitor
	zyncQuePodMonitor := zync.ZyncQuePodMonitor()
	common.TagObjectToDelete(zyncQuePodMonitor)
	err = r.ReconcilePodMonitor(zyncQuePodMonitor, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// Zync Secret
	zyncSecret := zync.Secret()
	common.TagObjectToDelete(zyncSecret)
	err = r.ReconcileSecret(zyncSecret, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	if !r.apiManager.IsExternal(appsv1alpha1.ZyncDatabase) {
		// ZyncDB Service
		zyncDBService := zync.DatabaseService()
		common.TagObjectToDelete(zyncDBService)
		err = r.ReconcileService(zyncDBService, reconcilers.DeleteOnlyMutator)
		if err != nil {
			return err
		}

		// ZyncDB Deployment
		zyncDBDeployment := zync.DatabaseDeployment(ampImages.Options.ZyncDatabasePostgreSQLImage)
		common.TagObjectToDelete(zyncDBDeployment)
		err = r.ReconcileDeployment(zyncDBDeployment, reconcilers.DeleteOnlyMutator)
		if err != nil {
			return err
		}
	}

	// Zync Service
	zyncService := zync.Service()
	common.TagObjectToDelete(zyncService)
	err = r.ReconcileService(zyncService, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// ZyncQue Deployment
	zyncQueDeployment, err := zync.QueDeployment(r.Context(), r.Client(), ampImages.Options.ZyncImage)
	if err != nil {
		return err
	}
	common.TagObjectToDelete(zyncQueDeployment)
	err = r.ReconcileDeployment(zyncQueDeployment, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// Zync Deployment
	zyncDeployment, err := zync.Deployment(r.Context(), r.Client(), ampImages.Options.ZyncImage)
	if err != nil {
		return err
	}
	common.TagObjectToDelete(zyncDeployment)
	err = r.ReconcileDeployment(zyncDeployment, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// ZyncQue RoleBinding
	zyncQueRoleBinding := zync.QueRoleBinding()
	common.TagObjectToDelete(zyncQueRoleBinding)
	err = r.ReconcileRoleBinding(zyncQueRoleBinding, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// ZyncQue ServiceAccount
	zyncQueServiceAccount := zync.QueServiceAccount()
	common.TagObjectToDelete(zyncQueServiceAccount)
	err = r.ReconcileServiceAccount(zyncQueServiceAccount, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	// ZyncQue Role
	zyncQueRole := zync.QueRole()
	common.TagObjectToDelete(zyncQueRole)
	err = r.ReconcileRole(zyncQueRole, reconcilers.DeleteOnlyMutator)
	if err != nil {
		return err
	}

	return nil
}

func Zync(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.Zync, error) {
	optsProvider := NewZyncOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetZyncOptions()
	if err != nil {
		return nil, err
	}
	return component.NewZync(opts), nil
}

func zyncDatabaseTLSEnvVarMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
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

func zyncDeploymentVolumesMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNames := []string{
		"tls-secret",
		"writable-tls",
	}

	return reconcilers.WeakDeploymentVolumesMutator(desired, existing, volumeNames)
}

func zyncDeploymentInitContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNames := []string{
		"tls-secret",
		"writable-tls",
	}

	return reconcilers.WeakDeploymentInitContainerVolumeMountsMutator(desired, existing, volumeNames)
}

func zyncDeploymentContainerVolumeMountsMutator(desired, existing *k8sappsv1.Deployment) (bool, error) {
	volumeNames := []string{
		"writable-tls",
	}
	return reconcilers.WeakDeploymentInitContainerVolumeMountsMutator(desired, existing, volumeNames)
}
