package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemMySQLReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemMySQLReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &SystemMySQLReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemMySQLReconciler) Reconcile() (reconcile.Result, error) {
	systemMySQLImage, err := SystemMySQLImage(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	systemMySQL, err := SystemMySQL(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// MySQL Deployment
	deploymentMutator := reconcilers.DeploymentMutator(
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
		reconcilers.DeploymentPodContainerImageMutator,
	)
	err = r.ReconcileDeployment(systemMySQL.Deployment(systemMySQLImage.Options.Image), deploymentMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	// Overriding the Deployment health check because the mysql-storage PVC is ReadWriteOnce and so it can't be assigned across multiple nodes (pods)
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.SystemMySQLDeploymentName, r.apiManager.GetNamespace(), true, r.Client(), nil)
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

	// Service
	err = r.ReconcileService(systemMySQL.Service(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// Main CM
	err = r.ReconcileConfigMap(systemMySQL.MainConfigConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Extra CM
	err = r.ReconcileConfigMap(systemMySQL.ExtraConfigConfigMap(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// PVC
	err = r.ReconcilePersistentVolumeClaim(systemMySQL.PersistentVolumeClaim(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Secret
	err = r.ReconcileSecret(systemMySQL.SystemDatabaseSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func SystemMySQL(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.SystemMysql, error) {
	optsProvider := NewSystemMysqlOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetMysqlOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemMysql(opts), nil
}
