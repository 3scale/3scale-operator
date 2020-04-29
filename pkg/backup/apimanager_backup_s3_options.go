package backup

import (
	validator "github.com/go-playground/validator/v10"
)

const (
	BackupDestinationS3CredentialsSecretAccessKeyFieldName       = "AccessKeyID"
	BackupDestinationS3CredentialsSecretSecretAccessKeyFieldName = "SecretAccessKey"
)

// TODO add struct tags validation
type APIManagerBackupS3Options struct {
	BackupDestinationS3 BackupDestinationS3 `validate:"required"`
}

type BackupDestinationS3 struct {
	Bucket         string                         `validate:"required"`
	Credentials    BackupDestinationS3Credentials `validate:"required"`
	Region         *string
	Endpoint       *string
	Path           *string
	ForcePathStyle *bool
}

type BackupDestinationS3Credentials struct {
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}

func NewAPIManagerBackupS3Options() *APIManagerBackupS3Options {
	return &APIManagerBackupS3Options{}
}

func (a *APIManagerBackupS3Options) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
