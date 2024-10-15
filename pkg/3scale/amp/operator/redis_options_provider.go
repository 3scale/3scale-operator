package operator

import (
	"context"
	"fmt"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RedisOptionsProvider struct {
	apimanager   *appsv1alpha1.APIManager
	namespace    string
	client       client.Client
	options      *component.RedisOptions
	secretSource *helper.SecretSource
}

func NewRedisOptionsProvider(apimanager *appsv1alpha1.APIManager, namespace string, client client.Client) *RedisOptionsProvider {
	return &RedisOptionsProvider{
		apimanager:   apimanager,
		namespace:    namespace,
		client:       client,
		options:      component.NewRedisOptions(),
		secretSource: helper.NewSecretSource(client, namespace),
	}
}

func (r *RedisOptionsProvider) GetRedisOptions() (*component.RedisOptions, error) {
	r.options.AmpRelease = version.ThreescaleVersionMajorMinor()
	r.options.BackendImageTag = version.ThreescaleVersionMajorMinor()
	r.options.SystemImageTag = version.ThreescaleVersionMajorMinor()

	r.options.BackendImage = BackendRedisImageURL()
	if r.apimanager.Spec.Backend != nil && r.apimanager.Spec.Backend.RedisImage != nil {
		r.options.BackendImage = *r.apimanager.Spec.Backend.RedisImage
	}

	r.options.SystemImage = SystemRedisImageURL()
	if r.apimanager.Spec.System != nil && r.apimanager.Spec.System.RedisImage != nil {
		r.options.SystemImage = *r.apimanager.Spec.System.RedisImage
	}

	r.options.SystemCommonLabels = r.systemCommonLabels()
	r.options.SystemRedisLabels = r.systemRedisLabels()
	r.options.SystemRedisPodTemplateLabels = r.systemRedisPodTemplateLabels()
	r.options.BackendCommonLabels = r.backendCommonLabels()
	r.options.BackendRedisLabels = r.backendRedisLabels()
	r.options.BackendRedisPodTemplateLabels = r.backendRedisPodTemplateLabels()

	var err error
	r.options.SystemRedisPodTemplateAnnotations, err = r.systemRedisPodTemplateAnnotations()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions systemRedisPodTemplateAnnotations: %w", err)
	}
	r.options.BackendRedisPodTemplateAnnotations, err = r.backendRedisPodTemplateAnnotations()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions backendRedisPodTemplateAnnotations: %w", err)
	}

	r.setResourceRequirementsOptions()
	r.setNodeAffinityAndTolerationsOptions()

	r.setPersistentVolumeClaimOptions()
	r.setPriorityClassNames()
	r.setTopologySpreadConstraints()

	// Should the operator be reading redis secrets?
	// When HA is disabled, do we support external redis?
	// If answer is true, why does the operator deploy redis?
	// If the answer is no, then it would be sufficient to set default URL's (internal redis url)
	// to options and reconciliate secret for owner reference

	err = r.setSecretBasedOptions()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions reading secret options: %w", err)
	}

	err = r.options.Validate()
	if err != nil {
		return nil, fmt.Errorf("GetRedisOptions validating: %w", err)
	}
	return r.options, nil
}

func (r *RedisOptionsProvider) setSecretBasedOptions() error {
	cases := []struct {
		field       *string
		secretName  string
		secretField string
		defValue    string
	}{
		{
			&r.options.BackendStorageURL,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageURLFieldName,
			component.DefaultBackendRedisStorageURL(),
		},
		{
			&r.options.BackendQueuesURL,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesURLFieldName,
			component.DefaultBackendRedisQueuesURL(),
		},
		{
			&r.options.BackendRedisStorageSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelHostsFieldName,
			component.DefaultBackendStorageSentinelHosts(),
		},
		{
			&r.options.BackendRedisStorageSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisStorageSentinelRoleFieldName,
			component.DefaultBackendStorageSentinelRole(),
		},
		{
			&r.options.BackendRedisQueuesSentinelHosts,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelHostsFieldName,
			component.DefaultBackendQueuesSentinelHosts(),
		},
		{
			&r.options.BackendRedisQueuesSentinelRole,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisQueuesSentinelRoleFieldName,
			component.DefaultBackendQueuesSentinelRole(),
		},
		{
			&r.options.SystemRedisURL,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisURLFieldName,
			component.DefaultSystemRedisURL(),
		},
		{
			&r.options.SystemRedisSentinelsHosts,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisSentinelHosts,
			component.DefaultSystemRedisSentinelHosts(),
		},
		{
			&r.options.SystemRedisSentinelsRole,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisSentinelRole,
			component.DefaultSystemRedisSentinelRole(),
		},

		//TLS
		{
			&r.options.SystemRedisCAFile,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisCAFile,
			component.DefaultSystemRedisCAFile(),
		},
		{
			&r.options.SystemRedisClientCertificate,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisClientCertificate,
			component.DefaultSystemRedisClientCertificate(),
		},
		{
			&r.options.SystemRedisPrivateKey,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisPrivateKey,
			component.DefaultSystemRedisPrivateKey(),
		},
		{
			&r.options.SystemRedisSSL,
			component.SystemSecretSystemRedisSecretName,
			component.SystemSecretSystemRedisSSL,
			component.DefaultSystemRedisSSL(),
		},
		// TLS / Backend
		{
			&r.options.BackendConfigCAFile,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisConfigCAFile,
			component.DefaultBackendConfigCAFile(),
		},
		{
			&r.options.BackendConfigClientCertificate,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisConfigClientCertificate,
			component.DefaultBackendConfigClientCertificate(),
		},
		{
			&r.options.BackendConfigPrivateKey,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisConfigPrivateKey,
			component.DefaultBackendConfigPrivateKey(),
		},
		{
			&r.options.BackendConfigSSL,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisConfigSSL,
			component.DefaultBackendConfigSSL(),
		},
		{
			&r.options.BackendConfigQueuesCAFile,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisConfigQueuesCAFile,
			component.DefaultBackendConfigQueuesCAFile(),
		},
		{
			&r.options.BackendConfigQueuesClientCertificate,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisConfigQueuesClientCertificate,
			component.DefaultBackendConfigQueuesClientCertificate(),
		},
		{
			&r.options.BackendConfigQueuesPrivateKey,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisConfigQueuesPrivateKey,
			component.DefaultBackendConfigQueuesPrivateKey(),
		},
		{
			&r.options.BackendConfigQueuesSSL,
			component.BackendSecretBackendRedisSecretName,
			component.BackendSecretBackendRedisConfigQueuesSSL,
			component.DefaultBackendConfigQueuesSSL(),
		},
	}

	for _, option := range cases {
		val, err := r.secretSource.FieldValue(option.secretName, option.secretField, option.defValue)
		if err != nil {
			return err
		}
		*option.field = val
	}

	return nil
}

func (r *RedisOptionsProvider) setResourceRequirementsOptions() {
	if *r.apimanager.Spec.ResourceRequirementsEnabled {
		r.options.BackendRedisContainerResourceRequirements = component.DefaultBackendRedisContainerResourceRequirements()
		r.options.SystemRedisContainerResourceRequirements = component.DefaultSystemRedisContainerResourceRequirements()
	} else {
		r.options.BackendRedisContainerResourceRequirements = &v1.ResourceRequirements{}
		r.options.SystemRedisContainerResourceRequirements = &v1.ResourceRequirements{}
	}

	// Deployment-level ResourceRequirements CR fields have priority over
	// spec.resourceRequirementsEnabled, overwriting that setting when they are
	// defined
	if r.apimanager.Spec.Backend.RedisResources != nil {
		r.options.BackendRedisContainerResourceRequirements = r.apimanager.Spec.Backend.RedisResources
	}
	if r.apimanager.Spec.System.RedisResources != nil {
		r.options.SystemRedisContainerResourceRequirements = r.apimanager.Spec.System.RedisResources
	}
}

func (r *RedisOptionsProvider) setPersistentVolumeClaimOptions() {
	if r.apimanager.Spec.System != nil &&
		r.apimanager.Spec.System.RedisPersistentVolumeClaimSpec != nil {
		r.options.SystemRedisPVCStorageClass = r.apimanager.Spec.System.RedisPersistentVolumeClaimSpec.StorageClassName
	}
	if r.apimanager.Spec.Backend != nil &&
		r.apimanager.Spec.Backend.RedisPersistentVolumeClaimSpec != nil {
		r.options.BackendRedisPVCStorageClass = r.apimanager.Spec.Backend.RedisPersistentVolumeClaimSpec.StorageClassName
	}
}

func (r *RedisOptionsProvider) setNodeAffinityAndTolerationsOptions() {
	r.options.BackendRedisAffinity = r.apimanager.Spec.Backend.RedisAffinity
	r.options.BackendRedisTolerations = r.apimanager.Spec.Backend.RedisTolerations
	r.options.SystemRedisAffinity = r.apimanager.Spec.System.RedisAffinity
	r.options.SystemRedisTolerations = r.apimanager.Spec.System.RedisTolerations
}

func (r *RedisOptionsProvider) systemCommonLabels() map[string]string {
	return component.SystemCommonLabels(*r.apimanager.Spec.AppLabel)
}

func (r *RedisOptionsProvider) backendCommonLabels() map[string]string {
	return component.BackendCommonLabels(*r.apimanager.Spec.AppLabel)
}

func (r *RedisOptionsProvider) systemRedisLabels() map[string]string {
	return component.SystemRedisLabels(*r.apimanager.Spec.AppLabel)
}

func (r *RedisOptionsProvider) backendRedisLabels() map[string]string {
	return component.BackendRedisLabels(*r.apimanager.Spec.AppLabel)
}

func (r *RedisOptionsProvider) systemRedisPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("system-redis", helper.ApplicationType)

	for k, v := range r.systemRedisLabels() {
		labels[k] = v
	}

	for k, v := range r.apimanager.Spec.System.RedisLabels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "system-redis"

	return labels
}

func (r *RedisOptionsProvider) backendRedisPodTemplateLabels() map[string]string {
	labels := helper.MeteringLabels("backend-redis", helper.ApplicationType)

	for k, v := range r.backendRedisLabels() {
		labels[k] = v
	}

	for k, v := range r.apimanager.Spec.Backend.RedisLabels {
		labels[k] = v
	}

	labels[reconcilers.DeploymentLabelSelector] = "backend-redis"

	return labels
}

func (r *RedisOptionsProvider) setPriorityClassNames() {
	if r.apimanager.Spec.System != nil && r.apimanager.Spec.System.RedisPriorityClassName != nil {
		r.options.SystemRedisPriorityClassName = *r.apimanager.Spec.System.RedisPriorityClassName
	}
	if r.apimanager.Spec.Backend != nil && r.apimanager.Spec.Backend.RedisPriorityClassName != nil {
		r.options.BackendRedisPriorityClassName = *r.apimanager.Spec.Backend.RedisPriorityClassName
	}
}

func (r *RedisOptionsProvider) setTopologySpreadConstraints() {
	if r.apimanager.Spec.System != nil && r.apimanager.Spec.System.RedisTopologySpreadConstraints != nil {
		r.options.SystemRedisTopologySpreadConstraints = r.apimanager.Spec.System.RedisTopologySpreadConstraints
	}
	if r.apimanager.Spec.Backend != nil && r.apimanager.Spec.Backend.RedisTopologySpreadConstraints != nil {
		r.options.BackendRedisTopologySpreadConstraints = r.apimanager.Spec.Backend.RedisTopologySpreadConstraints
	}
}

func (r *RedisOptionsProvider) systemRedisPodTemplateAnnotations() (map[string]string, error) {
	annotations := make(map[string]string)
	for k, v := range r.apimanager.Spec.System.RedisAnnotations {
		annotations[k] = v
	}
	// configmap must exist, it has been previously checked
	resourceVersion, err := r.redisConfigMapResourceVersion()
	if err != nil {
		return nil, err
	}
	annotations["redisConfigMapResourceVersion"] = resourceVersion
	return annotations, nil
}

func (r *RedisOptionsProvider) backendRedisPodTemplateAnnotations() (map[string]string, error) {
	annotations := make(map[string]string)
	for k, v := range r.apimanager.Spec.Backend.RedisAnnotations {
		annotations[k] = v
	}
	// configmap must exist, it has been previously checked
	resourceVersion, err := r.redisConfigMapResourceVersion()
	if err != nil {
		return nil, err
	}
	annotations["redisConfigMapResourceVersion"] = resourceVersion
	return annotations, nil
}

func (r *RedisOptionsProvider) redisConfigMapResourceVersion() (string, error) {
	// get resourceVersion from CM and use it as Annotation "generationID" for redis Pods
	key := client.ObjectKey{
		Namespace: r.namespace,
		Name:      "redis-config",
	}
	cm := &v1.ConfigMap{}
	err := r.client.Get(context.Background(), key, cm)
	if err != nil {
		return "", err
	}
	return cm.GetResourceVersion(), nil
}
