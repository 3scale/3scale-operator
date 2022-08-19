package component

import "github.com/go-playground/validator/v10"

type S3Options struct {
	AwsAccessKeyId       string `validate:"required"`
	AwsSecretAccessKey   string `validate:"required"`
	AwsRegion            string `validate:"required"`
	AwsBucket            string `validate:"required"`
	AwsProtocol          string `validate:"required"`
	AwsHostname          string `validate:"required"`
	AwsPathStyle         string `validate:"required"`
	AwsCredentialsSecret string `validate:"required"`
}

func NewS3Options() *S3Options {
	return &S3Options{}
}

func (s3 *S3Options) Validate() error {
	validate := validator.New()
	return validate.Struct(s3)
}
