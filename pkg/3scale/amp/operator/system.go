package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorSystemOptionsProvider) GetSystemOptions() (*component.SystemOptions, error) {
	optProv := component.SystemOptionsBuilder{}
	optProv.AdminAccessToken(*o.APIManagerSpec.AdminAccessToken)
	optProv.AdminPassword(*o.APIManagerSpec.AdminPassword)
	optProv.AdminUsername(*o.APIManagerSpec.AdminUsername)
	optProv.AmpRelease(o.APIManagerSpec.AmpRelease)
	optProv.ApicastAccessToken(*o.APIManagerSpec.ApicastAccessToken)
	optProv.ApicastRegistryURL(*o.APIManagerSpec.ApicastRegistryURL)
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.MasterAccessToken(*o.APIManagerSpec.MasterAccessToken)
	optProv.MasterName(*o.APIManagerSpec.MasterName)
	optProv.MasterUsername(*o.APIManagerSpec.MasterUser)
	optProv.MasterPassword(*o.APIManagerSpec.MasterPassword)
	optProv.RecaptchaPublicKey(*o.APIManagerSpec.RecaptchaPublicKey)
	optProv.RecaptchaPrivateKey(*o.APIManagerSpec.RecaptchaPrivateKey)
	optProv.AppSecretKeyBase(*o.APIManagerSpec.SystemAppSecretKeyBase)
	optProv.BackendSharedSecret(*o.APIManagerSpec.SystemBackendSharedSecret)
	optProv.TenantName(*o.APIManagerSpec.TenantName)
	optProv.WildcardDomain(o.APIManagerSpec.WildcardDomain)
	optProv.MysqlDatabaseName(*o.APIManagerSpec.MysqlDatabase)
	optProv.MysqlRootPassword(*o.APIManagerSpec.MysqlRootPassword)
	optProv.StorageClassName(o.APIManagerSpec.RwxStorageClass)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create System Options - %s", err)
	}
	return res, nil
}
