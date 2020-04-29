package restore

import (
	validator "github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
)

type APIManagerRestorePVCOptions struct {
	PersistentVolumeClaimVolumeSource v1.PersistentVolumeClaimVolumeSource `validate:"required"`
}

func NewAPIManagerRestorePVCOptions() *APIManagerRestorePVCOptions {
	return &APIManagerRestorePVCOptions{}
}

func (a *APIManagerRestorePVCOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
