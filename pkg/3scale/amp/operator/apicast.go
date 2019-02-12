package operator

import (
	"fmt"
	"strconv"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (o *OperatorApicastOptionsProvider) GetApicastOptions() (*component.ApicastOptions, error) {
	optProv := component.ApicastOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	optProv.ManagementAPI(*o.APIManagerSpec.ApicastManagementApi)
	optProv.OpenSSLVerify(strconv.FormatBool(*o.APIManagerSpec.ApicastOpenSSLVerify)) // TODO is this a good place to make the conversion?
	optProv.ResponseCodes(strconv.FormatBool(*o.APIManagerSpec.ApicastResponseCodes)) // TODO is this a good place to make the conversion?
	optProv.TenantName(*o.APIManagerSpec.TenantName)
	optProv.WildcardDomain(o.APIManagerSpec.WildcardDomain)

	err := o.setSecretBasedOptions(&optProv)
	if err != nil {
		return nil, err
	}

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Apicast Options - %s", err)
	}
	return res, nil
}

func (o *OperatorApicastOptionsProvider) setSecretBasedOptions(aob *component.ApicastOptionsBuilder) error {
	err := o.setApicastRedisOptions(aob)
	if err != nil {
		return fmt.Errorf("unable to create Apicast Redis Secret Options - %s", err)
	}

	return nil
}

func (o *OperatorApicastOptionsProvider) setApicastRedisOptions(aob *component.ApicastOptionsBuilder) error {
	currSecret, err := getSecret(component.ZyncSecretName, o.Namespace, o.Client)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	} else {
		// If a field of a secret already exists in the deployed secret then
		// We do not modify it.
		secretData := currSecret.Data
		var result *string
		result = getSecretDataValue(secretData, component.ApicastSecretRedisProductionURLFieldName)
		if result != nil {
			aob.RedisProductionURL(*result)
		}
		result = getSecretDataValue(secretData, component.ApicastSecretRedisStagingURLFieldName)
		if result != nil {
			aob.RedisStagingURL(*result)
		}
	}
	return nil
}
