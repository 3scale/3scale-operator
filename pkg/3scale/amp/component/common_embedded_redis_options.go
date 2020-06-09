package component

import (
	"github.com/go-playground/validator/v10"
)

type CommonEmbeddedRedisOptions struct {
	ConfigMapLabels map[string]string `validate:"required"`
}

func NewCommonEmbeddedRedisOptions() *CommonEmbeddedRedisOptions {
	return &CommonEmbeddedRedisOptions{}
}

func (r *CommonEmbeddedRedisOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
