package backup

import (
	validator "github.com/go-playground/validator/v10"
	"k8s.io/apimachinery/pkg/api/resource"
)

type APIManagerBackupPVCOptions struct {
	BackupDestinationPVC BackupDestinationPVC `validate:"required"`
}

type BackupDestinationPVC struct {
	Name            string `validate:"required"`
	StorageClass    *string
	VolumeName      *string
	StorageRequests *resource.Quantity // TODO should we validate resource.Quantity in case we define it?
}

func NewAPIManagerBackupPVCOptions() *APIManagerBackupPVCOptions {
	return &APIManagerBackupPVCOptions{}
}

func (a *APIManagerBackupPVCOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
