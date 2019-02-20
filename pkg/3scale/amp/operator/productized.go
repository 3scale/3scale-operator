package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
)

func (o *OperatorProductizedOptionsProvider) GetProductizedOptions() (*component.ProductizedOptions, error) {
	pob := component.ProductizedOptionsBuilder{}

	productVersion := o.APIManagerSpec.ProductVersion
	imageProvider, err := product.NewImageProvider(productVersion)
	if err != nil {
		return nil, err
	}

	pob.AmpRelease(string(productVersion))

	if o.APIManagerSpec.ApicastSpec != nil && o.APIManagerSpec.ApicastSpec.Image != nil {
		pob.ApicastImage(*o.APIManagerSpec.ApicastSpec.Image)
	} else {
		pob.ApicastImage(imageProvider.GetApicastImage())
	}

	if o.APIManagerSpec.BackendSpec != nil && o.APIManagerSpec.BackendSpec.Image != nil {
		pob.BackendImage(*o.APIManagerSpec.BackendSpec.Image)
	} else {
		pob.BackendImage(imageProvider.GetBackendImage())
	}

	if o.APIManagerSpec.WildcardRouterSpec != nil && o.APIManagerSpec.WildcardRouterSpec.Image != nil {
		pob.RouterImage(*o.APIManagerSpec.WildcardRouterSpec.Image)
	} else {
		pob.RouterImage(imageProvider.GetWildcardRouterImage())
	}

	if o.APIManagerSpec.SystemSpec != nil && o.APIManagerSpec.SystemSpec.Image != nil {
		pob.SystemImage(*o.APIManagerSpec.SystemSpec.Image)
	} else {
		pob.SystemImage(imageProvider.GetSystemImage())
	}

	if o.APIManagerSpec.ZyncSpec != nil && o.APIManagerSpec.ZyncSpec.Image != nil {
		pob.ZyncImage(*o.APIManagerSpec.ZyncSpec.Image)
	} else {
		pob.ZyncImage(imageProvider.GetZyncImage())
	}

	res, err := pob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Productized Options - %s", err)
	}
	return res, nil
}
