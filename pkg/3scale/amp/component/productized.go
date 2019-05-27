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
			param.Value = "registry.access.redhat.com/3scale-amp25/system"
		case "AMP_BACKEND_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp25/backend"
		case "AMP_APICAST_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp25/apicast-gateway"
		case "AMP_ZYNC_IMAGE":
			param.Value = "registry.access.redhat.com/3scale-amp25/zync"
		}
	}
}
