package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
)

func (o *OperatorRedisOptionsProvider) GetRedisOptions() (*component.RedisOptions, error) {
	optProv := component.RedisOptionsBuilder{}
	productVersion := o.APIManagerSpec.ProductVersion
	imageProvider, err := product.NewImageProvider(productVersion)
	if err != nil {
		return nil, err
	}

	optProv.AppLabel(*o.APIManagerSpec.AppLabel)

	if o.APIManagerSpec.BackendSpec != nil && o.APIManagerSpec.BackendSpec.RedisImage != nil {
		optProv.BackendImage(*o.APIManagerSpec.BackendSpec.RedisImage)
	} else {
		optProv.BackendImage(imageProvider.GetBackendRedisImage())
	}
	if o.APIManagerSpec.SystemSpec != nil && o.APIManagerSpec.SystemSpec.RedisImage != nil {
		optProv.SystemImage(*o.APIManagerSpec.SystemSpec.RedisImage)
	} else {
		optProv.SystemImage(imageProvider.GetSystemRedisImage())
	}

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Redis Options - %s", err)
	}
	return res, nil
}
