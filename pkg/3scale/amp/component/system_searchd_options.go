package component

import (
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type SearchdPVCOptions struct {
	StorageClass    *string
	VolumeName      string
	StorageRequests resource.Quantity `validate:"required"`
}

type SystemSearchdOptions struct {
	ContainerResourceRequirements v1.ResourceRequirements `validate:"required"`
	ImageTag                      string                  `validate:"required"`

	Affinity    *v1.Affinity    `validate:"-"`
	Tolerations []v1.Toleration `validate:"-"`

	Labels            map[string]string `validate:"required"`
	PodTemplateLabels map[string]string `validate:"required"`

	PVCOptions                SearchdPVCOptions             `validate:"required"`
	PriorityClassName         string                        `validate:"-"`
	TopologySpreadConstraints []v1.TopologySpreadConstraint `validate:"-"`
}

func NewSystemSearchdOptions() *SystemSearchdOptions {
	return &SystemSearchdOptions{}
}

func (s *SystemSearchdOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

func DefaultSearchdContainerResourceRequirements() v1.ResourceRequirements {
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
