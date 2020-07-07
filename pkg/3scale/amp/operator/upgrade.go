package operator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	// TODO Upgrade should be grouped by component and not by feature.
	// Otherwise each component may get multiple update requests.

	res, err := u.upgradeImages()
	if err != nil {
		return res, fmt.Errorf("Upgrading images: %w", err)
	}
	if res.Requeue {
		return res, nil
	}

	// upgrade system-master-apicast secret
	res, err = u.upgradeSystemMasterApicastSecret()
	if err != nil {
		return res, fmt.Errorf("Upgrading system master apicast secret: %w", err)
	}
	if res.Requeue {
		return res, nil
	}

	res, err = u.upgradePodTemplateLabels()
	if err != nil {
		return res, fmt.Errorf("Upgrading pod template labels: %w", err)
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

	res, err = u.deleteOldImageStreamsTags()
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
	redis, err := Redis(u.apiManager)
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
	redis, err := Redis(u.apiManager)
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
	redis, err := Redis(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	reconciler := NewBaseAPIManagerLogicReconciler(u.BaseReconciler, u.apiManager)
	return reconcile.Result{}, reconciler.ReconcileImagestream(redis.BackendImageStream(), reconcilers.GenericImageStreamMutator)
}

func (u *UpgradeApiManager) upgradeSystemRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.apiManager)
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

func (u *UpgradeApiManager) deleteOldImageStreamsTags() (reconcile.Result, error) {
	res, err := u.deleteAmpOldImageStreamsTags()
	if res.Requeue || err != nil {
		return res, err
	}

	if !u.apiManager.IsExternalDatabaseEnabled() {
		res, err = u.deleteBackendRedisOldImageStreamTags()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.deleteSystemRedisOldImageStreamTags()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.deleteSystemDatabaseOldImageStreamTags()
		if res.Requeue || err != nil {
			return res, err
		}
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) deleteAmpOldImageStreamsTags() (reconcile.Result, error) {
	ampimages, err := AmpImages(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.deleteOldImageStreamTags(ampimages.APICastImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.deleteOldImageStreamTags(ampimages.BackendImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.deleteOldImageStreamTags(ampimages.ZyncImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.deleteOldImageStreamTags(ampimages.ZyncDatabasePostgreSQLImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.deleteOldImageStreamTags(ampimages.SystemImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.deleteOldImageStreamTags(ampimages.SystemMemcachedImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) deleteBackendRedisOldImageStreamTags() (reconcile.Result, error) {
	redis, err := Redis(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.deleteOldImageStreamTags(redis.BackendImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) deleteSystemRedisOldImageStreamTags() (reconcile.Result, error) {
	redis, err := Redis(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}

	res, err := u.deleteOldImageStreamTags(redis.SystemImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) deleteSystemDatabaseOldImageStreamTags() (reconcile.Result, error) {
	if u.apiManager.Spec.System.DatabaseSpec != nil && u.apiManager.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.deleteSystemPostgreSQLOldImageStreamTags()
	}

	// default is MySQL
	return u.deleteSystemMySQLOldImageStreamTags()
}

func (u *UpgradeApiManager) deleteSystemPostgreSQLOldImageStreamTags() (reconcile.Result, error) {
	postgresqlImage, err := SystemPostgreSQLImage(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}
	res, err := u.deleteOldImageStreamTags(postgresqlImage.ImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}
	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) deleteSystemMySQLOldImageStreamTags() (reconcile.Result, error) {
	mysqlImage, err := SystemMySQLImage(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}
	res, err := u.deleteOldImageStreamTags(mysqlImage.ImageStream().GetName())
	if res.Requeue || err != nil {
		return res, err
	}
	return reconcile.Result{}, err
}

func (u *UpgradeApiManager) deleteOldImageStreamTags(imageStreamName string) (reconcile.Result, error) {
	existingImageStream := &imagev1.ImageStream{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: imageStreamName, Namespace: u.apiManager.Namespace}, existingImageStream)
	if err != nil {
		return reconcile.Result{}, err
	}

	latestTagName := "latest"
	return u.deleteImageStreamTag(latestTagName, existingImageStream)
}

// deleteImageStreamTag deletes the corresponding ImageStreamTag object
// in case the tag is found in the Spec definition of the existing ImageStream
// object.
// Deleting the tag element directly from the Spec section of the ImageStream
// does not completely remove the tag: The ImageStream object still has the
// removed tag in the status section of the object and the ImageStreamTag
// object still exists.
// Instead of doing that, this method searches for a tag in Spec section of the
// existing ImageStream object and if it exists then it tries to delete the
// corresponding ImageStreamTag object.
// When deleting the ImageStreamTag object the tag is automatically removed
// from the Spec and Status sections of the corresponding
// ImageStream object
func (u *UpgradeApiManager) deleteImageStreamTag(tagRefName string, existing *imagev1.ImageStream) (reconcile.Result, error) {
	pos := u.findTagReference(tagRefName, existing.Spec.Tags)
	if pos != -1 {
		existingIsTag := &imagev1.ImageStreamTag{}
		// We use ApiClientReader instead of the Client due to
		// The operator-sdk Client automatically performs a Watch on all the objects
		// That are obtained with Get, but the ImageStreamTag kind does not have
		// the Watch verb, which caused errors.
		err := u.APIClientReader().Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s:%s", existing.Name, tagRefName), Namespace: u.apiManager.Namespace}, existingIsTag)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = u.DeleteResource(existingIsTag)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) findTagReference(tagRefName string, tagRefs []imagev1.TagReference) int {
	for i := range tagRefs {
		if tagRefs[i].Name == tagRefName {
			return i
		}
	}
	return -1
}

func (u *UpgradeApiManager) upgradeSystemMasterApicastSecret() (reconcile.Result, error) {
	existing := &v1.Secret{}
	secretNamespacedName := types.NamespacedName{
		Name:      component.SystemSecretSystemMasterApicastSecretName,
		Namespace: u.apiManager.Namespace,
	}
	err := u.Client().Get(context.TODO(), secretNamespacedName, existing)
	// NotFound also regarded as error, as secret is expected to exist
	if err != nil {
		return reconcile.Result{}, err
	}

	if _, ok := existing.Data["BASE_URL"]; ok {
		// Remove unused BASE_URL field
		patchJSON := []byte(`[{"op": "remove", "path": "/data/BASE_URL"}]`)
		// Apply JSON patch https://tools.ietf.org/html/rfc6902
		patch := client.ConstantPatch(types.JSONPatchType, patchJSON)
		err = u.Client().Patch(context.TODO(), existing, patch)
		if err != nil {
			return reconcile.Result{}, err
		}
		u.Logger().Info(fmt.Sprintf("Upgrade: patch object %s", common.ObjectInfo(existing)))
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradePodTemplateLabels() (reconcile.Result, error) {
	apicast, err := Apicast(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}
	backend, err := Backend(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}
	zync, err := Zync(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}
	memcached, err := Memcached(u.apiManager)
	if err != nil {
		return reconcile.Result{}, err
	}
	system, err := System(u.apiManager, u.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	deploymentConfigs := []*appsv1.DeploymentConfig{
		apicast.StagingDeploymentConfig(),
		apicast.ProductionDeploymentConfig(),
		backend.ListenerDeploymentConfig(),
		backend.WorkerDeploymentConfig(),
		backend.CronDeploymentConfig(),
		zync.DeploymentConfig(),
		zync.QueDeploymentConfig(),
		zync.DatabaseDeploymentConfig(),
		memcached.DeploymentConfig(),
		system.AppDeploymentConfig(),
		system.SidekiqDeploymentConfig(),
		system.SphinxDeploymentConfig(),
	}

	if !u.apiManager.IsExternalDatabaseEnabled() {
		redis, err := Redis(u.apiManager)
		if err != nil {
			return reconcile.Result{}, err
		}
		deploymentConfigs = append(deploymentConfigs, redis.SystemDeploymentConfig())
		deploymentConfigs = append(deploymentConfigs, redis.BackendDeploymentConfig())
	}

	if u.apiManager.IsSystemPostgreSQLEnabled() {
		systemPostgreSQL, err := SystemPostgreSQL(u.apiManager, u.Client())
		if err != nil {
			return reconcile.Result{}, err
		}
		deploymentConfigs = append(deploymentConfigs, systemPostgreSQL.DeploymentConfig())
	}

	if u.apiManager.IsSystemMysqlEnabled() {
		systemMySQL, err := SystemMySQL(u.apiManager, u.Client())
		if err != nil {
			return reconcile.Result{}, err
		}
		deploymentConfigs = append(deploymentConfigs, systemMySQL.DeploymentConfig())
	}

	updated := false
	for _, desired := range deploymentConfigs {
		updatedTmp, err := u.ensurePodTemplateLabels(desired)
		if err != nil {
			return reconcile.Result{}, err
		}
		updated = updated || updatedTmp
	}

	return reconcile.Result{Requeue: updated}, nil
}

func (u *UpgradeApiManager) ensurePodTemplateLabels(desired *appsv1.DeploymentConfig) (bool, error) {
	u.Logger().V(1).Info(fmt.Sprintf("ensurePodTemplateLabels object %s", common.ObjectInfo(desired)))
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return false, err
	}

	updated := false

	diff := cmp.Diff(existing.Spec.Template.Labels, desired.Spec.Template.Labels)
	helper.MergeMapStringString(&updated, &existing.Spec.Template.Labels, desired.Spec.Template.Labels)

	if updated {
		u.Logger().V(1).Info(fmt.Sprintf("DC %s template lables changed: %s", desired.Name, diff))
		err = u.UpdateResource(existing)
		if err != nil {
			return false, err
		}
	}

	return updated, nil
}

func (u *UpgradeApiManager) upgradeMonitoringSettings() (reconcile.Result, error) {
	updated := false

	updatedTmp, err := u.ensureApicastStagingMonitoringSettings()
	if err != nil {
		return reconcile.Result{}, err
	}
	updated = updated || updatedTmp

	updatedTmp, err = u.ensureApicastProductionMonitoringSettings()
	if err != nil {
		return reconcile.Result{}, err
	}
	updated = updated || updatedTmp

	updatedTmp, err = u.ensureBackendWorkerMonitoringSettings()
	if err != nil {
		return reconcile.Result{}, err
	}
	updated = updated || updatedTmp

	updatedTmp, err = u.ensureBackendListenerMonitoringSettings()
	if err != nil {
		return reconcile.Result{}, err
	}
	updated = updated || updatedTmp

	// Ports need to be exposed in DC for monitoring
	updatedTmp, err = u.ensureZyncMonitoringSettings()
	if err != nil {
		return reconcile.Result{}, err
	}
	updated = updated || updatedTmp

	// Ports need to be exposed in DC for monitoring
	updatedTmp, err = u.ensureSystemSidekiqMonitoringSettings()
	if err != nil {
		return reconcile.Result{}, err
	}
	updated = updated || updatedTmp

	return reconcile.Result{Requeue: updated}, nil
}

func (u *UpgradeApiManager) ensureApicastStagingMonitoringSettings() (bool, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: component.ApicastStagingName, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return false, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			component.ApicastStagingName, len(existing.Spec.Template.Spec.Containers))
	}

	update := false
	if _, ok := helper.FindEnvVar(existing.Spec.Template.Spec.Containers[0].Env, "APICAST_EXTENDED_METRICS"); !ok {
		existing.Spec.Template.Spec.Containers[0].Env = append(existing.Spec.Template.Spec.Containers[0].Env, helper.EnvVarFromValue("APICAST_EXTENDED_METRICS", "true"))
		update = true
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Enabling metrics to DC %s", component.ApicastStagingName))
		err = u.UpdateResource(existing)
		if err != nil {
			return false, err
		}
	}

	return update, nil
}

func (u *UpgradeApiManager) ensureApicastProductionMonitoringSettings() (bool, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: component.ApicastProductionName, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return false, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			component.ApicastProductionName, len(existing.Spec.Template.Spec.Containers))
	}

	update := false
	if _, ok := helper.FindEnvVar(existing.Spec.Template.Spec.Containers[0].Env, "APICAST_EXTENDED_METRICS"); !ok {
		existing.Spec.Template.Spec.Containers[0].Env = append(existing.Spec.Template.Spec.Containers[0].Env, helper.EnvVarFromValue("APICAST_EXTENDED_METRICS", "true"))
		update = true
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Enabling metrics to DC %s", component.ApicastProductionName))
		err = u.UpdateResource(existing)
		if err != nil {
			return false, err
		}
	}

	return update, nil
}

func (u *UpgradeApiManager) ensureBackendWorkerMonitoringSettings() (bool, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: component.BackendWorkerName, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return false, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			component.BackendWorkerName, len(existing.Spec.Template.Spec.Containers))
	}

	update := false
	if _, ok := helper.FindEnvVar(existing.Spec.Template.Spec.Containers[0].Env, "CONFIG_WORKER_PROMETHEUS_METRICS_PORT"); !ok {
		existing.Spec.Template.Spec.Containers[0].Env = append(existing.Spec.Template.Spec.Containers[0].Env, helper.EnvVarFromValue("CONFIG_WORKER_PROMETHEUS_METRICS_PORT", component.BackendWorkerMetricsPortStr))
		update = true
	}

	if _, ok := helper.FindEnvVar(existing.Spec.Template.Spec.Containers[0].Env, "CONFIG_WORKER_PROMETHEUS_METRICS_ENABLED"); !ok {
		existing.Spec.Template.Spec.Containers[0].Env = append(existing.Spec.Template.Spec.Containers[0].Env, helper.EnvVarFromValue("CONFIG_WORKER_PROMETHEUS_METRICS_ENABLED", "true"))
		update = true
	}

	if _, ok := helper.FindContainerPortByName(existing.Spec.Template.Spec.Containers[0].Ports, "metrics"); !ok {
		existing.Spec.Template.Spec.Containers[0].Ports = append(existing.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{Name: "metrics", ContainerPort: component.BackendWorkerMetricsPort, Protocol: v1.ProtocolTCP})
		update = true
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Enabling metrics to DC %s", component.BackendWorkerName))
		err = u.UpdateResource(existing)
		if err != nil {
			return false, err
		}
	}

	return update, nil
}

func (u *UpgradeApiManager) ensureBackendListenerMonitoringSettings() (bool, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: component.BackendListenerName, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return false, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			component.BackendListenerName, len(existing.Spec.Template.Spec.Containers))
	}

	update := false
	if _, ok := helper.FindEnvVar(existing.Spec.Template.Spec.Containers[0].Env, "CONFIG_LISTENER_PROMETHEUS_METRICS_PORT"); !ok {
		existing.Spec.Template.Spec.Containers[0].Env = append(existing.Spec.Template.Spec.Containers[0].Env, helper.EnvVarFromValue("CONFIG_LISTENER_PROMETHEUS_METRICS_PORT", component.BackendListenerMetricsPortStr))
		update = true
	}

	if _, ok := helper.FindEnvVar(existing.Spec.Template.Spec.Containers[0].Env, "CONFIG_LISTENER_PROMETHEUS_METRICS_ENABLED"); !ok {
		existing.Spec.Template.Spec.Containers[0].Env = append(existing.Spec.Template.Spec.Containers[0].Env, helper.EnvVarFromValue("CONFIG_LISTENER_PROMETHEUS_METRICS_ENABLED", "true"))
		update = true
	}

	if _, ok := helper.FindContainerPortByName(existing.Spec.Template.Spec.Containers[0].Ports, "metrics"); !ok {
		existing.Spec.Template.Spec.Containers[0].Ports = append(existing.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{Name: "metrics", ContainerPort: component.BackendListenerMetricsPort, Protocol: v1.ProtocolTCP})
		update = true
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Enabling metrics to DC %s", component.BackendListenerName))
		err = u.UpdateResource(existing)
		if err != nil {
			return false, err
		}
	}

	return update, nil
}

func (u *UpgradeApiManager) ensureZyncMonitoringSettings() (bool, error) {
	existing := &appsv1.DeploymentConfig{}
	err := u.Client().Get(context.TODO(), types.NamespacedName{Name: component.ZyncName, Namespace: u.apiManager.Namespace}, existing)
	if err != nil {
		return false, err
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		return false, fmt.Errorf("DeploymentConfig %s spec.template.spec.containers length is %d, should be 1",
			component.ZyncName, len(existing.Spec.Template.Spec.Containers))
	}

	update := false
	if _, ok := helper.FindContainerPortByName(existing.Spec.Template.Spec.Containers[0].Ports, "metrics"); !ok {
		existing.Spec.Template.Spec.Containers[0].Ports = append(existing.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{Name: "metrics", ContainerPort: component.ZyncMetricsPort, Protocol: v1.ProtocolTCP})
		update = true
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Enabling metrics to DC %s", component.ZyncName))
		err = u.UpdateResource(existing)
		if err != nil {
			return false, err
		}
	}

	return update, nil
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
	if _, ok := helper.FindContainerPortByName(existing.Spec.Template.Spec.Containers[0].Ports, "metrics"); !ok {
		existing.Spec.Template.Spec.Containers[0].Ports = append(existing.Spec.Template.Spec.Containers[0].Ports, v1.ContainerPort{Name: "metrics", ContainerPort: component.SystemSidekiqMetricsPort, Protocol: v1.ProtocolTCP})
		update = true
	}

	if update {
		u.Logger().Info(fmt.Sprintf("Enabling metrics to DC %s", component.SystemSidekiqName))
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
