package component

import "fmt"

type RedisOptions struct {
	appLabel string
}

type RedisOptionsBuilder struct {
	options RedisOptions
}

func (r *RedisOptionsBuilder) AppLabel(appLabel string) {
	r.options.appLabel = appLabel
}

func (r *RedisOptionsBuilder) Build() (*RedisOptions, error) {
	err := r.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	r.setNonRequiredOptions()

	return &r.options, nil
}

func (r *RedisOptionsBuilder) setRequiredOptions() error {
	if r.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}

	return nil
}

func (r *RedisOptionsBuilder) setNonRequiredOptions() {

}
