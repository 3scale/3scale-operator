package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type SystemPostgreSQLImageAdapter struct {
}

func NewSystemPostgreSQLImageAdapter() Adapter {
	return NewAppenderAdapter(&SystemPostgreSQLImageAdapter{})
}

func (a *SystemPostgreSQLImageAdapter) Parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "SYSTEM_DATABASE_IMAGE",
			Description: "System PostgreSQL image to use",
			Required:    true,
			Value:       component.SystemPostgreSQLImageURL(),
		},
	}
}

func (r *SystemPostgreSQLImageAdapter) Objects() ([]common.KubernetesObject, error) {
	systemPostgreSQLOptions, err := r.options()
	if err != nil {
		return nil, err
	}
	systemPostgreSQLComponent := component.NewSystemPostgreSQLImage(systemPostgreSQLOptions)
	return systemPostgreSQLComponent.Objects(), nil
}

func (r *SystemPostgreSQLImageAdapter) options() (*component.SystemPostgreSQLImageOptions, error) {
	o := component.NewSystemPostgreSQLImageOptions()
	o.AppLabel = "${APP_LABEL}"
	o.AmpRelease = "${AMP_RELEASE}"
	o.Image = "${SYSTEM_DATABASE_IMAGE}"
	tmp := component.InsecureImportPolicy
	o.InsecureImportPolicy = &tmp

	err := o.Validate()
	return o, err
}
