package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorMysqlOptionsProvider) GetMysqlOptions() (*component.MysqlOptions, error) {
	optProv := component.MysqlOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.DatabaseName(*o.APIManagerSpec.MysqlDatabase)
	optProv.Image(*o.APIManagerSpec.MysqlImage)
	optProv.User(*o.APIManagerSpec.MysqlUser)
	optProv.Password(*o.APIManagerSpec.MysqlPassword)
	optProv.RootPassword(*o.APIManagerSpec.MysqlRootPassword)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Mysql Options - %s", err)
	}
	return res, nil
}
