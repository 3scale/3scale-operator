package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type BackendRedisAdapter struct {
}

func NewBackendRedisAdapter() Adapter {
	return NewAppenderAdapter(&BackendRedisAdapter{})
}

func (r *BackendRedisAdapter) Parameters() []templatev1.Parameter {
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

func (r *BackendRedisAdapter) Objects() ([]common.KubernetesObject, error) {
	redisOptions, err := r.options()
	if err != nil {
		return nil, err
	}
	redisComponent := component.NewBackendRedis(redisOptions)
	objects := r.componentObjects(redisComponent)
	return objects, nil
}

func (r *BackendRedisAdapter) componentObjects(c *component.BackendRedis) []common.KubernetesObject {
	backendRedisObjects := r.backendRedisComponentObjects(c)
	objects := backendRedisObjects
	return objects
}

func (r *BackendRedisAdapter) backendRedisComponentObjects(c *component.BackendRedis) []common.KubernetesObject {
	dc := c.DeploymentConfig()
	bs := c.Service()
	cm := c.ConfigMap()
	bpvc := c.PersistentVolumeClaim()
	bis := c.ImageStream()
	objects := []common.KubernetesObject{
		dc,
		bs,
		cm,
		bpvc,
		bis,
	}
	return objects
}

func (r *BackendRedisAdapter) options() (*component.BackendRedisOptions, error) {
	ro := component.NewBackendRedisOptions()
	ro.AmpRelease = "${AMP_RELEASE}"
	ro.ImageTag = "${AMP_RELEASE}"
	ro.Image = "${REDIS_IMAGE}"

	ro.ContainerResourceRequirements = component.DefaultBackendRedisContainerResourceRequirements()
	tmp := component.InsecureImportPolicy
	ro.InsecureImportPolicy = &tmp

	ro.BackendCommonLabels = r.backendCommonLabels()
	ro.RedisLabels = r.backendRedisLabels()
	ro.PodTemplateLabels = r.backendRedisPodTemplateLabels()

	err := ro.Validate()
	return ro, err
}

func (r *BackendRedisAdapter) backendCommonLabels() map[string]string {
	return map[string]string{
		"app":                  "${APP_LABEL}",
		"threescale_component": "backend",
	}
}

func (r *BackendRedisAdapter) backendRedisLabels() map[string]string {
	labels := r.backendCommonLabels()
	labels["threescale_component_element"] = "redis"
	return labels
}

func (r *BackendRedisAdapter) backendRedisPodTemplateLabels() map[string]string {
	labels := r.backendRedisLabels()
	labels["deploymentConfig"] = "backend-redis"
	return labels
}
