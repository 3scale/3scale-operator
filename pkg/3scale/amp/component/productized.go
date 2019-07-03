package component

import (
	templatev1 "github.com/openshift/api/template/v1"
)

type Productized struct {
	Options *ProductizedOptions
}

func NewProductized(options *ProductizedOptions) *Productized {
	return &Productized{Options: options}
}

func (productized *Productized) UpdateAmpImagesParameters(template *templatev1.Template) {
	for paramIdx := range template.Parameters {
		param := &template.Parameters[paramIdx]
		switch param.Name {
		case "AMP_SYSTEM_IMAGE":
			param.Value = "registry.redhat.io/3scale-amp26/system"
		case "AMP_BACKEND_IMAGE":
			param.Value = "registry.redhat.io/3scale-amp26/backend"
		case "AMP_APICAST_IMAGE":
			param.Value = "registry.redhat.io/3scale-amp26/apicast-gateway"
		case "AMP_ZYNC_IMAGE":
			param.Value = "registry.redhat.io/3scale-amp26/zync"
		}
	}
}
