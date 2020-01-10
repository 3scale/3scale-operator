package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
)

func (o *OperatorS3OptionsProvider) GetS3Options() (*component.S3Options, error) {
	sob := component.S3OptionsBuilder{}
	SystemS3Spec := *o.APIManagerSpec.System.FileStorageSpec.S3
	sob.AwsRegion(SystemS3Spec.AWSRegion)
	sob.AwsBucket(SystemS3Spec.AWSBucket)
	sob.StorageServiceEndpointScheme(helper.GetStringPointerValueOrDefault(SystemS3Spec.EndpointScheme, ""))
	sob.StorageServiceEndpointHost(helper.GetStringPointerValueOrDefault(SystemS3Spec.EndpointHost, ""))

	err := o.setSecretBasedOptions(&sob)
	if err != nil {
		return nil, err
	}

	res, err := sob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create S3 Options - %s", err)
	}
	return res, nil
}

func (o *OperatorS3OptionsProvider) setSecretBasedOptions(sob *component.S3OptionsBuilder) error {
	err := o.setAWSSecretOptions(sob)
	if err != nil {
		return fmt.Errorf("unable to create S3 Secret Options - %s", err)
	}

	return nil
}

func (o *OperatorS3OptionsProvider) setAWSSecretOptions(sob *component.S3OptionsBuilder) error {
	awsCredentialsSecretName := o.APIManagerSpec.System.FileStorageSpec.S3.AWSCredentials.Name
	currSecret, err := helper.GetSecret(awsCredentialsSecretName, o.Namespace, o.Client)
	if err != nil {
		return err
	}

	// If a field of a secret already exists in the deployed secret then
	// We do not modify it. Otherwise we set a default value
	secretData := currSecret.Data
	var result *string
	result = helper.GetSecretDataValue(secretData, component.S3SecretAWSAccessKeyIdFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.S3SecretAWSAccessKeyIdFieldName, awsCredentialsSecretName)
	}
	sob.AwsAccessKeyId(*result)

	result = helper.GetSecretDataValue(secretData, component.S3SecretAWSSecretAccessKeyFieldName)
	if result == nil {
		return fmt.Errorf("Secret field '%s' is required in secret '%s'", component.S3SecretAWSSecretAccessKeyFieldName, awsCredentialsSecretName)
	}
	sob.AwsSecretAccessKey(*result)

	sob.AWSCredentialsSecret(awsCredentialsSecretName)

	return nil
}
