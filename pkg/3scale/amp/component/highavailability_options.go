package component

import "github.com/go-playground/validator/v10"

type HighAvailabilityOptions struct {
	BackendRedisQueuesEndpoint          string `validate:"required"`
	BackendRedisQueuesSentinelHosts     string
	BackendRedisQueuesSentinelRole      string
	BackendRedisStorageEndpoint         string `validate:"required"`
	BackendRedisStorageSentinelHosts    string
	BackendRedisStorageSentinelRole     string
	SystemDatabaseURL                   string `validate:"required"`
	SystemRedisURL                      string `validate:"required"`
	SystemRedisSentinelsHosts           string
	SystemRedisSentinelsRole            string
	SystemRedisNamespace                string
	SystemMessageBusRedisURL            string `validate:"required"`
	SystemMessageBusRedisSentinelsHosts string
	SystemMessageBusRedisSentinelsRole  string
	SystemMessageBusRedisNamespace      string

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
