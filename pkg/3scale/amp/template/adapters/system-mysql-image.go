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
	o := component.NewSystemMySQLImageOptions()
	o.AppLabel = "${APP_LABEL}"
	o.AmpRelease = "${AMP_RELEASE}"
	o.Image = "${SYSTEM_DATABASE_IMAGE}"
	tmp := component.InsecureImportPolicy
	o.InsecureImportPolicy = &tmp

	err := o.Validate()
	return o, err
}
