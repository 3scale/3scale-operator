package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SystemPostgreSQLReconciler struct {
	*BaseAPIManagerLogicReconciler
}

func NewSystemPostgreSQLReconciler(baseAPIManagerLogicReconciler *BaseAPIManagerLogicReconciler) DependencyReconciler {
	return &SystemPostgreSQLReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *SystemPostgreSQLReconciler) Reconcile() (reconcile.Result, error) {
	systemPostgreSQLImage, err := SystemPostgreSQLImage(r.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	systemPostgreSQL, err := SystemPostgreSQL(r.apiManager, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	// PostgreSQL Deployment
	deploymentMutator := reconcilers.DeploymentMutator(
		reconcilers.DeploymentContainerResourcesMutator,
		reconcilers.DeploymentAffinityMutator,
		reconcilers.DeploymentTolerationsMutator,
		reconcilers.DeploymentPodTemplateLabelsMutator,
		reconcilers.DeploymentPriorityClassMutator,
		reconcilers.DeploymentTopologySpreadConstraintsMutator,
		reconcilers.DeploymentPodTemplateAnnotationsMutator,
	)
	err = r.ReconcileDeployment(systemPostgreSQL.Deployment(systemPostgreSQLImage.Options.Image), deploymentMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// 3scale 2.14 -> 2.15
	// Overriding the Deployment health check because the postgresql-data PVC is ReadWriteOnce and so it can't be assigned across multiple nodes (pods)
	isMigrated, err := upgrade.MigrateDeploymentConfigToDeployment(component.SystemPostgreSQLDeploymentName, r.apiManager.GetNamespace(), true, r.Client(), nil)
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
	err = r.ReconcileService(systemPostgreSQL.Service(), reconcilers.ServiceMutator(serviceMutators...))
	if err != nil {
		return reconcile.Result{}, err
	}

	// PVC
	err = r.ReconcilePersistentVolumeClaim(systemPostgreSQL.DataPersistentVolumeClaim(), reconcilers.CreateOnlyMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	// DB secret
	err = r.ReconcileSecret(systemPostgreSQL.SystemDatabaseSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func SystemPostgreSQL(apimanager *appsv1alpha1.APIManager, client client.Client) (*component.SystemPostgreSQL, error) {
	optsProvider := NewSystemPostgresqlOptionsProvider(apimanager, apimanager.Namespace, client)
	opts, err := optsProvider.GetSystemPostgreSQLOptions()
	if err != nil {
		return nil, err
	}
	return component.NewSystemPostgreSQL(opts), nil
}
