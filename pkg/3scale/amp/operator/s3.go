package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorS3OptionsProvider) GetS3Options() (*component.S3Options, error) {
	sob := component.S3OptionsBuilder{}
	sob.AwsRegion(*o.APIManagerSpec.AwsRegion)
	sob.AwsBucket(*o.APIManagerSpec.AwsBucket)
	sob.FileUploadStorage(*o.APIManagerSpec.FileUploadStorage)
	res, err := sob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create S3 Options - %s", err)
	}
	return res, nil
}

func (o *OperatorS3OptionsProvider) setSecretBasedOptions(sob *component.S3OptionsBuilder) error {
	err := o.setAWSSecretOptions(sob)
	if err != nil {
		return fmt.Errorf("unable to create Zync Secret Options - %s", err)
	}

	return nil
}

func (o *OperatorS3OptionsProvider) setAWSSecretOptions(sob *component.S3OptionsBuilder) error {
	currSecret, err := getSecret(component.S3SecretAWSSecretName, o.Namespace, o.Client)
	if err != nil {
		return err
	}

	// If a field of a secret already exists in the deployed secret then
	// We do not modify it. Otherwise we set a default value
	secretData := currSecret.Data
	var result *string
	result = getSecretDataValue(secretData, component.S3SecretAWSAccessKeyIdFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.S3SecretAWSAccessKeyIdFieldName, component.S3SecretAWSSecretName)
	}
	sob.AwsAccessKeyId(*result)

	result = getSecretDataValue(secretData, component.S3SecretAWSSecretAccessKeyFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.S3SecretAWSSecretAccessKeyFieldName, component.S3SecretAWSSecretName)
	}
	sob.AwsSecretAccessKey(*result)

	return nil
}
