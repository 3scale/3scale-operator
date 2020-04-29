package restore

import (
	validator "github.com/go-playground/validator/v10"
)

type APIManagerRestoreOptions struct {
	Namespace                   string                       `validate:"required"` // Namespace where the K8s related objects to the restore will be created/looked
	APIManagerRestoreName       string                       `validate:"required"` // Name of the APIManagerRestore CR. NOT the backup or APIManager name
	APIManagerRestorePVCOptions *APIManagerRestorePVCOptions `validate:"required_without=APIManagerRestoreS3Options"`
	APIManagerRestoreS3Options  *APIManagerRestoreS3Options  `validate:"required_without=APIManagerRestorePVCOptions"`
}

func NewAPIManagerRestoreOptions() *APIManagerRestoreOptions {
	return &APIManagerRestoreOptions{}
}

func (a *APIManagerRestoreOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
