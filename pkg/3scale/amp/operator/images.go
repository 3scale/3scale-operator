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

func SystemSearchdImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_SYSTEM_SEARCHD", component.SystemSearchdImageURL())
}

func ZyncPostgreSQLImageURL() string {
	return helper.GetEnvVar("RELATED_IMAGE_ZYNC_POSTGRESQL", component.ZyncPostgreSQLImageURL())
}
