package component

import "fmt"

type RedisOptionsBuilder struct {
	options RedisOptions
}

func (r *RedisOptionsBuilder) AppLabel(appLabel string) {
	r.options.appLabel = appLabel
}

func (r *RedisOptionsBuilder) BackendImage(image string) {
	r.options.backendImage = image
}

func (r *RedisOptionsBuilder) SystemImage(image string) {
	r.options.systemImage = image
}

func (r *RedisOptionsBuilder) BackendMemoryLimit(memoryLimit string) {
	r.options.backendMemoryLimit = memoryLimit
}
func (r *RedisOptionsBuilder) SystemMemoryLimit(memoryLimit string) {
	r.options.systemMemoryLimit = memoryLimit
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
	if r.options.backendImage == "" {
		return fmt.Errorf("no Redis Backend Image has been provided")
	}

	if r.options.systemImage == "" {
		return fmt.Errorf("no System Redis Image has been provided")
	}

	if r.options.backendMemoryLimit == "" {
		return fmt.Errorf("No Backend Redis Memory Limit has been provided")
	}
	if r.options.systemMemoryLimit == "" {
		return fmt.Errorf("No Redis System Memory Limit has been provided")
	}

	return nil
}

func (r *RedisOptionsBuilder) setNonRequiredOptions() {

}
