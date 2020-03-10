package component

import (
	"fmt"

	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ZyncOptions container object with all required to create components
type ZyncOptions struct {
	DatabaseURL                           string                  `validate:"required"`
	ContainerResourceRequirements         v1.ResourceRequirements `validate:"-"`
	QueContainerResourceRequirements      v1.ResourceRequirements `validate:"-"`
	DatabaseContainerResourceRequirements v1.ResourceRequirements `validate:"-"`
	AppLabel                              string                  `validate:"required"`
	AuthenticationToken                   string                  `validate:"required"`
	DatabasePassword                      string                  `validate:"required"`
	SecretKeyBase                         string                  `validate:"required"`
	ZyncReplicas                          int32
	ZyncQueReplicas                       int32
}

func NewZyncOptions() *ZyncOptions {
	return &ZyncOptions{}
}

func (z *ZyncOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(z)
}

func DefaultZyncSecretKeyBase() string {
	return oprand.String(16)
}

func DefaultZyncDatabasePassword() string {
	return oprand.String(16)
}

func DefaultZyncAuthenticationToken() string {
	return oprand.String(16)
}

func DefaultZyncDatabaseURL(password string) string {
	return fmt.Sprintf("postgresql://zync:%s@zync-database:5432/zync_production", password)
}

func DefaultZyncContainerResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("150m"),
			v1.ResourceMemory: resource.MustParse("250M"),
		},
	}
}

func DefaultZyncQueContainerResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("250M"),
		},
	}
}

func DefaultZyncDatabaseContainerResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("2G"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("50m"),
			v1.ResourceMemory: resource.MustParse("250M"),
		},
	}
}
