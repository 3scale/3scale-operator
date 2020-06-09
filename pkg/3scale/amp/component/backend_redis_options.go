package component

import (
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type BackendRedisOptions struct {
	AmpRelease                    string                   `validate:"required"`
	Image                         string                   `validate:"required"`
	ImageTag                      string                   `validate:"required"`
	ContainerResourceRequirements *v1.ResourceRequirements `validate:"required"`
	InsecureImportPolicy          *bool                    `validate:"required"`

	BackendCommonLabels map[string]string `validate:"required"`
	RedisLabels         map[string]string `validate:"required"`
	PodTemplateLabels   map[string]string `validate:"required"`
}

func NewBackendRedisOptions() *BackendRedisOptions {
	return &BackendRedisOptions{}
}

func (r *BackendRedisOptions) Validate() error {
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
