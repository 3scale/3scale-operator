package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorHighAvailabilityOptionsProvider) GetHighAvailabilityOptions() (*component.HighAvailabilityOptions, error) {
	hob := component.HighAvailabilityOptionsBuilder{}
	hob.ApicastProductionRedisURL(*o.APIManagerSpec.ApicastProductionRedisURL)
	hob.ApicastStagingRedisURL(*o.APIManagerSpec.ApicastStagingRedisURL)
	hob.BackendRedisQueuesEndpoint(*o.APIManagerSpec.BackendRedisQueuesEndpoint)
	hob.BackendRedisStorageEndpoint(*o.APIManagerSpec.BackendRedisStorageEndpoint)
	hob.SystemDatabaseURL(*o.APIManagerSpec.SystemDatabaseURL)
	hob.SystemRedisURL(*o.APIManagerSpec.SystemRedisURL)
	res, err := hob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create HighAvailability Options - %s", err)
	}
	return res, nil
}
