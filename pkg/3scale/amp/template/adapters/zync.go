package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type Zync struct {
	generatePodDisruptionBudget bool
}

func NewZyncAdapter(generatePDB bool) Adapter {
	return NewAppenderAdapter(&Zync{generatePodDisruptionBudget:generatePDB})
}

func (z *Zync) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "ZYNC_DATABASE_PASSWORD",
			DisplayName: "Zync Database PostgreSQL Connection Password",
			Description: "Password for the Zync Database PostgreSQL connection user.",
			Generate:    "expression",
			From:        "[a-zA-Z0-9]{16}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:     "ZYNC_SECRET_KEY_BASE",
			Generate: "expression",
			From:     "[a-zA-Z0-9]{16}",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "ZYNC_AUTHENTICATION_TOKEN",
			Generate: "expression",
			From:     "[a-zA-Z0-9]{16}",
			Required: true,
		},
	}
}

func (z *Zync) Objects() ([]common.KubernetesObject, error) {
	zyncOptions, err := z.options()
	if err != nil {
		return nil, err
	}
	zyncComponent := component.NewZync(zyncOptions)
	objects := zyncComponent.Objects()
	if z.generatePodDisruptionBudget {
		objects = append(objects, zyncComponent.PDBObjects()...)
	}
	return objects, nil
}

func (z *Zync) options() (*component.ZyncOptions, error) {
	zo := component.NewZyncOptions()
	zo.AppLabel = "${APP_LABEL}"
	zo.AuthenticationToken = "${ZYNC_AUTHENTICATION_TOKEN}"
	zo.DatabasePassword = "${ZYNC_DATABASE_PASSWORD}"
	zo.SecretKeyBase = "${ZYNC_SECRET_KEY_BASE}"

	zo.ZyncReplicas = 1
	zo.ZyncQueReplicas = 1

	zo.ContainerResourceRequirements = component.DefaultZyncContainerResourceRequirements()
	zo.QueContainerResourceRequirements = component.DefaultZyncQueContainerResourceRequirements()
	zo.DatabaseContainerResourceRequirements = component.DefaultZyncDatabaseContainerResourceRequirements()

	zo.DatabaseURL = component.DefaultZyncDatabaseURL(zo.DatabasePassword)

	err := zo.Validate()
	return zo, err
}
