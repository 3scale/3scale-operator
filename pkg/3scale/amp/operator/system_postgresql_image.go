package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
)

func (o *OperatorSystemPostgreSQLImageOptionsProvider) GetSystemPostgreSQLImageOptions() (*component.SystemPostgreSQLImageOptions, error) {
	optProv := component.SystemPostgreSQLImageOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.AmpRelease(product.ThreescaleRelease)
	if o.APIManagerSpec.System.DatabaseSpec != nil &&
		o.APIManagerSpec.System.DatabaseSpec.PostgreSQL != nil &&
		o.APIManagerSpec.System.DatabaseSpec.PostgreSQL.Image != nil {
		optProv.Image(*o.APIManagerSpec.System.DatabaseSpec.PostgreSQL.Image)
	} else {
		optProv.Image(SystemPostgreSQLImageURL())
	}
	optProv.InsecureImportPolicy(*o.APIManagerSpec.ImageStreamTagImportInsecure)

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create PostgreSQL Image Options - %s", err)
	}
	return res, nil
}
