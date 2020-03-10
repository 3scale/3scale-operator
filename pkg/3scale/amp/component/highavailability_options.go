package component

import "github.com/go-playground/validator/v10"

type HighAvailabilityOptions struct {
	AppLabel                            string `validate:"required"`
	BackendRedisQueuesEndpoint          string `validate:"required"`
	BackendRedisQueuesSentinelHosts     string `validate:"required"`
	BackendRedisQueuesSentinelRole      string `validate:"required"`
	BackendRedisStorageEndpoint         string `validate:"required"`
	BackendRedisStorageSentinelHosts    string `validate:"required"`
	BackendRedisStorageSentinelRole     string `validate:"required"`
	SystemDatabaseURL                   string `validate:"required"`
	SystemRedisURL                      string `validate:"required"`
	SystemRedisSentinelsHosts           string `validate:"required"`
	SystemRedisSentinelsRole            string `validate:"required"`
	SystemMessageBusRedisURL            string `validate:"required"`
	SystemMessageBusRedisSentinelsHosts string `validate:"required"`
	SystemMessageBusRedisSentinelsRole  string `validate:"required"`
}

func NewHighAvailabilityOptions() *HighAvailabilityOptions {
	return &HighAvailabilityOptions{}
}

func (h *HighAvailabilityOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(h)
}
