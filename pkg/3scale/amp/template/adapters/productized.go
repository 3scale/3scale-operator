package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	templatev1 "github.com/openshift/api/template/v1"
)

type ProductizationAdapter struct {
}

func NewProductizedAdapter(options []string) Adapter {
	return &ProductizationAdapter{}
}

func (p *ProductizationAdapter) Adapt(template *templatev1.Template) {
	options, err := p.options()
	if err != nil {
		panic(err)
	}
	productizedComponent := component.NewProductized(options)
	productizedComponent.UpdateAmpImagesParameters(template)
}

func (p *ProductizationAdapter) options() (*component.ProductizedOptions, error) {
	pob := component.ProductizedOptionsBuilder{}
	// Amprelease needs to exist, but not used
	// for the current need (UpdateAmpImagesParameters method)
	// TODO move validation of options to the method which uses the options
	// and knows what is required and what is not.
	pob.AmpRelease("---")
	pob.ApicastImage("${AMP_APICAST_IMAGE}")
	pob.BackendImage("${AMP_BACKEND_IMAGE}")
	pob.SystemImage("${AMP_SYSTEM_IMAGE}")
	pob.ZyncImage("${AMP_ZYNC_IMAGE}")
	return pob.Build()
}
