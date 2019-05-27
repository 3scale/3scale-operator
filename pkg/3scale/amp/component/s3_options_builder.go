package component

import "fmt"

type S3Options struct {
	// s3NonRequiredOptions

	// s3RequiredOptions
	awsAccessKeyId       string
	awsSecretAccessKey   string
	awsRegion            string
	awsBucket            string
	awsCredentialsSecret string
}

type S3OptionsBuilder struct {
	options S3Options
}

func (s3 *S3OptionsBuilder) AwsAccessKeyId(awsAccessKeyId string) {
	s3.options.awsAccessKeyId = awsAccessKeyId
}

func (s3 *S3OptionsBuilder) AwsSecretAccessKey(awsSecretAccessKey string) {
	s3.options.awsSecretAccessKey = awsSecretAccessKey
}

func (s3 *S3OptionsBuilder) AwsRegion(awsRegion string) {
	s3.options.awsRegion = awsRegion
}

func (s3 *S3OptionsBuilder) AwsBucket(awsBucket string) {
	s3.options.awsBucket = awsBucket
}

func (s3 *S3OptionsBuilder) AWSCredentialsSecret(awsCredentials string) {
	s3.options.awsCredentialsSecret = awsCredentials
}

func (s3 *S3OptionsBuilder) Build() (*S3Options, error) {
	err := s3.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	s3.setNonRequiredOptions()

	return &s3.options, nil

}

func (s3 *S3OptionsBuilder) setRequiredOptions() error {
	if s3.options.awsAccessKeyId == "" {
		return fmt.Errorf("no AWS access key id has been provided")
	}
	if s3.options.awsSecretAccessKey == "" {
		return fmt.Errorf("no AWS secret access key has been provided")
	}
	if s3.options.awsRegion == "" {
		return fmt.Errorf("no AWS region has been provided")
	}
	if s3.options.awsBucket == "" {
		return fmt.Errorf("no AWS bucket has been provided")
	}
	if s3.options.awsCredentialsSecret == "" {
		return fmt.Errorf("no AWS credentials secret has been provided")
	}

	return nil
}

func (s3 *S3OptionsBuilder) setNonRequiredOptions() {

}
