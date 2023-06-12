package component

import (
	"fmt"

	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type SystemPostgreSQLOptions struct {
	ContainerResourceRequirements v1.ResourceRequirements `validate:"-"`
	ImageTag                      string                  `validate:"required"`
	User                          string                  `validate:"required"`
	Password                      string                  `validate:"required"`
	DatabaseName                  string                  `validate:"required"`
	DatabaseURL                   string                  `validate:"required"`
	PVCStorageClass               *string
	PVCVolumeName                 *string
	PVCStorageRequests            resource.Quantity             `validate:"required"`
	Affinity                      *v1.Affinity                  `validate:"-"`
	Tolerations                   []v1.Toleration               `validate:"-"`
	CommonLabels                  map[string]string             `validate:"required"`
	DeploymentLabels              map[string]string             `validate:"required"`
	PodTemplateLabels             map[string]string             `validate:"required"`
	PriorityClassName             string                        `validate:"-"`
	TopologySpreadConstraints     []v1.TopologySpreadConstraint `validate:"-"`
}

func NewSystemPostgreSQLOptions() *SystemPostgreSQLOptions {
	return &SystemPostgreSQLOptions{}
}

func (s *SystemPostgreSQLOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

func DefaultSystemPostgresqlResourceRequirements() v1.ResourceRequirements {
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

func DefaultSystemPostgresqlUser() string {
	return "system"
}

func DefaultSystemPostgresqlPassword() string {
	return oprand.String(8)
}

func DefaultSystemPostgresqlDatabaseName() string {
	return "system"
}

func DefaultSystemPostgresqlDatabaseURL(username, password, databasename string) string {
	return fmt.Sprintf("postgresql://%s:%s@system-postgresql/%s", username, password, databasename)
}

func DefaultSystemPostgresqlStorageResources() resource.Quantity {
	return resource.MustParse("1Gi")
}
