package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorMysqlOptionsProvider) GetMysqlOptions() (*component.MysqlOptions, error) {
	optProv := component.MysqlOptionsBuilder{}
	optProv.AppLabel(*o.AmpSpec.AppLabel)
	optProv.DatabaseName(*o.AmpSpec.MysqlDatabase)
	optProv.Image(*o.AmpSpec.MysqlImage)
	optProv.User(*o.AmpSpec.MysqlUser)
	optProv.Password(*o.AmpSpec.MysqlPassword)
	optProv.RootPassword(*o.AmpSpec.MysqlRootPassword)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Mysql Options - %s", err)
	}
	return res, nil
}
