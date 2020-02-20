package operator

import (
	"os"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func TestImageURLFromEnv(t *testing.T) {
	newImageURL := "http://quay.io/myorg/someimage"

	cases := []struct {
		name       string
		envVarName string
		imageURL   func() string
	}{
		{"ApicastURL", "APICAST_IMAGE", func() string { return ApicastImageURL() }},
		{"BackendURL", "BACKEND_IMAGE", func() string { return BackendImageURL() }},
		{"SystemURL", "SYSTEM_IMAGE", func() string { return SystemImageURL() }},
		{"ZyncURL", "ZYNC_IMAGE", func() string { return ZyncImageURL() }},
		{"SystemMemcachedURL", "SYSTEM_MEMCACHED_IMAGE", func() string { return SystemMemcachedImageURL() }},
		{"SystemRedisImageURL", "SYSTEM_REDIS_IMAGE", func() string { return SystemRedisImageURL() }},
		{"SystemMySQLImageURL", "SYSTEM_MYSQL_IMAGE", func() string { return SystemMySQLImageURL() }},
		{"SystemPostgreSQLImageURL", "SYSTEM_POSTGRESQL_IMAGE", func() string { return SystemPostgreSQLImageURL() }},
		{"ZyncPostgreSQLImageURL", "ZYNC_POSTGRESQL_IMAGE", func() string { return ZyncPostgreSQLImageURL() }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			os.Setenv(tc.envVarName, newImageURL)
			defer func() {
				os.Unsetenv(tc.envVarName)
			}()
			imageURL := tc.imageURL()
			if imageURL != newImageURL {
				t.Fatalf("image url does not match. Expected: %s, got: %s", newImageURL, imageURL)
			}
		})
	}
}

func TestImageURLDefault(t *testing.T) {
	cases := []struct {
		name     string
		imageURL func() string
		expected func() string
	}{
		{"ApicastURL", func() string { return ApicastImageURL() }, func() string { return component.ApicastImageURL() }},
		{"BackendURL", func() string { return BackendImageURL() }, func() string { return component.BackendImageURL() }},
		{"SystemURL", func() string { return SystemImageURL() }, func() string { return component.SystemImageURL() }},
		{"ZyncURL", func() string { return ZyncImageURL() }, func() string { return component.ZyncImageURL() }},
		{"SystemMemcachedImageURL", func() string { return SystemMemcachedImageURL() }, func() string { return component.SystemMemcachedImageURL() }},
		{"SystemRedisImageURL", func() string { return SystemRedisImageURL() }, func() string { return component.SystemRedisImageURL() }},
		{"SystemMySQLImageURL", func() string { return SystemMySQLImageURL() }, func() string { return component.SystemMySQLImageURL() }},
		{"SystemPostgreSQLImageURL", func() string { return SystemPostgreSQLImageURL() }, func() string { return component.SystemPostgreSQLImageURL() }},
		{"ZyncPostgreSQLImageURL", func() string { return ZyncPostgreSQLImageURL() }, func() string { return component.ZyncPostgreSQLImageURL() }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			imageURL := tc.imageURL()
			if imageURL != tc.expected() {
				t.Fatalf("image url does not match. Expected: %s, got: %s", tc.expected(), imageURL)
			}
		})
	}
}
