package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ZyncReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewZyncReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) *ZyncReconciler {
	return &ZyncReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *ZyncReconciler) Reconcile() (reconcile.Result, error) {
	zync, err := Zync(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que Role
	err = r.ReconcileRole(zync.QueRole(), reconcilers.CreateOnlyMutator)
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

	// Zync DC
	zyncMutators := []reconcilers.DCMutateFn{
		reconcilers.DeploymentConfigImageChangeTriggerMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		reconcilers.DeploymentConfigPodTemplateLabelsMutator,
		reconcilers.DeploymentConfigPriorityClassMutator,
		reconcilers.DeploymentConfigTopologySpreadConstraintsMutator,
		reconcilers.DeploymentConfigPodTemplateAnnotationsMutator,
	}
	if r.apiManager.Spec.Zync.AppSpec.Replicas != nil {
		zyncMutators = append(zyncMutators, reconcilers.DeploymentConfigReplicasMutator)
	}
	err = r.ReconcileDeploymentConfig(zync.DeploymentConfig(), reconcilers.DeploymentConfigMutator(zyncMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Que DC
	zyncQueMutators := []reconcilers.DCMutateFn{
		reconcilers.DeploymentConfigImageChangeTriggerMutator,
		reconcilers.DeploymentConfigContainerResourcesMutator,
		reconcilers.DeploymentConfigAffinityMutator,
		reconcilers.DeploymentConfigTolerationsMutator,
		reconcilers.DeploymentConfigPodTemplateLabelsMutator,
		reconcilers.DeploymentConfigPriorityClassMutator,
		reconcilers.DeploymentConfigTopologySpreadConstraintsMutator,
		reconcilers.DeploymentConfigPodTemplateAnnotationsMutator,
	}
	if r.apiManager.Spec.Zync.QueSpec.Replicas != nil {
		zyncQueMutators = append(zyncQueMutators, reconcilers.DeploymentConfigReplicasMutator)
	}
	err = r.ReconcileDeploymentConfig(zync.QueDeploymentConfig(), reconcilers.DeploymentConfigMutator(zyncQueMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Zync Service
	err = r.ReconcileService(zync.Service(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !r.apiManager.IsExternal(appsv1alpha1.ZyncDatabase) {
		// Zync DB DC
		zyncDBDCMutator := reconcilers.DeploymentConfigMutator(
			reconcilers.DeploymentConfigImageChangeTriggerMutator,
			reconcilers.DeploymentConfigContainerResourcesMutator,
			reconcilers.DeploymentConfigAffinityMutator,
			reconcilers.DeploymentConfigTolerationsMutator,
			reconcilers.DeploymentConfigPodTemplateLabelsMutator,
			reconcilers.DeploymentConfigPriorityClassMutator,
			reconcilers.DeploymentConfigTopologySpreadConstraintsMutator,
			reconcilers.DeploymentConfigPodTemplateAnnotationsMutator,
		)
		err = r.ReconcileDeploymentConfig(zync.DatabaseDeploymentConfig(), zyncDBDCMutator)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Zync DB Service
		err = r.ReconcileService(zync.DatabaseService(), reconcilers.CreateOnlyMutator)
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

	err = r.ReconcileGrafanaDashboard(zync.ZyncGrafanaDashboard(sumRate), reconcilers.GenericGrafanaDashboardsMutator)
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

	if len(r.apiManager.Status.Deployments.Starting) == 0 && len(r.apiManager.Status.Deployments.Stopped) == 0 && len(r.apiManager.Status.Deployments.Ready) > 0 {
		exist, err := r.routesExist()
		if err != nil {
			return reconcile.Result{}, err
		}
		if exist {
			return reconcile.Result{}, nil
		} else {
			// If the system-provider route does not exist at this point (i.e. when Deployments are ready)
			// we can force a resync of routes. see below for more details on why this is required:
			// https://access.redhat.com/documentation/en-us/red_hat_3scale_api_management/2.7/html/operating_3scale/backup-restore#creating_equivalent_zync_routes
			// This scenario will manifest during a backup and restore and also if the product ns was accidentally deleted.
			podName, err := r.findSystemSidekiqPod(r.apiManager)
			if err != nil {
				return reconcile.Result{}, err
			}
			if podName != "" {
				// Execute a resync of routes
				_, _, err := r.executeCommandOnPod("system-sidekiq", r.apiManager.Namespace, podName, []string{"bundle", "exec", "rake", "zync:resync:domains"})
				if err != nil {
					return reconcile.Result{}, err
				}
			}
		}
	}

	return reconcile.Result{}, nil
}

func Zync(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.Zync, error) {
	optsProvider := NewZyncOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetZyncOptions()
	if err != nil {
		return nil, err
	}
	return component.NewZync(opts), nil
}
