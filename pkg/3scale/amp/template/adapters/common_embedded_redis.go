package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type CommonEmbeddedRedisAdapter struct {
}

func NewCommonEmbeddedRedisAdapter() Adapter {
	return NewAppenderAdapter(&CommonEmbeddedRedisAdapter{})
}

func (r *CommonEmbeddedRedisAdapter) Parameters() []templatev1.Parameter {
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

func (r *CommonEmbeddedRedisAdapter) Objects() ([]common.KubernetesObject, error) {
	redisOptions, err := r.options()
	if err != nil {
		return nil, err
	}
	redisComponent := component.NewCommonEmbeddedRedis(redisOptions)
	objects := r.componentObjects(redisComponent)
	return objects, nil
}

func (r *CommonEmbeddedRedisAdapter) componentObjects(c *component.CommonEmbeddedRedis) []common.KubernetesObject {
	commonEmbeddedRedisObjects := r.commonEmbeddedRedisComponentObjects(c)
	objects := commonEmbeddedRedisObjects
	return objects
}

func (r *CommonEmbeddedRedisAdapter) commonEmbeddedRedisComponentObjects(c *component.CommonEmbeddedRedis) []common.KubernetesObject {
	cm := c.ConfigMap()
	objects := []common.KubernetesObject{
		cm,
	}
	return objects
}

func (r *CommonEmbeddedRedisAdapter) options() (*component.CommonEmbeddedRedisOptions, error) {
	ro := component.NewCommonEmbeddedRedisOptions()
	ro.ConfigMapLabels = r.configMapLabels()

	err := ro.Validate()
	return ro, err
}

func (r *CommonEmbeddedRedisAdapter) configMapLabels() map[string]string {
	return map[string]string{
		"app":                          "${APP_LABEL}",
		"threescale_component":         "system",
		"threescale_component_element": "redis",
	}
}
