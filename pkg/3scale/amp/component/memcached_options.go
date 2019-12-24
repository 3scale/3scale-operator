package component

import (
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type MemcachedOptions struct {
	AppLabel             string                  `validate:"required"`
	ResourceRequirements v1.ResourceRequirements `validate:"-"`
}

func NewMemcachedOptions() *MemcachedOptions {
	return &MemcachedOptions{}
}

func (m *MemcachedOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}

func DefaultMemcachedResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("96Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("64Mi"),
		},
	}
}
