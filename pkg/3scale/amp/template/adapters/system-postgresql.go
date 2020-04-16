package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type SystemPostgreSQLAdapter struct {
}

func NewSystemPostgreSQLAdapter() Adapter {
	return NewAppenderAdapter(&SystemPostgreSQLAdapter{})
}

func (a *SystemPostgreSQLAdapter) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_USER",
			DisplayName: "System PostgreSQL User",
			Description: "Username for PostgreSQL user that will be used for accessing the database.",
			Value:       "system",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_PASSWORD",
			DisplayName: "System PostgreSQL Password",
			Description: "Password for the System's PostgreSQL user.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE",
			DisplayName: "System PostgreSQL Database Name",
			Description: "Name of the System's PostgreSQL database accessed.",
			Value:       "system",
			Required:    true,
		},
	}
}

func (r *SystemPostgreSQLAdapter) Objects() ([]common.KubernetesObject, error) {
	systemPostgreSQLOptions, err := r.options()
	if err != nil {
		return nil, err
	}
	systemPostgreSQLComponent := component.NewSystemPostgreSQL(systemPostgreSQLOptions)
	objects := r.componentObjects(systemPostgreSQLComponent)
	return objects, nil
}

func (r *SystemPostgreSQLAdapter) componentObjects(c *component.SystemPostgreSQL) []common.KubernetesObject {
	deploymentConfig := c.DeploymentConfig()
	service := c.Service()
	persistentVolumeClaim := c.DataPersistentVolumeClaim()
	systemDatabaseSecret := c.SystemDatabaseSecret()

	objects := []common.KubernetesObject{
		deploymentConfig,
		service,
		persistentVolumeClaim,
		systemDatabaseSecret,
	}

	return objects
}

func (r *SystemPostgreSQLAdapter) options() (*component.SystemPostgreSQLOptions, error) {
	o := component.NewSystemPostgreSQLOptions()
	o.AppLabel = "${APP_LABEL}"
	o.ImageTag = "${AMP_RELEASE}"
	o.DatabaseName = "${SYSTEM_DATABASE}"
	o.User = "${SYSTEM_DATABASE_USER}"
	o.Password = "${SYSTEM_DATABASE_PASSWORD}"
	o.DatabaseURL = "postgresql://${SYSTEM_DATABASE_USER}:" + "${SYSTEM_DATABASE_PASSWORD}" + "@system-postgresql/" + "${SYSTEM_DATABASE}"

	o.ContainerResourceRequirements = component.DefaultSystemPostgresqlResourceRequirements()

	err := o.Validate()
	return o, err
}
