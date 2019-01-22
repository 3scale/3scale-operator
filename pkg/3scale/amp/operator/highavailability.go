package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorHighAvailabilityOptionsProvider) GetHighAvailabilityOptions() (*component.HighAvailabilityOptions, error) {
	hob := component.HighAvailabilityOptionsBuilder{}
	hob.ApicastProductionRedisURL(*o.AmpSpec.ApicastProductionRedisURL)
	hob.ApicastStagingRedisURL(*o.AmpSpec.ApicastStagingRedisURL)
	hob.BackendRedisQueuesEndpoint(*o.AmpSpec.BackendRedisQueuesEndpoint)
	hob.BackendRedisStorageEndpoint(*o.AmpSpec.BackendRedisStorageEndpoint)
	hob.SystemDatabaseURL(*o.AmpSpec.SystemDatabaseURL)
	hob.SystemRedisURL(*o.AmpSpec.SystemRedisURL)
	res, err := hob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create HighAvailability Options - %s", err)
	}
	return res, nil
}
