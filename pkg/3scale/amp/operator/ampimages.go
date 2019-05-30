package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
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
	if o.APIManagerSpec.Apicast != nil && o.APIManagerSpec.Apicast.Image != nil {
		optProv.ApicastImage(*o.APIManagerSpec.Apicast.Image)
	} else {
		optProv.ApicastImage(imageProvider.GetApicastImage())
	}

	if o.APIManagerSpec.Backend != nil && o.APIManagerSpec.Backend.Image != nil {
		optProv.BackendImage(*o.APIManagerSpec.Backend.Image)
	} else {
		optProv.BackendImage(imageProvider.GetBackendImage())
	}

	if o.APIManagerSpec.WildcardRouter != nil && o.APIManagerSpec.WildcardRouter.Image != nil {
		optProv.RouterImage(*o.APIManagerSpec.WildcardRouter.Image)
	} else {
		optProv.RouterImage(imageProvider.GetWildcardRouterImage())
	}

	if o.APIManagerSpec.System != nil && o.APIManagerSpec.System.Image != nil {
		optProv.SystemImage(*o.APIManagerSpec.System.Image)
	} else {
		optProv.SystemImage(imageProvider.GetSystemImage())
	}

	if o.APIManagerSpec.Zync != nil && o.APIManagerSpec.Zync.Image != nil {
		optProv.ZyncImage(*o.APIManagerSpec.Zync.Image)
	} else {
		optProv.ZyncImage(imageProvider.GetZyncImage())
	}

	if o.APIManagerSpec.Zync != nil && o.APIManagerSpec.Zync.PostgreSQLImage != nil {
		optProv.PostgreSQLImage(*o.APIManagerSpec.Zync.PostgreSQLImage)
	} else {
		optProv.PostgreSQLImage(imageProvider.GetZyncPostgreSQLImage())
	}

	if o.APIManagerSpec.Backend != nil && o.APIManagerSpec.Backend.RedisImage != nil {
		optProv.BackendRedisImage(*o.APIManagerSpec.Backend.RedisImage)
	} else {
		optProv.BackendRedisImage(imageProvider.GetBackendRedisImage())
	}

	if o.APIManagerSpec.System != nil && o.APIManagerSpec.System.RedisImage != nil {
		optProv.SystemRedisImage(*o.APIManagerSpec.System.RedisImage)
	} else {
		optProv.SystemRedisImage(imageProvider.GetSystemRedisImage())
	}

	if o.APIManagerSpec.System != nil && o.APIManagerSpec.System.MemcachedImage != nil {
		optProv.SystemMemcachedImage(*o.APIManagerSpec.System.MemcachedImage)
	} else {
		optProv.SystemMemcachedImage(imageProvider.GetSystemMemcachedImage())
	}

	optProv.InsecureImportPolicy(*o.APIManagerSpec.ImageStreamTagImportInsecure)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create AMPImages Options - %s", err)
	}
	return res, nil
}
