package component

import "github.com/go-playground/validator/v10"

type SystemPostgreSQLImageOptions struct {
	AppLabel             string `validate:"required"`
	AmpRelease           string `validate:"required"`
	Image                string `validate:"required"`
	InsecureImportPolicy *bool  `validate:"required"`
}

func NewSystemPostgreSQLImageOptions() *SystemPostgreSQLImageOptions {
	return &SystemPostgreSQLImageOptions{}
}

func (s *SystemPostgreSQLImageOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}
