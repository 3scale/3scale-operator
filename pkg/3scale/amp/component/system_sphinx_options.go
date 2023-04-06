package component

import (
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type SphinxPVCOptions struct {
	StorageClass    *string
	VolumeName      string
	StorageRequests resource.Quantity `validate:"required"`
}

type SystemSphinxOptions struct {
	ContainerResourceRequirements v1.ResourceRequirements `validate:"required"`
	ImageTag                      string                  `validate:"required"`

	Affinity    *v1.Affinity    `validate:"-"`
	Tolerations []v1.Toleration `validate:"-"`

	Labels            map[string]string `validate:"required"`
	PodTemplateLabels map[string]string `validate:"required"`

	PVCOptions SphinxPVCOptions `validate:"required"`
}

func NewSystemSphinxOptions() *SystemSphinxOptions {
	return &SystemSphinxOptions{}
}

func (s *SystemSphinxOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

func DefaultSphinxContainerResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1000m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("80m"),
			v1.ResourceMemory: resource.MustParse("250Mi"),
		},
	}
}
