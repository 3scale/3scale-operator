package component

import "fmt"

type HighAvailabilityOptions struct {
	// nonRequiredHighAvailabilityOptions

	//requiredHighAvailabilityOptions
	appLabel                            string
	backendRedisQueuesEndpoint          string
	backendRedisQueuesSentinelHosts     string
	backendRedisQueuesSentinelRole      string
	backendRedisStorageEndpoint         string
	backendRedisStorageSentinelHosts    string
	backendRedisStorageSentinelRole     string
	systemDatabaseURL                   string
	systemRedisURL                      string
	systemRedisSentinelsHosts           string
	systemRedisSentinelsRole            string
	systemMessageBusRedisURL            string
	systemMessageBusRedisSentinelsHosts string
	systemMessageBusRedisSentinelsRole  string
}

type HighAvailabilityOptionsBuilder struct {
	options HighAvailabilityOptions
}

func (ha *HighAvailabilityOptionsBuilder) AppLabel(appLabel string) {
	ha.options.appLabel = appLabel
}

func (ha *HighAvailabilityOptionsBuilder) BackendRedisQueuesEndpoint(backendRedisQueuesEndpoint string) {
	ha.options.backendRedisQueuesEndpoint = backendRedisQueuesEndpoint
}

func (ha *HighAvailabilityOptionsBuilder) BackendRedisStorageEndpoint(backendRedisStorageEndpoint string) {
	ha.options.backendRedisStorageEndpoint = backendRedisStorageEndpoint
}

func (ha *HighAvailabilityOptionsBuilder) SystemDatabaseURL(systemDatabaseURL string) {
	ha.options.systemDatabaseURL = systemDatabaseURL
}

func (ha *HighAvailabilityOptionsBuilder) SystemRedisURL(systemRedisURL string) {
	ha.options.systemRedisURL = systemRedisURL
}

func (ha *HighAvailabilityOptionsBuilder) SystemMessageBusRedisURL(url string) {
	ha.options.systemMessageBusRedisURL = url
}

func (ha *HighAvailabilityOptionsBuilder) SystemRedisSentinelsHosts(hosts string) {
	ha.options.systemRedisSentinelsHosts = hosts
}

func (ha *HighAvailabilityOptionsBuilder) SystemRedisSentinelsRole(role string) {
	ha.options.systemRedisSentinelsRole = role
}

func (ha *HighAvailabilityOptionsBuilder) SystemMessageBusRedisSentinelsHosts(hosts string) {
	ha.options.systemMessageBusRedisSentinelsHosts = hosts
}

func (ha *HighAvailabilityOptionsBuilder) SystemMessageBusRedisSentinelsRole(role string) {
	ha.options.systemMessageBusRedisSentinelsRole = role
}

func (ha *HighAvailabilityOptionsBuilder) BackendRedisQueuesSentinelHosts(hosts string) {
	ha.options.backendRedisQueuesSentinelHosts = hosts
}

func (ha *HighAvailabilityOptionsBuilder) BackendRedisQueuesSentinelRole(role string) {
	ha.options.backendRedisQueuesSentinelRole = role
}

func (ha *HighAvailabilityOptionsBuilder) BackendRedisStorageSentinelHosts(hosts string) {
	ha.options.backendRedisStorageSentinelHosts = hosts
}

func (ha *HighAvailabilityOptionsBuilder) BackendRedisStorageSentinelRole(role string) {
	ha.options.backendRedisStorageSentinelRole = role
}

func (ha *HighAvailabilityOptionsBuilder) Build() (*HighAvailabilityOptions, error) {

	err := ha.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	ha.setNonRequiredOptions()

	return &ha.options, nil
}

func (ha *HighAvailabilityOptionsBuilder) setRequiredOptions() error {
	if ha.options.backendRedisQueuesEndpoint == "" {
		return fmt.Errorf("no Backend Redis queues endpoint option has been provided")
	}
	if ha.options.backendRedisStorageEndpoint == "" {
		return fmt.Errorf("no Backend Redis storage endpoint has been provided")
	}
	if ha.options.systemDatabaseURL == "" {
		return fmt.Errorf("no System database URL has been provided")
	}
	if ha.options.systemRedisURL == "" {
		return fmt.Errorf("no System redis URL has been provided")
	}
	if ha.options.systemMessageBusRedisURL == "" {
		return fmt.Errorf("no System Message Bus redis URL has been provided")
	}

	return nil
}

func (ha *HighAvailabilityOptionsBuilder) setNonRequiredOptions() {

}
