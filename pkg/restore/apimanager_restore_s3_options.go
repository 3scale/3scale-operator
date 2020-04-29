package restore

import (
	validator "github.com/go-playground/validator/v10"
)

type APIManagerRestoreS3Options struct {
	Bucket         string                          `validate:"required"`
	Credentials    RestoreDestinationS3Credentials `validate:"required"`
	Region         *string
	Endpoint       *string
	Path           *string
	ForcePathStyle *bool
}

type RestoreDestinationS3Credentials struct {
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}

func NewAPIManagerRestoreS3Options() *APIManagerRestoreS3Options {
	return &APIManagerRestoreS3Options{}
}

func (a *APIManagerRestoreS3Options) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
