package operator

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	backendStorageURL   = "redis://storage.redis.example.com"
	backendQueueURL     = "redis://queue.redis.example.com"
	systemRedisURL      = "redis://system.redis.example.com"
	systemMessageBusURL = "redis://messagebus.redis.example.com"
	systemDatabaseURL   = "mysql://mysql.example.com"
)

func getSystemRedisSecretMissingRedisURL() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemRedisMessageBusRedisURLFieldName: systemMessageBusURL,
	}
	return GetTestSecret(namespace, component.SystemSecretSystemRedisSecretName, data)
}

func getSystemRedisSecretMissingMessageBusRedisURL() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemRedisURLFieldName: systemRedisURL,
	}
	return GetTestSecret(namespace, component.SystemSecretSystemRedisSecretName, data)
}

func getSystemRedisSecretForHighAvailabilityTest() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemRedisURLFieldName:                systemRedisURL,
		component.SystemSecretSystemRedisMessageBusRedisURLFieldName: systemMessageBusURL,
	}
	return GetTestSecret(namespace, component.SystemSecretSystemRedisSecretName, data)
}

func getSystemDatabaseSecretMissingDatabaseURL() *v1.Secret {
	data := map[string]string{}
	return GetTestSecret(namespace, component.SystemSecretSystemDatabaseSecretName, data)
}

func getSystemDatabaseSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemDatabaseURLFieldName: systemDatabaseURL,
	}
	return GetTestSecret(namespace, component.SystemSecretSystemDatabaseSecretName, data)
}

func getBackendRedisSecretMissingStorageURL() *v1.Secret {
	data := map[string]string{
		component.BackendSecretBackendRedisQueuesURLFieldName: backendQueueURL,
	}
	return GetTestSecret(namespace, component.BackendSecretBackendRedisSecretName, data)
}

func getBackendRedisSecretMissingQueueURL() *v1.Secret {
	data := map[string]string{
		component.BackendSecretBackendRedisStorageURLFieldName: backendStorageURL,
	}
	return GetTestSecret(namespace, component.BackendSecretBackendRedisSecretName, data)
}
func getBackendRedisSecret() *v1.Secret {
	data := map[string]string{
		component.BackendSecretBackendRedisStorageURLFieldName: backendStorageURL,
		component.BackendSecretBackendRedisQueuesURLFieldName:  backendQueueURL,
	}
	return GetTestSecret(namespace, component.BackendSecretBackendRedisSecretName, data)
}

func TestGetHighAvailabilityOptionsProvider(t *testing.T) {
	objs := []runtime.Object{getSystemRedisSecretForHighAvailabilityTest(), getSystemDatabaseSecret(), getBackendRedisSecret()}
	cl := fake.NewFakeClient(objs...)
	optsProvider := NewHighAvailabilityOptionsProvider(namespace, cl)
	opts, err := optsProvider.GetHighAvailabilityOptions()
	if err != nil {
		t.Fatal(err)
	}
	expectedOptions := &component.HighAvailabilityOptions{
		AppLabel:                            "-",
		BackendRedisQueuesEndpoint:          backendQueueURL,
		BackendRedisQueuesSentinelHosts:     "-",
		BackendRedisQueuesSentinelRole:      "-",
		BackendRedisStorageEndpoint:         backendStorageURL,
		BackendRedisStorageSentinelHosts:    "-",
		BackendRedisStorageSentinelRole:     "-",
		SystemRedisURL:                      systemRedisURL,
		SystemRedisSentinelsHosts:           "-",
		SystemRedisSentinelsRole:            "-",
		SystemMessageBusRedisURL:            systemMessageBusURL,
		SystemMessageBusRedisSentinelsHosts: "-",
		SystemMessageBusRedisSentinelsRole:  "-",
		SystemDatabaseURL:                   systemDatabaseURL,
	}

	if !reflect.DeepEqual(expectedOptions, opts) {
		t.Errorf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts))
	}
}

func TestGetHighAvailabilityOptionsInvalid(t *testing.T) {
	cases := []struct {
		testName             string
		backendRedisSecret   *v1.Secret
		systemRedisSecret    *v1.Secret
		systemDatabaseSecret *v1.Secret
		errSubstr            string
	}{
		{
			"NoBackendRedisSecretFound",
			nil,
			getSystemRedisSecretForHighAvailabilityTest(),
			getSystemDatabaseSecret(),
			fmt.Sprintf("\"%s\" not found", component.BackendSecretBackendRedisSecretName),
		},
		{
			"NoSystemRedisSecretFound",
			getBackendRedisSecret(),
			nil,
			getSystemDatabaseSecret(),
			fmt.Sprintf("\"%s\" not found", component.SystemSecretSystemRedisSecretName),
		},
		{
			"NoSystemDatabaseSecretFound",
			getBackendRedisSecret(),
			getSystemRedisSecretForHighAvailabilityTest(),
			nil,
			fmt.Sprintf("\"%s\" not found", component.SystemSecretSystemDatabaseSecretName),
		},
		{
			"BackendRedisStorageURLMissing",
			getBackendRedisSecretMissingStorageURL(),
			getSystemRedisSecretForHighAvailabilityTest(),
			getSystemDatabaseSecret(),
			component.BackendSecretBackendRedisStorageURLFieldName,
		},
		{
			"BackendRedisQueueURLMissing",
			getBackendRedisSecretMissingQueueURL(),
			getSystemRedisSecretForHighAvailabilityTest(),
			getSystemDatabaseSecret(),
			component.BackendSecretBackendRedisQueuesURLFieldName,
		},
		{
			"SystemRedisURLMissing",
			getBackendRedisSecret(),
			getSystemRedisSecretMissingRedisURL(),
			getSystemDatabaseSecret(),
			component.SystemSecretSystemRedisURLFieldName,
		},
		{
			"SystemRedisMessagebusURLMissing",
			getBackendRedisSecret(),
			getSystemRedisSecretMissingMessageBusRedisURL(),
			getSystemDatabaseSecret(),
			component.SystemSecretSystemRedisMessageBusRedisURLFieldName,
		},
		{
			"SystemDatabaseURLMissing",
			getBackendRedisSecret(),
			getSystemRedisSecretForHighAvailabilityTest(),
			getSystemDatabaseSecretMissingDatabaseURL(),
			component.SystemSecretSystemDatabaseURLFieldName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
			if tc.backendRedisSecret != nil {
				objs = append(objs, tc.backendRedisSecret)
			}
			if tc.systemRedisSecret != nil {
				objs = append(objs, tc.systemRedisSecret)
			}
			if tc.systemDatabaseSecret != nil {
				objs = append(objs, tc.systemDatabaseSecret)
			}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewHighAvailabilityOptionsProvider(namespace, cl)
			_, err := optsProvider.GetHighAvailabilityOptions()
			if err == nil {
				subT.Fatal("expected to fail")
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				subT.Fatalf("expected error regexp: %s, got: (%v)", tc.errSubstr, err)
			}
		})
	}
}
