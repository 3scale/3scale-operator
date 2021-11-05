package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type RedisAdapter struct {
}

func NewRedisAdapter() Adapter {
	return NewAppenderAdapter(&RedisAdapter{})
}

func (r *RedisAdapter) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		{
			Name:        "REDIS_IMAGE",
			Description: "Redis image to use",
			Required:    true,
			// We use backend-redis image because we have to choose one
			// but in templates there's no distinction between Backend Redis image
			// used and System Redis image. They are always the same
			Value: component.BackendRedisImageURL(),
		},
	}
}

func (r *RedisAdapter) Objects() ([]common.KubernetesObject, error) {
	redisOptions, err := r.options()
	if err != nil {
		return nil, err
	}
	redisComponent := component.NewRedis(redisOptions)
	objects := r.componentObjects(redisComponent)
	return objects, nil
}

func (r *RedisAdapter) componentObjects(c *component.Redis) []common.KubernetesObject {
	backendRedisObjects := r.backendRedisComponentObjects(c)
	systemRedisObjects := r.systemRedisComponentObjects(c)

	objects := backendRedisObjects
	objects = append(objects, systemRedisObjects...)
	return objects
}

func (r *RedisAdapter) backendRedisComponentObjects(c *component.Redis) []common.KubernetesObject {
	dc := c.BackendDeploymentConfig()
	bs := c.BackendService()
	cm := c.BackendConfigMap()
	bpvc := c.BackendPVC()
	bis := c.BackendImageStream()
	backendRedisSecret := c.BackendRedisSecret()
	systemRedisSecret := c.SystemRedisSecret()
	objects := []common.KubernetesObject{
		dc,
		bs,
		cm,
		bpvc,
		bis,
		backendRedisSecret,
		systemRedisSecret,
	}
	return objects
}

func (r *RedisAdapter) systemRedisComponentObjects(c *component.Redis) []common.KubernetesObject {
	systemRedisDC := c.SystemDeploymentConfig()
	systemRedisPVC := c.SystemPVC()
	systemRedisService := c.SystemService()
	systemRedisImageStream := c.SystemImageStream()

	objects := []common.KubernetesObject{
		systemRedisDC,
		systemRedisPVC,
		systemRedisService,
		systemRedisImageStream,
	}

	return objects
}

func (r *RedisAdapter) options() (*component.RedisOptions, error) {
	ro := component.NewRedisOptions()
	ro.AmpRelease = "${AMP_RELEASE}"
	ro.BackendImageTag = "${AMP_RELEASE}"
	ro.BackendImage = "${REDIS_IMAGE}"
	ro.SystemImageTag = "${AMP_RELEASE}"
	ro.SystemImage = "${REDIS_IMAGE}"

	ro.BackendRedisContainerResourceRequirements = component.DefaultBackendRedisContainerResourceRequirements()
	ro.SystemRedisContainerResourceRequirements = component.DefaultSystemRedisContainerResourceRequirements()
	tmp := component.InsecureImportPolicy
	ro.InsecureImportPolicy = &tmp

	ro.SystemCommonLabels = r.systemCommonLabels()
	ro.SystemRedisLabels = r.systemRedisLabels()
	ro.SystemRedisPodTemplateLabels = r.systemRedisPodTemplateLabels()
	ro.BackendCommonLabels = r.backendCommonLabels()
	ro.BackendRedisLabels = r.backendRedisLabels()
	ro.BackendRedisPodTemplateLabels = r.backendRedisPodTemplateLabels()

	ro.SystemRedisURL = "${SYSTEM_REDIS_URL}"
	ro.SystemRedisSentinelsHosts = component.DefaultSystemRedisSentinelHosts()
	ro.SystemRedisSentinelsRole = component.DefaultSystemRedisSentinelRole()
	ro.SystemRedisNamespace = "${SYSTEM_REDIS_NAMESPACE}"

	ro.BackendStorageURL = component.DefaultBackendRedisStorageURL()
	ro.BackendQueuesURL = component.DefaultBackendRedisQueuesURL()
	ro.BackendRedisStorageSentinelHosts = component.DefaultBackendStorageSentinelHosts()
	ro.BackendRedisStorageSentinelRole = component.DefaultBackendStorageSentinelRole()
	ro.BackendRedisQueuesSentinelHosts = component.DefaultBackendQueuesSentinelHosts()
	ro.BackendRedisQueuesSentinelRole = component.DefaultBackendQueuesSentinelRole()

	err := ro.Validate()
	return ro, err
}

func (r *RedisAdapter) systemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  "${APP_LABEL}",
		"threescale_component": "system",
	}
}

func (r *RedisAdapter) systemRedisLabels() map[string]string {
	labels := r.systemCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

func (r *RedisAdapter) systemRedisPodTemplateLabels() map[string]string {
	labels := r.systemRedisLabels()
	labels["deploymentConfig"] = "system-redis"
	return labels
}

func (r *RedisAdapter) backendCommonLabels() map[string]string {
	return map[string]string{
		"app":                  "${APP_LABEL}",
		"threescale_component": "backend",
	}
}

func (r *RedisAdapter) backendRedisLabels() map[string]string {
	labels := r.backendCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

func (r *RedisAdapter) backendRedisPodTemplateLabels() map[string]string {
	labels := r.backendRedisLabels()
	labels["deploymentConfig"] = "backend-redis"
	return labels
}
