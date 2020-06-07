package restore

import (
	validator "github.com/go-playground/validator/v10"
	"k8s.io/apimachinery/pkg/types"
)

type APIManagerRestoreOptions struct {
	Namespace             string    `validate:"required"` // Namespace where the K8s related objects to the restore will be created/looked
	APIManagerRestoreName string    `validate:"required"` // Name of the APIManagerRestore CR. NOT the backup or APIManager name
	APIManagerRestoreUID  types.UID `validate:"required"` // UID of the APIManagerRestore CR

	APIManagerRestorePVCOptions *APIManagerRestorePVCOptions `validate:"required"`
	OCCLIImageURL               string                       `validate:"required"`
}

func NewAPIManagerRestoreOptions() *APIManagerRestoreOptions {
	return &APIManagerRestoreOptions{}
}

func (a *APIManagerRestoreOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
