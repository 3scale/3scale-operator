package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
)

func ApicastImageURL() string {
	return helper.GetEnvVar("APICAST_IMAGE", component.ApicastImageURL())
}

func ApisonatorImageURL() string {
	return helper.GetEnvVar("APISONATOR_IMAGE", component.ApisonatorImageURL())
}

func PortaImageURL() string {
	return helper.GetEnvVar("PORTA_IMAGE", component.PortaImageURL())
}

func ZyncImageURL() string {
	return helper.GetEnvVar("ZYNC_IMAGE", component.ZyncImageURL())
}

func PortaMemcachedImageURL() string {
	return helper.GetEnvVar("PORTA_MEMCACHED_IMAGE", component.PortaMemcachedImageURL())
}
