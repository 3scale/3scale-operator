package component

import (
	"github.com/go-playground/validator/v10"
)

type RedisConfigMapOptions struct {
	Labels    map[string]string `validate:"required"`
	Namespace string
}

func NewRedisConfigMapOptions() *RedisConfigMapOptions {
	return &RedisConfigMapOptions{}
}

func (r *RedisConfigMapOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
