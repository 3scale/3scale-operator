package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
)

func (o *OperatorProductizedOptionsProvider) GetProductizedOptions() (*component.ProductizedOptions, error) {
	pob := component.ProductizedOptionsBuilder{}

	productVersion := product.CurrentProductVersion()
	imageProvider, err := product.NewImageProvider(productVersion)
	if err != nil {
		return nil, err
	}

	pob.AmpRelease(string(productVersion))

	if o.APIManagerSpec.Apicast != nil && o.APIManagerSpec.Apicast.Image != nil {
		pob.ApicastImage(*o.APIManagerSpec.Apicast.Image)
	} else {
		pob.ApicastImage(imageProvider.GetApicastImage())
	}

	if o.APIManagerSpec.Backend != nil && o.APIManagerSpec.Backend.Image != nil {
		pob.BackendImage(*o.APIManagerSpec.Backend.Image)
	} else {
		pob.BackendImage(imageProvider.GetBackendImage())
	}

	if o.APIManagerSpec.WildcardRouter != nil && o.APIManagerSpec.WildcardRouter.Image != nil {
		pob.RouterImage(*o.APIManagerSpec.WildcardRouter.Image)
	} else {
		pob.RouterImage(imageProvider.GetWildcardRouterImage())
	}

	if o.APIManagerSpec.System != nil && o.APIManagerSpec.System.Image != nil {
		pob.SystemImage(*o.APIManagerSpec.System.Image)
	} else {
		pob.SystemImage(imageProvider.GetSystemImage())
	}

	if o.APIManagerSpec.Zync != nil && o.APIManagerSpec.Zync.Image != nil {
		pob.ZyncImage(*o.APIManagerSpec.Zync.Image)
	} else {
		pob.ZyncImage(imageProvider.GetZyncImage())
	}

	res, err := pob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Productized Options - %s", err)
	}
	return res, nil
}
