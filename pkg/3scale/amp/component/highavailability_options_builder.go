package component

import "fmt"

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
