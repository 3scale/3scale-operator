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

func (a *RedisAdapter) Parameters() []templatev1.Parameter {
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
	return redisComponent.Objects(), nil
}

func (r *RedisAdapter) options() (*component.RedisOptions, error) {
	ro := component.NewRedisOptions()
	ro.AppLabel = "${APP_LABEL}"
	ro.AmpRelease = "${AMP_RELEASE}"
	ro.BackendImage = "${REDIS_IMAGE}"
	ro.SystemImage = "${REDIS_IMAGE}"
	ro.BackendRedisContainerResourceRequirements = component.DefaultBackendRedisContainerResourceRequirements()
	ro.SystemRedisContainerResourceRequirements = component.DefaultSystemRedisContainerResourceRequirements()
	tmp := component.InsecureImportPolicy
	ro.InsecureImportPolicy = &tmp

	err := ro.Validate()
	return ro, err
}
