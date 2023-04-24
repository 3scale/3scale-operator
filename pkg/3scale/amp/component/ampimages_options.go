package component

import (
	"github.com/go-playground/validator/v10"
	v1 "k8s.io/api/core/v1"
)

// AmpImagesOptions container object with all required to create components
type AmpImagesOptions struct {
	AppLabel                    string `validate:"required"`
	AmpRelease                  string `validate:"required"`
	ApicastImage                string `validate:"required"`
	BackendImage                string `validate:"required"`
	SystemImage                 string `validate:"required"`
	ZyncImage                   string `validate:"required"`
	ZyncDatabasePostgreSQLImage string `validate:"required"`
	SystemMemcachedImage        string `validate:"required"`
	SystemSearchdImage          string `validate:"required"`
	InsecureImportPolicy        bool
	ImagePullSecrets            []v1.LocalObjectReference `validate:"required"`
}

func NewAmpImagesOptions() *AmpImagesOptions {
	return &AmpImagesOptions{}
}

func (a *AmpImagesOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

func AmpImagesDefaultImagePullSecrets() []v1.LocalObjectReference {
	return []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "threescale-registry-auth"},
	}
}
