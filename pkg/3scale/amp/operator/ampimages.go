package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorAmpImagesOptionsProvider) GetAmpImagesOptions() (*component.AmpImagesOptions, error) {
	optProv := component.AmpImagesOptionsBuilder{}

	productVersion := o.APIManagerSpec.ProductVersion
	imageProvider, err := product.NewImageProvider(productVersion)
	if err != nil {
		return nil, err
	}

	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.AMPRelease(string(productVersion))
	if o.APIManagerSpec.ApicastSpec != nil && o.APIManagerSpec.ApicastSpec.Image != nil {
		optProv.ApicastImage(*o.APIManagerSpec.ApicastSpec.Image)
	} else {
		optProv.ApicastImage(imageProvider.GetApicastImage())
	}

	if o.APIManagerSpec.BackendSpec != nil && o.APIManagerSpec.BackendSpec.Image != nil {
		optProv.BackendImage(*o.APIManagerSpec.BackendSpec.Image)
	} else {
		optProv.BackendImage(imageProvider.GetBackendImage())
	}

	if o.APIManagerSpec.WildcardRouterSpec != nil && o.APIManagerSpec.WildcardRouterSpec.Image != nil {
		optProv.RouterImage(*o.APIManagerSpec.WildcardRouterSpec.Image)
	} else {
		optProv.RouterImage(imageProvider.GetWildcardRouterImage())
	}

	if o.APIManagerSpec.SystemSpec != nil && o.APIManagerSpec.SystemSpec.Image != nil {
		optProv.SystemImage(*o.APIManagerSpec.SystemSpec.Image)
	} else {
		optProv.SystemImage(imageProvider.GetSystemImage())
	}

	if o.APIManagerSpec.ZyncSpec != nil && o.APIManagerSpec.ZyncSpec.Image != nil {
		optProv.ZyncImage(*o.APIManagerSpec.ZyncSpec.Image)
	} else {
		optProv.ZyncImage(imageProvider.GetZyncImage())
	}

	if o.APIManagerSpec.ZyncSpec != nil && o.APIManagerSpec.ZyncSpec.PostgreSQLImage != nil {
		optProv.PostgreSQLImage(*o.APIManagerSpec.ZyncSpec.PostgreSQLImage)
	} else {
		optProv.PostgreSQLImage(imageProvider.GetZyncPostgreSQLImage())
	}

	if o.APIManagerSpec.BackendSpec != nil && o.APIManagerSpec.BackendSpec.RedisImage != nil {
		optProv.BackendRedisImage(*o.APIManagerSpec.BackendSpec.RedisImage)
	} else {
		optProv.BackendRedisImage(imageProvider.GetBackendRedisImage())
	}

	if o.APIManagerSpec.SystemSpec != nil && o.APIManagerSpec.SystemSpec.RedisImage != nil {
		optProv.SystemRedisImage(*o.APIManagerSpec.SystemSpec.RedisImage)
	} else {
		optProv.SystemRedisImage(imageProvider.GetSystemRedisImage())
	}

	optProv.InsecureImportPolicy(*o.APIManagerSpec.ImageStreamTagImportInsecure)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create AMPImages Options - %s", err)
	}
	return res, nil
}
