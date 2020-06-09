package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type SystemRedisAdapter struct {
}

func NewSystemRedisAdapter() Adapter {
	return NewAppenderAdapter(&SystemRedisAdapter{})
}

func (r *SystemRedisAdapter) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{}
}

func (r *SystemRedisAdapter) Objects() ([]common.KubernetesObject, error) {
	redisOptions, err := r.options()
	if err != nil {
		return nil, err
	}
	redisComponent := component.NewSystemRedis(redisOptions)
	objects := r.componentObjects(redisComponent)
	return objects, nil
}

func (r *SystemRedisAdapter) componentObjects(c *component.SystemRedis) []common.KubernetesObject {
	systemRedisObjects := r.systemRedisComponentObjects(c)

	objects := systemRedisObjects
	return objects
}

func (r *SystemRedisAdapter) systemRedisComponentObjects(c *component.SystemRedis) []common.KubernetesObject {
	systemRedisDC := c.DeploymentConfig()
	systemRedisPVC := c.PersistentVolumeClaim()
	systemRedisService := c.Service()
	systemRedisImageStream := c.ImageStream()

	objects := []common.KubernetesObject{
		systemRedisDC,
		systemRedisPVC,
		systemRedisService,
		systemRedisImageStream,
	}

	return objects
}

func (r *SystemRedisAdapter) options() (*component.SystemRedisOptions, error) {
	ro := component.NewSystemRedisOptions()
	ro.AmpRelease = "${AMP_RELEASE}"
	ro.ImageTag = "${AMP_RELEASE}"
	ro.Image = "${REDIS_IMAGE}"

	ro.ContainerResourceRequirements = component.DefaultSystemRedisContainerResourceRequirements()
	tmp := component.InsecureImportPolicy
	ro.InsecureImportPolicy = &tmp

	ro.SystemCommonLabels = r.systemCommonLabels()
	ro.RedisLabels = r.systemRedisLabels()
	ro.PodTemplateLabels = r.systemRedisPodTemplateLabels()

	err := ro.Validate()
	return ro, err
}

func (r *SystemRedisAdapter) systemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  "${APP_LABEL}",
		"threescale_component": "system",
	}
}

func (r *SystemRedisAdapter) systemRedisLabels() map[string]string {
	labels := r.systemCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

func (r *SystemRedisAdapter) systemRedisPodTemplateLabels() map[string]string {
	labels := r.systemRedisLabels()
	labels["deploymentConfig"] = "system-redis"
	return labels
}
