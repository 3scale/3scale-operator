package operator

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
)

func ApicastImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_APICAST", component.ApicastImageURL())
}

func BackendImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_BACKEND", component.BackendImageURL())
}

func SystemImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_SYSTEM", component.SystemImageURL())
}

func ZyncImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_ZYNC", component.ZyncImageURL())
}

func SystemMemcachedImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_SYSTEM_MEMCACHED", component.SystemMemcachedImageURL())
}

func BackendRedisImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_BACKEND_REDIS", component.BackendRedisImageURL())
}

func SystemRedisImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_SYSTEM_REDIS", component.SystemRedisImageURL())
}

func SystemMySQLImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_SYSTEM_MYSQL", component.SystemMySQLImageURL())
}

func SystemPostgreSQLImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_SYSTEM_POSTGRESQL", component.SystemPostgreSQLImageURL())
}

func ZyncPostgreSQLImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_ZYNC_POSTGRESQL", component.ZyncPostgreSQLImageURL())
}
