package component

import (
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type RedisOptions struct {
	AmpRelease                                string                   `validate:"required"`
	BackendImage                              string                   `validate:"required"`
	BackendImageTag                           string                   `validate:"required"`
	SystemImage                               string                   `validate:"required"`
	SystemImageTag                            string                   `validate:"required"`
	BackendRedisContainerResourceRequirements *v1.ResourceRequirements `validate:"required"`
	SystemRedisContainerResourceRequirements  *v1.ResourceRequirements `validate:"required"`
	InsecureImportPolicy                      *bool                    `validate:"required"`

	SystemCommonLabels            map[string]string `validate:"required"`
	SystemRedisLabels             map[string]string `validate:"required"`
	SystemRedisPodTemplateLabels  map[string]string `validate:"required"`
	BackendCommonLabels           map[string]string `validate:"required"`
	BackendRedisLabels            map[string]string `validate:"required"`
	BackendRedisPodTemplateLabels map[string]string `validate:"required"`
}

func NewRedisOptions() *RedisOptions {
	return &RedisOptions{}
}

func (r *RedisOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

func DefaultBackendRedisContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("2000m"),
			v1.ResourceMemory: resource.MustParse("32Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("1024Mi"),
		},
	}
}

func DefaultSystemRedisContainerResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("32Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("150m"),
			v1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}
}
