package component

import (
	"fmt"

	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type SystemMysqlOptions struct {
	AppLabel                      string                  `validate:"required"`
	DatabaseName                  string                  `validate:"required"`
	User                          string                  `validate:"required"`
	Password                      string                  `validate:"required"`
	RootPassword                  string                  `validate:"required"`
	DatabaseURL                   string                  `validate:"required"`
	ContainerResourceRequirements v1.ResourceRequirements `validate:"-"`
}

func NewSystemMysqlOptions() *SystemMysqlOptions {
	return &SystemMysqlOptions{}
}

func (s *SystemMysqlOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

func DefaultSystemMysqlResourceRequirements() v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("250m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
}

func DefaultSystemMysqlUser() string {
	return "mysql"
}

func DefaultSystemMysqlPassword() string {
	return oprand.String(8)
}

func DefaultSystemMysqlRootPassword() string {
	return oprand.String(8)
}

func DefaultSystemMysqlDatabaseName() string {
	return "system"
}

func DefaultSystemMysqlDatabaseURL(password, name string) string {
	return fmt.Sprintf("mysql2://root:%s@system-mysql/%s", password, name)
}
