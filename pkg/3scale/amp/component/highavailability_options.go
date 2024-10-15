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
	SystemDatabaseSslCa              string
	SystemDatabaseSslMode            string
	SystemDatabaseSslCert            string
	SystemDatabaseSslKey             string
	SystemRedisURL                   string
	SystemRedisSentinelsHosts        string
	SystemRedisSentinelsRole         string
	ZyncDatabaseURL                  string
	ZyncDatabasePassword             string
	ZyncDatabaseSslCa                string
	ZyncDatabaseSslMode              string
	ZyncDatabaseSslCert              string
	ZyncDatabaseSslKey               string
	SystemRedisSslCa                 string
	SystemRedisSslCert               string
	SystemRedisSslKey                string
	BackendRedisSslCa                string
	BackendRedisSslCert              string
	BackendRedisSslKey               string
	BackendRedisQueuesSslCa          string
	BackendRedisQueuesSslCert        string
	BackendRedisQueuesSslKey         string
}

func NewHighAvailabilityOptions() *HighAvailabilityOptions {
	return &HighAvailabilityOptions{}
}

func (h *HighAvailabilityOptions) Validate() error {
	validate := validator.New()
	return validate.Struct(h)
}
