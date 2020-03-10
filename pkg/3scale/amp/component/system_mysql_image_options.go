package component

import "github.com/go-playground/validator/v10"

type SystemMySQLImageOptions struct {
	AppLabel             string `validate:"required"`
	AmpRelease           string `validate:"required"`
	Image                string `validate:"required"`
	InsecureImportPolicy *bool  `validate:"required"`
}

func NewSystemMySQLImageOptions() *SystemMySQLImageOptions {
	return &SystemMySQLImageOptions{}
}

func (s *SystemMySQLImageOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}
