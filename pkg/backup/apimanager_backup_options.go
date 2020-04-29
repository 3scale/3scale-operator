package backup

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	validator "github.com/go-playground/validator/v10"
)

type APIManagerBackupOptions struct {
	Namespace                  string                      `validate:"required"` // Namespace where the K8s related objects to the backup will be created/looked
	APIManagerBackupName       string                      `validate:"required"` // Name of the APIManager CR. NOT the APIManager cr name
	APIManagerName             string                      `validate:"required"` // Name of the APIManager CR. NOT the backup cr name
	APIManager                 *appsv1alpha1.APIManager    `validate:"required"` // Should we make this required?
	APIManagerBackupPVCOptions *APIManagerBackupPVCOptions `validate:"required"`
}

func NewAPIManagerBackupOptions() *APIManagerBackupOptions {
	return &APIManagerBackupOptions{}
}

func (a *APIManagerBackupOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
