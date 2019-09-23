package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type SystemMysqlImageAdapter struct {
}

func NewSystemMysqlImageAdapter() Adapter {
	return NewAppenderAdapter(&SystemMysqlImageAdapter{})
}

func (a *SystemMysqlImageAdapter) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		{
			Name:        "SYSTEM_DATABASE_IMAGE",
			Description: "System MySQL image to use",
			Value:       component.SystemMySQLImageURL(),
			Required:    true,
		},
	}
}

func (r *SystemMysqlImageAdapter) Objects() ([]common.KubernetesObject, error) {
	systemMysqlOptions, err := r.options()
	if err != nil {
		return nil, err
	}
	systemMysqlImageComponent := component.NewSystemMySQLImage(systemMysqlOptions)
	return systemMysqlImageComponent.Objects(), nil
}

func (r *SystemMysqlImageAdapter) options() (*component.SystemMySQLImageOptions, error) {
	aob := component.SystemMySQLImageOptionsBuilder{}
	aob.AppLabel("${APP_LABEL}")
	aob.AmpRelease("${AMP_RELEASE}")
	aob.Image("${SYSTEM_DATABASE_IMAGE}")
	aob.InsecureImportPolicy(component.InsecureImportPolicy)

	return aob.Build()
}
