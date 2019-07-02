package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type SystemMysqlImageAdapter struct {
}

func NewSystemMysqlImageAdapter(options []string) Adapter {
	return NewAppenderAdapter(&SystemMysqlImageAdapter{})
}

func (a *SystemMysqlImageAdapter) Parameters() []templatev1.Parameter {
	productVersion := product.CurrentProductVersion()
	imageProvider, err := product.NewImageProvider(productVersion)
	if err != nil {
		panic(err)
	}
	return []templatev1.Parameter{
		{
			Name:        "SYSTEM_DATABASE_IMAGE",
			Description: "System MySQL image to use",
			Value:       imageProvider.GetSystemMySQLImage(),
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
