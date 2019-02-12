package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	oprand "github.com/3scale/3scale-operator/pkg/crypto/rand"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (o *OperatorZyncOptionsProvider) GetZyncOptions() (*component.ZyncOptions, error) {
	optProv := component.ZyncOptionsBuilder{}
	optProv.AppLabel(*o.APIManagerSpec.AppLabel)
	o.setSecretBasedOptions(&optProv)

	err := o.setZyncSecretOptions(&optProv)
	if err != nil {
		return nil, err
	}

	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Zync Options - %s", err)
	}
	return res, nil
}

func (o *OperatorZyncOptionsProvider) setSecretBasedOptions(zob *component.ZyncOptionsBuilder) error {
	err := o.setZyncSecretOptions(zob)
	if err != nil {
		return fmt.Errorf("unable to create Zync Secret Options - %s", err)
	}

	return nil
}

func (o *OperatorZyncOptionsProvider) setZyncSecretOptions(zob *component.ZyncOptionsBuilder) error {
	defaultZyncSecretKeyBase := oprand.String(16)
	defaultZyncDatabasePassword := oprand.String(16)
	defaultZyncAuthenticationToken := oprand.String(16)

	currSecret, err := getSecret(component.ZyncSecretName, o.Namespace, o.Client)
	if err != nil {
		if errors.IsNotFound(err) {
			// Set options defaults
			zob.SecretKeyBase(defaultZyncSecretKeyBase)
			zob.DatabasePassword(defaultZyncDatabasePassword)
			zob.AuthenticationToken(defaultZyncAuthenticationToken)
		} else {
			return err
		}
	} else {
		// If a field of a secret already exists in the deployed secret then
		// We do not modify it. Otherwise we set a default value
		secretData := currSecret.Data
		zob.SecretKeyBase(getSecretDataValueOrDefault(secretData, component.ZyncSecretKeyBaseFieldName, defaultZyncSecretKeyBase))
		zob.DatabasePassword(getSecretDataValueOrDefault(secretData, component.ZyncSecretDatabasePasswordFieldName, defaultZyncDatabasePassword))
		zob.AuthenticationToken(getSecretDataValueOrDefault(secretData, component.ZyncSecretAuthenticationTokenFieldName, defaultZyncAuthenticationToken))
	}
	return nil
}
