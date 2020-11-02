package backup

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	validator "github.com/go-playground/validator/v10"
	"k8s.io/apimachinery/pkg/types"
)

type APIManagerBackupOptions struct {
	Namespace                  string                      `validate:"required"` // Namespace where the K8s related objects to the backup will be created/looked
	APIManagerBackupName       string                      `validate:"required"` // Name of the APIManagerBackup CR. NOT the APIManager cr name
	APIManagerBackupUID        types.UID                   `validate:"required"` // UID of the APIManagerBackup CR
	APIManagerName             string                      `validate:"required"` // Name of the APIManager CR. NOT the APIManagerBackup cr name
	APIManager                 *appsv1alpha1.APIManager    `validate:"required"`
	APIManagerBackupPVCOptions *APIManagerBackupPVCOptions `validate:"required"`
	OCCLIImageURL              string                      `validate:"required"`
}

func NewAPIManagerBackupOptions() *APIManagerBackupOptions {
	return &APIManagerBackupOptions{}
}

func (a *APIManagerBackupOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
