package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorMemcachedOptionsProvider) GetMemcachedOptions() (*component.MemcachedOptions, error) {
	optProv := component.MemcachedOptionsBuilder{}
	optProv.AppLabel(*o.AmpSpec.AppLabel)
	optProv.Image(*o.AmpSpec.MemcachedImage)
	res, err := optProv.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create Memcached Options - %s", err)
	}
	return res, nil
}
