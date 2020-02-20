package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
)

func (o *OperatorSystemMySQLImageOptionsProvider) GetSystemMySQLImageOptions() (*component.SystemMySQLImageOptions, error) {
	optProv := component.SystemMySQLImageOptionsBuilder{}

	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.AmpRelease(product.ThreescaleRelease)
	if o.APIManagerSpec.System.DatabaseSpec != nil &&
		o.APIManagerSpec.System.DatabaseSpec.MySQL != nil &&
		o.APIManagerSpec.System.DatabaseSpec.MySQL.Image != nil {
		optProv.Image(*o.APIManagerSpec.System.DatabaseSpec.MySQL.Image)
	} else {
		optProv.Image(SystemMySQLImageURL())
	}
	optProv.InsecureImportPolicy(*o.APIManagerSpec.ImageStreamTagImportInsecure)

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Mysql Image Options - %s", err)
	}
	return res, nil
}
