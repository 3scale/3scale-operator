package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
)

func (o *OperatorSystemMySQLImageOptionsProvider) GetSystemMySQLImageOptions() (*component.SystemMySQLImageOptions, error) {
	optProv := component.SystemMySQLImageOptionsBuilder{}
	productVersion := o.APIManagerSpec.ProductVersion
	imageProvider, err := product.NewImageProvider(productVersion)
	if err != nil {
		return nil, err
	}

	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.AmpRelease(string(productVersion))
	if o.APIManagerSpec.System.DatabaseSpec.MySQL.Image != nil {
		optProv.Image(*o.APIManagerSpec.System.DatabaseSpec.MySQL.Image)
	} else {
		optProv.Image(imageProvider.GetSystemMySQLImage())
	}
	optProv.InsecureImportPolicy(*o.APIManagerSpec.ImageStreamTagImportInsecure)

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Mysql Image Options - %s", err)
	}
	return res, nil
}
