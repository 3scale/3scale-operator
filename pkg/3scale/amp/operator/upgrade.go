package operator

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpgradeApiManager struct {
	*reconcilers.BaseReconciler
	apiManager *appsv1alpha1.APIManager
	logger     logr.Logger
}

func NewUpgradeApiManager(b *reconcilers.BaseReconciler, apiManager *appsv1alpha1.APIManager) *UpgradeApiManager {
	return &UpgradeApiManager{
		BaseReconciler: b,
		apiManager:     apiManager,
		logger:         b.Logger().WithValues("APIManager Upgrade Controller", apiManager.Name),
	}
}

func (u *UpgradeApiManager) Upgrade() (reconcile.Result, error) {
	res, err := u.upgradeImages()
	if err != nil {
		return res, fmt.Errorf("Upgrading images: %w", err)
	}
	if res.Requeue {
		return res, nil
	}

	res, err = u.upgradeMonitoringSettings()
	if err != nil {
		return res, fmt.Errorf("Upgrading monitoring settings: %w", err)
	}
	if res.Requeue {
		return res, nil
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeImages() (reconcile.Result, error) {
	res, err := u.upgradeAMPImageStreams()
	if res.Requeue || err != nil {
		return res, err
	}

	if !u.apiManager.IsExternalDatabaseEnabled() {
		res, err = u.upgradeBackendRedisImageStream()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemRedisImageStream()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemDatabaseImageStream()
		if res.Requeue || err != nil {
			return res, err
		}
	}

	res, err = u.upgradeDeploymentConfigs()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeDeploymentConfigs() (reconcile.Result, error) {
	res, err := u.upgradeAPIcastDeploymentConfigs()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeBackendDeploymentConfigs()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeZyncDeploymentConfigs()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeMemcachedDeploymentConfig()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeSystemDeploymentConfigs()
	if res.Requeue || err != nil {
		return res, err
	}

	if !u.apiManager.IsExternalDatabaseEnabled() {
		res, err = u.upgradeBackendRedisDeploymentConfig()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemRedisDeploymentConfig()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemDatabaseDeploymentConfig()
		if res.Requeue || err != nil {
			return res, err
		}
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeAPIcastDeploymentConfigs() (reconcile.Result, error) {
	apicast, err := Apicast(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(apicast.StagingDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeDeploymentConfigImageChangeTrigger(apicast.ProductionDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeBackendDeploymentConfigs() (reconcile.Result, error) {
	backend, err := Backend(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(backend.ListenerDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeDeploymentConfigImageChangeTrigger(backend.WorkerDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeDeploymentConfigImageChangeTrigger(backend.CronDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeZyncDeploymentConfigs() (reconcile.Result, error) {
	zync, err := Zync(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(zync.DeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeDeploymentConfigImageChangeTrigger(zync.QueDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeDeploymentConfigImageChangeTrigger(zync.DatabaseDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeMemcachedDeploymentConfig() (reconcile.Result, error) {
	memcached, err := Memcached(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(memcached.DeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}
func (u *UpgradeApiManager) upgradeBackendRedisDeploymentConfig() (reconcile.Result, error) {
	redis, err := Redis(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(redis.BackendDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeSystemRedisDeploymentConfig() (reconcile.Result, error) {
	redis, err := Redis(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(redis.SystemDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeSystemDatabaseDeploymentConfig() (reconcile.Result, error) {
	if u.apiManager.Spec.System.DatabaseSpec != nil && u.apiManager.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.upgradeSystemPostgreSQLDeploymentConfig()
	}

	// default is MySQL
	return u.upgradeSystemMySQLDeploymentConfig()
}

func (u *UpgradeApiManager) upgradeSystemPostgreSQLDeploymentConfig() (reconcile.Result, error) {
	systemPostgreSQL, err := SystemPostgreSQL(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(systemPostgreSQL.DeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeSystemMySQLDeploymentConfig() (reconcile.Result, error) {
	systemMySQL, err := SystemMySQL(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(systemMySQL.DeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeSystemDeploymentConfigs() (reconcile.Result, error) {
	system, err := System(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeDeploymentConfigImageChangeTrigger(system.AppDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeDeploymentConfigImageChangeTrigger(system.SidekiqDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeDeploymentConfigImageChangeTrigger(system.SphinxDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeDeploymentConfigImageChangeTrigger(desired *appsv1.DeploymentConfig) (reconcile.Result, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return reconcile.Result{}, err
	}

	changed, err := u.ensureDeploymentConfigImageChangeTrigger(desired, existing)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changed {
		return reconcile.Result{Requeue: true}, u.UpdateResource(existing)
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) ensureDeploymentConfigImageChangeTrigger(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	desiredDeploymentTriggerImageChangePos, err := u.findDeploymentTriggerOnImageChange(desired.Spec.Triggers)
	if err != nil {
		return false, fmt.Errorf("unexpected: '%s' in DeploymentConfig '%s'", err, desired.Name)

	}
	existingDeploymentTriggerImageChangePos, err := u.findDeploymentTriggerOnImageChange(existing.Spec.Triggers)
	if err != nil {
		return false, fmt.Errorf("unexpected: '%s' in DeploymentConfig '%s'", err, existing.Name)
	}

	desiredDeploymentTriggerImageChangeParams := desired.Spec.Triggers[desiredDeploymentTriggerImageChangePos].ImageChangeParams
	existingDeploymentTriggerImageChangeParams := existing.Spec.Triggers[existingDeploymentTriggerImageChangePos].ImageChangeParams

	if !reflect.DeepEqual(existingDeploymentTriggerImageChangeParams.From.Name, desiredDeploymentTriggerImageChangeParams.From.Name) {
		diff := cmp.Diff(existingDeploymentTriggerImageChangeParams.From.Name, desiredDeploymentTriggerImageChangeParams.From.Name)
		u.Logger().V(1).Info(fmt.Sprintf("%s ImageStream tag name in imageChangeParams trigger changed: %s", desired.Name, diff))
		existingDeploymentTriggerImageChangeParams.From.Name = desiredDeploymentTriggerImageChangeParams.From.Name
		return true, nil
	}

	return false, nil
}

func (u *UpgradeApiManager) upgradeAMPImageStreams() (reconcile.Result, error) {
	// implement upgrade procedure by reconcile procedure
	reconciler := NewAMPImagesReconciler(NewBaseAPIManagerLogicReconciler(u.BaseReconciler, u.apiManager))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeBackendRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	reconciler := NewBaseAPIManagerLogicReconciler(u.BaseReconciler, u.apiManager)
	return reconcile.Result{}, reconciler.ReconcileImagestream(redis.BackendImageStream(), reconcilers.GenericImageStreamMutator)
}

func (u *UpgradeApiManager) upgradeSystemRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	reconciler := NewBaseAPIManagerLogicReconciler(u.BaseReconciler, u.apiManager)
	return reconcile.Result{}, reconciler.ReconcileImagestream(redis.SystemImageStream(), reconcilers.GenericImageStreamMutator)
}

func (u *UpgradeApiManager) upgradeSystemDatabaseImageStream() (reconcile.Result, error) {
	if u.apiManager.Spec.System.DatabaseSpec != nil && u.apiManager.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.upgradeSystemPostgreSQLImageStream()
	}

	// default is MySQL
	return u.upgradeSystemMySQLImageStream()
}

func (u *UpgradeApiManager) upgradeSystemMySQLImageStream() (reconcile.Result, error) {
	// implement upgrade procedure by reconcile procedure
	reconciler := NewSystemMySQLImageReconciler(NewBaseAPIManagerLogicReconciler(u.BaseReconciler, u.apiManager))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeSystemPostgreSQLImageStream() (reconcile.Result, error) {
	// implement upgrade procedure by reconcile procedure
	reconciler := NewSystemPostgreSQLImageReconciler(NewBaseAPIManagerLogicReconciler(u.BaseReconciler, u.apiManager))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) findDeploymentTriggerOnImageChange(triggerPolicies []appsv1.DeploymentTriggerPolicy) (int, error) {
	result := -1
	for i := range triggerPolicies {
		if triggerPolicies[i].Type == appsv1.DeploymentTriggerOnImageChange {
			if result != -1 {
				return -1, fmt.Errorf("found more than one imageChangeParams Deployment trigger policy")
			}
			result = i
		}
	}

	if result == -1 {
		return -1, fmt.Errorf("no imageChangeParams deployment trigger policy found")
	}

	return result, nil
}

func (u *UpgradeApiManager) upgradeMonitoringSettings() (reconcile.Result, error) {
	updated := false

	// Port and environment variable are exposed in DC for monitoring
	updatedTmp, err := u.ensureSystemAppMonitoringSettings()
	if err != nil {
		return reconcile.Result{}, err
	}
	updated = updated || updatedTmp

	// We explicitely set the metrics environment variable in system-sidekiq
	// for clarity
	updatedTmp, err = u.ensureSystemSidekiqMonitoringSettings()
	if err != nil {
		return reconcile.Result{}, err
	}
	updated = updated || updatedTmp

	return reconcile.Result{Requeue: updated}, nil
}

func (u *UpgradeApiManager) ensureSystemSidekiqMonitoringSettings() (bool, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: component.SystemSidekiqName, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return false, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			component.SystemSidekiqName, len(existing.Spec.Template.Spec.Containers))
	}

	update := false
	if _, ok := helper.FindEnvVar(existing.Spec.Template.Spec.Containers[0].Env, component.SystemSidekiqPrometheusExporterPortEnvVarName); !ok {
		existing.Spec.Template.Spec.Containers[0].Env = append(
			existing.Spec.Template.Spec.Containers[0].Env,
			helper.EnvVarFromValue(component.SystemSidekiqPrometheusExporterPortEnvVarName, strconv.Itoa(component.SystemSidekiqMetricsPort)),
		)
		update = true
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Adding prometheus metrics environment variable to DC %s", component.SystemSidekiqName))
		err = u.UpdateResource(existing)
		if err != nil {
			return false, err
		}
	}

	return update, nil
}

func (u *UpgradeApiManager) ensureSystemAppMonitoringSettings() (bool, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: component.SystemAppDeploymentName, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return false, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 3 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 3",
			existing.Name, len(existing.Spec.Template.Spec.Containers))
	}

	update := false
	for idx := range existing.Spec.Template.Spec.Containers {
		container := &existing.Spec.Template.Spec.Containers[idx]
		var containerPrometheusMetricsPort int
		var containerPrometheusMetricsPortName string
		switch containerName := container.Name; {
		case containerName == component.SystemAppMasterContainerName:
			containerPrometheusMetricsPort = component.SystemAppMasterContainerPrometheusPort
			containerPrometheusMetricsPortName = component.SystemAppMasterContainerMetricsPortName
		case containerName == component.SystemAppProviderContainerName:
			containerPrometheusMetricsPort = component.SystemAppProviderContainerPrometheusPort
			containerPrometheusMetricsPortName = component.SystemAppProviderContainerMetricsPortName
		case containerName == component.SystemAppDeveloperContainerName:
			containerPrometheusMetricsPort = component.SystemAppDeveloperContainerPrometheusPort
			containerPrometheusMetricsPortName = component.SystemAppDeveloperContainerMetricsPortName
		default:
			return false, fmt.Errorf("DeploymentConfig '%s' has unrecognized container name '%s'", existing.Name, containerName)
		}

		if _, ok := helper.FindEnvVar(container.Env, component.SystemAppPrometheusExporterPortEnvVarName); !ok {
			container.Env = append(
				container.Env,
				helper.EnvVarFromValue(component.SystemAppPrometheusExporterPortEnvVarName, strconv.Itoa(containerPrometheusMetricsPort)),
			)
			update = true
		}
		if _, ok := helper.FindContainerPortByName(container.Ports, containerPrometheusMetricsPortName); !ok {
			container.Ports = append(container.Ports, v1.ContainerPort{Name: containerPrometheusMetricsPortName, ContainerPort: int32(containerPrometheusMetricsPort), Protocol: v1.ProtocolTCP})
			update = true
		}
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Adding prometheus metrics environment variable and ports to DC %s", existing.Name))
		err = u.UpdateResource(existing)
		if err != nil {
			return false, err
		}
	}

	return update, nil
}

func (u *UpgradeApiManager) Logger() logr.Logger {
	return u.logger
}
