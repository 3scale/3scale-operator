package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type SystemMysqlAdapter struct {
}

func NewMysqlAdapter() Adapter {
	return NewAppenderAdapter(&SystemMysqlAdapter{})
}

func (m *SystemMysqlAdapter) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_USER",
			DisplayName: "System MySQL User",
			Description: "Username for System's MySQL user that will be used for accessing the database.",
			Value:       "mysql",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_PASSWORD",
			DisplayName: "System MySQL Password",
			Description: "Password for the System's MySQL user.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE",
			DisplayName: "System MySQL Database Name",
			Description: "Name of the System's MySQL database accessed.",
			Value:       "system",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_ROOT_PASSWORD",
			DisplayName: "System MySQL Root password.",
			Description: "Password for Root user.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
	}
}

func (m *SystemMysqlAdapter) Objects() ([]common.KubernetesObject, error) {
	mysqlOptions, err := m.options()
	if err != nil {
		return nil, err
	}
	mysqlComponent := component.NewSystemMysql(mysqlOptions)
	return mysqlComponent.Objects(), nil
}

func (a *SystemMysqlAdapter) options() (*component.SystemMysqlOptions, error) {
	mo := component.NewSystemMysqlOptions()
	mo.AppLabel = "${APP_LABEL}"
	mo.DatabaseName = "${SYSTEM_DATABASE}"
	mo.User = "${SYSTEM_DATABASE_USER}"
	mo.Password = "${SYSTEM_DATABASE_PASSWORD}"
	mo.RootPassword = "${SYSTEM_DATABASE_ROOT_PASSWORD}"
	mo.DatabaseURL = "mysql2://root:" + "${SYSTEM_DATABASE_ROOT_PASSWORD}" + "@system-mysql/" + "${SYSTEM_DATABASE}"

	mo.ContainerResourceRequirements = component.DefaultSystemMysqlResourceRequirements()

	err := mo.Validate()
	return mo, err
}
