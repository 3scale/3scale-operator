package operator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpgradeApiManager struct {
	Cr              *appsv1alpha1.APIManager
	Client          client.Client
	Logger          logr.Logger
	ApiClientReader client.Reader
	Scheme          *runtime.Scheme
	Cfg             *restclient.Config
}

func (u *UpgradeApiManager) Upgrade() (reconcile.Result, error) {
	res, err := u.upgradeImages()
	if res.Requeue || err != nil {
		return res, err
	}

	// upgrade system-master-apicast secret
	res, err = u.upgradeSystemMasterApicastSecret()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeImages() (reconcile.Result, error) {
	res, err := u.upgradeAMPImageStreams()
	if res.Requeue || err != nil {
		return res, err
	}

	if !u.Cr.IsExternalDatabaseEnabled() {
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

	if !u.Cr.IsExternalDatabaseEnabled() {
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
	apicast, err := Apicast(u.Cr)
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
	backend, err := Backend(u.Cr, u.Client)
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
	zync, err := Zync(u.Cr, u.Client)
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
	memcached, err := Memcached(u.Cr)
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
	redis, err := Redis(u.Cr)
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
	redis, err := Redis(u.Cr)
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
	if u.Cr.Spec.System.DatabaseSpec != nil && u.Cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.upgradeSystemPostgreSQLDeploymentConfig()
	}

	// default is MySQL
	return u.upgradeSystemMySQLDeploymentConfig()
}

func (u *UpgradeApiManager) upgradeSystemPostgreSQLDeploymentConfig() (reconcile.Result, error) {
	systemPostgreSQL, err := SystemPostgreSQL(u.Cr, u.Client)
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
	systemMySQL, err := SystemMySQL(u.Cr, u.Client)
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
	system, err := System(u.Cr, u.Client)
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
	err := u.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: u.Cr.Namespace}, existing)
	if err != nil {
		return reconcile.Result{}, err
	}

	changed, err := u.ensureDeploymentConfigImageChangeTrigger(desired, existing)
	if err != nil {
		return reconcile.Result{}, err
	}
	if changed {
		u.Logger.Info(fmt.Sprintf("Update object %s", ObjectInfo(existing)))
		err := u.Client.Update(context.TODO(), existing)
		return reconcile.Result{Requeue: true}, err
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
		u.Logger.V(1).Info(fmt.Sprintf("%s ImageStream tag name in imageChangeParams trigger changed: %s", desired.Name, diff))
		existingDeploymentTriggerImageChangeParams.From.Name = desiredDeploymentTriggerImageChangeParams.From.Name
		return true, nil
	}

	return false, nil
}

func (u *UpgradeApiManager) upgradeAMPImageStreams() (reconcile.Result, error) {
	// implement upgrade procedure by reconcile procedure
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger, u.Cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewAMPImagesReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeBackendRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.Cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger, u.Cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewImageStreamBaseReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr), NewImageStreamGenericReconciler())
	return reconcile.Result{}, reconciler.Reconcile(redis.BackendImageStream())
}

func (u *UpgradeApiManager) upgradeSystemRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.Cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger, u.Cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewImageStreamBaseReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr), NewImageStreamGenericReconciler())
	return reconcile.Result{}, reconciler.Reconcile(redis.SystemImageStream())
}

func (u *UpgradeApiManager) upgradeSystemDatabaseImageStream() (reconcile.Result, error) {
	if u.Cr.Spec.System.DatabaseSpec != nil && u.Cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.upgradeSystemPostgreSQLImageStream()
	}

	// default is MySQL
	return u.upgradeSystemMySQLImageStream()
}

func (u *UpgradeApiManager) upgradeSystemMySQLImageStream() (reconcile.Result, error) {
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger, u.Cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewSystemMySQLImageReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeSystemPostgreSQLImageStream() (reconcile.Result, error) {
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger, u.Cfg)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewSystemPostgreSQLImageReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
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

	if !u.Cr.IsExternalDatabaseEnabled() {
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
	ampimages, err := AmpImages(u.Cr)
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
	redis, err := Redis(u.Cr)
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
	redis, err := Redis(u.Cr)
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
	if u.Cr.Spec.System.DatabaseSpec != nil && u.Cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.deleteSystemPostgreSQLOldImageStreamTags()
	}

	// default is MySQL
	return u.deleteSystemMySQLOldImageStreamTags()
}

func (u *UpgradeApiManager) deleteSystemPostgreSQLOldImageStreamTags() (reconcile.Result, error) {
	postgresqlImage, err := SystemPostgreSQLImage(u.Cr)
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
	mysqlImage, err := SystemMySQLImage(u.Cr)
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
	err := u.Client.Get(context.TODO(), types.NamespacedName{Name: imageStreamName, Namespace: u.Cr.Namespace}, existingImageStream)
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
		err := u.ApiClientReader.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s:%s", existing.Name, tagRefName), Namespace: u.Cr.Namespace}, existingIsTag)
		if err != nil {
			return reconcile.Result{}, err
		}
		u.Logger.Info(fmt.Sprintf("Delete object ImageStreamTag/%s", existingIsTag.GetName()))
		err = u.Client.Delete(context.TODO(), existingIsTag)
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
		Namespace: u.Cr.Namespace,
	}
	err := u.Client.Get(context.TODO(), secretNamespacedName, existing)
	// NotFound also regarded as error, as secret is expected to exist
	if err != nil {
		return reconcile.Result{}, err
	}

	if _, ok := existing.Data["BASE_URL"]; ok {
		// Remove unused BASE_URL field
		patchJSON := []byte(`[{"op": "remove", "path": "/data/BASE_URL"}]`)
		// Apply JSON patch https://tools.ietf.org/html/rfc6902
		patch := client.ConstantPatch(types.JSONPatchType, patchJSON)
		err = u.Client.Patch(context.TODO(), existing, patch)
		if err != nil {
			return reconcile.Result{}, err
		}
		u.Logger.Info(fmt.Sprintf("Upgrade: patch object %s", ObjectInfo(existing)))
	}

	return reconcile.Result{}, nil
}
