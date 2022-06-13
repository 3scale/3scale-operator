package component

import "github.com/go-playground/validator/v10"

type HighAvailabilityOptions struct {
	BackendRedisQueuesEndpoint       string
	BackendRedisQueuesSentinelHosts  string
	BackendRedisQueuesSentinelRole   string
	BackendRedisStorageEndpoint      string
	BackendRedisStorageSentinelHosts string
	BackendRedisStorageSentinelRole  string
	SystemDatabaseURL                string
	SystemRedisURL                   string
	SystemRedisSentinelsHosts        string
	SystemRedisSentinelsRole         string
	SystemRedisNamespace             string

	BackendRedisLabels   map[string]string `validate:"required"`
	SystemRedisLabels    map[string]string `validate:"required"`
	SystemDatabaseLabels map[string]string `validate:"required"`
}

func NewHighAvailabilityOptions() *HighAvailabilityOptions {
	return &HighAvailabilityOptions{}
}

func (h *HighAvailabilityOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(h)
}
