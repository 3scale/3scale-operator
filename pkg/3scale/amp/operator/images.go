package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
)

func ApicastImageURL() string {
	return helper.GetEnvVar("APICAST_IMAGE", component.ApicastImageURL())
}

func BackendImageURL() string {
	return helper.GetEnvVar("BACKEND_IMAGE", component.BackendImageURL())
}

func SystemImageURL() string {
	return helper.GetEnvVar("SYSTEM_IMAGE", component.SystemImageURL())
}

func ZyncImageURL() string {
	return helper.GetEnvVar("ZYNC_IMAGE", component.ZyncImageURL())
}

func SystemMemcachedImageURL() string {
	return helper.GetEnvVar("SYSTEM_MEMCACHED_IMAGE", component.SystemMemcachedImageURL())
}
