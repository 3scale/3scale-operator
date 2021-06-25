package operator

import (
	"context"
	"fmt"
	"reflect"

	commonapps "github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
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
	res, err := u.deleteAPIManagerOwnerReferencesFromSecrets()
	if err != nil {
		return res, fmt.Errorf("Deleting secrets APIManager owner references: %w", err)
	}
	if res.Requeue {
		return res, nil
	}

	res, err = u.upgradeSystemAppUserSessionTTL()
	if err != nil {
		return res, fmt.Errorf("Upgrading system app user session ttl: %w", err)
	}

	res, err = u.upgradeBackendRouteEnv()
	if err != nil {
		return res, fmt.Errorf("Upgrading backend route env vars: %w", err)
	}
	if res.Requeue {
		return res, nil
	}

	res, err = u.upgradeZyncPodTemplateAnnotations()
	if err != nil {
		return res, fmt.Errorf("Upgrading Zync DC PodTemplate: %w", err)
	}
	if res.Requeue {
		return res, nil
	}

	res, err = u.upgradeBackendCronDeploymentConfigMemoryLimits()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeImages()
	if err != nil {
		return res, fmt.Errorf("Upgrading images: %w", err)
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
	apicast, err := Apicast(u.apiManager, u.Client())
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

func (u *UpgradeApiManager) backendCronResourcesOverriden() bool {
	return u.apiManager.Spec.Backend != nil && u.apiManager.Spec.Backend.CronSpec != nil &&
		u.apiManager.Spec.Backend.CronSpec.Resources != nil
}

func (u *UpgradeApiManager) upgradeBackendCronDeploymentConfigMemoryLimits() (reconcile.Result, error) {
	backend, err := Backend(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}
	desired := backend.CronDeploymentConfig()

	// We only upgrade the memory limits when resource requirements flag is enabled
	// and the user has not intentionally overriden them at backend-cron level
	if (u.apiManager.Spec.ResourceRequirementsEnabled != nil && *u.apiManager.Spec.ResourceRequirementsEnabled) && !u.backendCronResourcesOverriden() {
		existing := &appsv1.DeploymentConfig{}
		err := u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
		if err != nil {
			return reconcile.Result{}, err
		}

		changed, err := u.ensureBackendCronMemoryLimit(desired, existing)
		if err != nil {
			return reconcile.Result{}, err
		}
		if changed {
			u.Logger().Info(fmt.Sprintf("Upgrading '%s' DeploymentConfig Resource Memory Limits", desired.Name))
			return reconcile.Result{Requeue: true}, u.UpdateResource(existing)
		}

	}
	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) ensureBackendCronMemoryLimit(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	if len(existing.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("Existing DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			existing.Name, len(existing.Spec.Template.Spec.Containers))
	}
	if len(desired.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("Desired DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			existing.Name, len(desired.Spec.Template.Spec.Containers))
	}

	existingResources := &existing.Spec.Template.Spec.Containers[0].Resources
	desiredResources := &desired.Spec.Template.Spec.Containers[0].Resources
	existingLimits := existingResources.Limits
	desiredLimits := desiredResources.Limits

	changed := false
	if existingLimits.Memory().Cmp(*desiredLimits.Memory()) != 0 {
		existing.Spec.Template.Spec.Containers[0].Resources.Limits[v1.ResourceMemory] = desiredLimits.Memory().DeepCopy()
		changed = true
	}
	return changed, nil
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
	desired := redis.BackendDeploymentConfig()

	existing := &appsv1.DeploymentConfig{}
	err = u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return reconcile.Result{}, err
	}

	changed := false
	tmpChanged, err := u.ensureDeploymentConfigImageChangeTrigger(desired, existing)
	if err != nil {
		return reconcile.Result{}, err
	}
	changed = changed || tmpChanged

	tmpChanged, err = u.ensureRedisCommand(desired, existing)
	if err != nil {
		return reconcile.Result{}, err
	}
	changed = changed || tmpChanged

	tmpChanged, err = u.ensureRedisPodTemplateLabels(desired, existing)
	if err != nil {
		return reconcile.Result{}, err
	}
	changed = changed || tmpChanged

	if changed {
		return reconcile.Result{Requeue: true}, u.UpdateResource(existing)
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) ensureRedisPodTemplateLabels(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	changed := false
	existingPodTemplateLabels := &existing.Spec.Template.Labels
	desiredPodTemplateLabels := desired.Spec.Template.Labels
	helper.MergeMapStringString(&changed, existingPodTemplateLabels, desiredPodTemplateLabels)
	if changed {
		u.Logger().V(1).Info(fmt.Sprintf("%s DeploymentConfig's PodTemplate labels changed", desired.Name))
	}

	return changed, nil
}

func (u *UpgradeApiManager) upgradeSystemRedisDeploymentConfig() (reconcile.Result, error) {
	redis, err := Redis(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}
	desired := redis.SystemDeploymentConfig()

	existing := &appsv1.DeploymentConfig{}
	err = u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return reconcile.Result{}, err
	}

	changed := false
	tmpChanged, err := u.ensureDeploymentConfigImageChangeTrigger(desired, existing)
	if err != nil {
		return reconcile.Result{}, err
	}
	changed = changed || tmpChanged

	tmpChanged, err = u.ensureRedisCommand(desired, existing)
	if err != nil {
		return reconcile.Result{}, err
	}
	changed = changed || tmpChanged

	tmpChanged, err = u.ensureRedisPodTemplateLabels(desired, existing)
	if err != nil {
		return reconcile.Result{}, err
	}
	changed = changed || tmpChanged

	if changed {
		return reconcile.Result{Requeue: true}, u.UpdateResource(existing)
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

func (u *UpgradeApiManager) ensureRedisCommand(desired, existing *appsv1.DeploymentConfig) (bool, error) {
	if len(existing.Spec.Template.Spec.Containers) == 0 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1", existing.Name, len(existing.Spec.Template.Spec.Containers))

	}
	if len(desired.Spec.Template.Spec.Containers) == 0 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1", desired.Name, len(desired.Spec.Template.Spec.Containers))
	}

	changed := false
	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].Command, desired.Spec.Template.Spec.Containers[0].Command) {
		existing.Spec.Template.Spec.Containers[0].Command = desired.Spec.Template.Spec.Containers[0].Command
		u.Logger().V(1).Info(fmt.Sprintf("DeploymentConfig %s container command changed", desired.Name))
		changed = true
	}

	return changed, nil
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

func (u *UpgradeApiManager) upgradeSystemAppUserSessionTTL() (reconcile.Result, error) {
	system, err := System(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(u.BaseReconciler, u.apiManager)

	// SystemApp Secret
	err = baseAPIManagerLogicReconciler.ReconcileSecret(system.AppSecret(), reconcilers.DefaultsOnlySecretMutator)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeSystemAppUserSessionTTLEnv(system.AppDeploymentConfig())
	return res, err
}

func (u *UpgradeApiManager) upgradeBackendRouteEnv() (reconcile.Result, error) {
	system, err := System(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.upgradeSystemAppBackendRouteEnv(system.AppDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeSidekiqBackendRouteEnv(system.SidekiqDeploymentConfig())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeSystemAppUserSessionTTLEnv(desired *appsv1.DeploymentConfig) (reconcile.Result, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 3 {
		return reconcile.Result{}, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 3",
			existing.Name, len(existing.Spec.Template.Spec.Containers))
	}
	desiredName := common.ObjectInfo(desired)

	update := false
	for idx := 0; idx < 3; idx++ {
		existingContainer := &existing.Spec.Template.Spec.Containers[idx]
		desiredContainer := &desired.Spec.Template.Spec.Containers[idx]
		desiredUserSessionTTLEnvVarIdx := helper.FindEnvVar(desiredContainer.Env, component.SystemSecretSystemAppUserSessionTTLFieldName)
		if desiredUserSessionTTLEnvVarIdx < 0 {
			return reconcile.Result{}, fmt.Errorf("%s desired spec.template.spec.containers env var '%s' does not exist", desiredName, component.SystemSecretSystemAppUserSessionTTLFieldName)
		}
		tmpUpdate := helper.EnsureEnvVar(desiredContainer.Env[desiredUserSessionTTLEnvVarIdx], &existingContainer.Env)
		update = update || tmpUpdate
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Upgrading USER_SESSION_TTL environment variable to DC %s", existing.Name))
		err = u.UpdateResource(existing)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{Requeue: update}, nil
}

func (u *UpgradeApiManager) upgradeSystemAppBackendRouteEnv(desired *appsv1.DeploymentConfig) (reconcile.Result, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 3 {
		return reconcile.Result{}, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 3",
			existing.Name, len(existing.Spec.Template.Spec.Containers))
	}

	desiredName := common.ObjectInfo(desired)

	update := false
	for idx := 0; idx < 3; idx++ {
		existingContainer := &existing.Spec.Template.Spec.Containers[idx]
		desiredContainer := &desired.Spec.Template.Spec.Containers[idx]
		desiredBackendRouteEnvVarIdx := helper.FindEnvVar(desiredContainer.Env, "BACKEND_ROUTE")
		if desiredBackendRouteEnvVarIdx < 0 {
			return reconcile.Result{}, fmt.Errorf("%s desired spec.template.spec.containers env var '%s' does not exist", desiredName, "BACKEND_ROUTE")
		}
		tmpUpdate := helper.EnsureEnvVar(desiredContainer.Env[desiredBackendRouteEnvVarIdx], &existingContainer.Env)
		update = update || tmpUpdate
	}

	// Pre hook pod env vars
	desiredBackendRouteEnvVarIdx := helper.FindEnvVar(desired.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env, "BACKEND_ROUTE")
	if desiredBackendRouteEnvVarIdx < 0 {
		return reconcile.Result{}, fmt.Errorf("%s desired spec.strategy.rollingparams.pre.execnewpod env var '%s' does not exist", desiredName, "BACKEND_ROUTE")
	}
	tmpUpdate := helper.EnsureEnvVar(desired.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env[desiredBackendRouteEnvVarIdx], &existing.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env)
	update = update || tmpUpdate

	if update {
		u.Logger().Info(fmt.Sprintf("Upgrading BACKEND_ROUTE environment variable to DC %s", existing.Name))
		err = u.UpdateResource(existing)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{Requeue: update}, nil
}

func (u *UpgradeApiManager) upgradeSidekiqBackendRouteEnv(desired *appsv1.DeploymentConfig) (reconcile.Result, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		return reconcile.Result{}, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			existing.Name, len(existing.Spec.Template.Spec.Containers))
	}

	desiredName := common.ObjectInfo(desired)

	existingContainer := &existing.Spec.Template.Spec.Containers[0]
	desiredContainer := &desired.Spec.Template.Spec.Containers[0]
	desiredBackendRouteEnvVarIdx := helper.FindEnvVar(desiredContainer.Env, "BACKEND_ROUTE")
	if desiredBackendRouteEnvVarIdx < 0 {
		return reconcile.Result{}, fmt.Errorf("%s desired spec.template.spec.containers env var '%s' does not exist", desiredName, "BACKEND_ROUTE")
	}
	update := helper.EnsureEnvVar(desiredContainer.Env[desiredBackendRouteEnvVarIdx], &existingContainer.Env)

	if update {
		u.Logger().Info(fmt.Sprintf("Upgrading BACKEND_ROUTE environment variable to DC %s", existing.Name))
		err = u.UpdateResource(existing)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{Requeue: update}, nil
}

func (u *UpgradeApiManager) upgradeZyncPodTemplateAnnotations() (reconcile.Result, error) {
	zync, err := Zync(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	desired := zync.DeploymentConfig()
	existing := &appsv1.DeploymentConfig{}
	err = u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return reconcile.Result{}, err
	}

	if existing.Spec.Template.Annotations == nil {
		existing.Spec.Template.Annotations = map[string]string{}
	}
	update := false

	for desiredAnnotationKey, desiredAnnotationVal := range desired.Spec.Template.Annotations {
		existingAnnotationVal, ok := existing.Spec.Template.Annotations[desiredAnnotationKey]
		if !ok || existingAnnotationVal != desiredAnnotationVal {
			existing.Spec.Template.Annotations[desiredAnnotationKey] = desiredAnnotationVal
			update = true
		}

		if existing.Annotations != nil {
			if _, ok := existing.Annotations[desiredAnnotationKey]; ok {
				delete(existing.Annotations, desiredAnnotationKey)
				update = true
			}
		}
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Upgrading zync DC %s PodTemplate annotations", existing.Name))
		err = u.UpdateResource(existing)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{Requeue: update}, nil
}

func (u *UpgradeApiManager) Logger() logr.Logger {
	return u.logger
}

func (u *UpgradeApiManager) deleteAPIManagerOwnerReferencesFromSecrets() (reconcile.Result, error) {
	secretsToUpdate := []string{
		component.BackendSecretInternalApiSecretName,
		component.BackendSecretBackendListenerSecretName,
		component.BackendSecretBackendRedisSecretName,
		component.SystemSecretSystemAppSecretName,
		component.SystemSecretSystemDatabaseSecretName,
		component.SystemSecretSystemEventsHookSecretName,
		component.SystemSecretSystemMasterApicastSecretName,
		component.SystemSecretSystemMemcachedSecretName,
		component.SystemSecretSystemRecaptchaSecretName,
		component.SystemSecretSystemRedisSecretName,
		component.SystemSecretSystemSeedSecretName,
		component.SystemSecretSystemSMTPSecretName,
		component.ZyncSecretName,
	}

	someChanged := false
	for _, secretName := range secretsToUpdate {
		existing := &v1.Secret{}
		err := u.Client().Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: u.apiManager.Namespace}, existing)
		if err != nil {
			return reconcile.Result{}, err
		}

		apimanagerOwnerReferenceIdx := -1
		for currIdx, secretOwnerReference := range existing.OwnerReferences {
			if secretOwnerReference.Kind == commonapps.APIManagerKind &&
				secretOwnerReference.Controller != nil && *secretOwnerReference.Controller {
				apimanagerOwnerReferenceIdx = currIdx
			}
		}

		if apimanagerOwnerReferenceIdx != -1 {
			u.Logger().Info(fmt.Sprintf("Removing APIManager OwnerReference from Secret %s", existing.Name))
			existing.OwnerReferences = append(existing.OwnerReferences[:apimanagerOwnerReferenceIdx], existing.OwnerReferences[apimanagerOwnerReferenceIdx+1:]...)
			err = u.UpdateResource(existing)
			if err != nil {
				return reconcile.Result{}, err
			}
			someChanged = true
		}
	}

	return reconcile.Result{Requeue: someChanged}, nil
}
