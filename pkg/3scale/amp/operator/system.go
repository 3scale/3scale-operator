package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorSystemOptionsProvider) GetSystemOptions() (*component.SystemOptions, error) {
	optProv := component.SystemOptionsBuilder{}
	optProv.AdminAccessToken(*o.AmpSpec.AdminAccessToken)
	optProv.AdminPassword(*o.AmpSpec.AdminPassword)
	optProv.AdminUsername(*o.AmpSpec.AdminUsername)
	optProv.AmpRelease(o.AmpSpec.AmpRelease)
	optProv.ApicastAccessToken(*o.AmpSpec.ApicastAccessToken)
	optProv.ApicastRegistryURL(*o.AmpSpec.ApicastRegistryURL)
	optProv.AppLabel(*o.AmpSpec.AppLabel)
	optProv.MasterAccessToken(*o.AmpSpec.MasterAccessToken)
	optProv.MasterName(*o.AmpSpec.MasterName)
	optProv.MasterUsername(*o.AmpSpec.MasterUser)
	optProv.MasterPassword(*o.AmpSpec.MasterPassword)
	optProv.RecaptchaPublicKey(*o.AmpSpec.RecaptchaPublicKey)
	optProv.RecaptchaPrivateKey(*o.AmpSpec.RecaptchaPrivateKey)
	optProv.AppSecretKeyBase(*o.AmpSpec.SystemAppSecretKeyBase)
	optProv.BackendSharedSecret(*o.AmpSpec.SystemBackendSharedSecret)
	optProv.TenantName(*o.AmpSpec.TenantName)
	optProv.WildcardDomain(o.AmpSpec.WildcardDomain)
	optProv.MysqlDatabaseName(*o.AmpSpec.MysqlDatabase)
	optProv.MysqlRootPassword(*o.AmpSpec.MysqlRootPassword)
	optProv.StorageClassName(o.AmpSpec.RwxStorageClass)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create System Options - %s", err)
	}
	return res, nil
}
