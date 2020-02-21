package operator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func backendRedisTestData() map[string]string {
	return map[string]string{
		component.BackendSecretBackendRedisStorageURLFieldName: "redis://storage.redis.example.com",
		component.BackendSecretBackendRedisQueuesURLFieldName:  "redis://queue.redis.example.com",
	}
}

func systemRedisTestData() map[string]string {
	return map[string]string{
		component.SystemSecretSystemRedisURLFieldName:                "redis://system.redis.example.com",
		component.SystemSecretSystemRedisMessageBusRedisURLFieldName: "redis://messagebus.redis.example.com",
	}
}

func systemDatabaseTestData() map[string]string {
	return map[string]string{
		component.SystemSecretSystemDatabaseURLFieldName: "mysql://mysql.example.com",
	}
}

func TestGetHighAvailabilityOptions(t *testing.T) {
	namespace := "someNS"

	backendRedisSecret := helper.GetTestSecret(namespace, component.BackendSecretBackendRedisSecretName, backendRedisTestData())
	systemRedisSecret := helper.GetTestSecret(namespace, component.SystemSecretSystemRedisSecretName, systemRedisTestData())
	systemDatabaseSecret := helper.GetTestSecret(namespace, component.SystemSecretSystemDatabaseSecretName, systemDatabaseTestData())

	objs := []runtime.Object{backendRedisSecret, systemRedisSecret, systemDatabaseSecret}
	cl := fake.NewFakeClient(objs...)
	optsProvider := NewHighAvailabilityOptionsProvider(namespace, cl)
	_, err := optsProvider.GetHighAvailabilityOptions()
	if err != nil {
		t.Fatal(err)
	}
	// created "opts" cannot be tested  here, it only has set methods
	// and cannot assert on setted values from a different package
	// TODO: refactor options provider structure
	// then validate setted resources
}

func TestGetHighAvailabilityOptionsInvalid(t *testing.T) {
	namespace := "someNS"

	cases := []struct {
		testName                 string
		backendRedisSecretData   map[string]string
		systemRedisSecretData    map[string]string
		systemDatabaseSecretData map[string]string
		errSubstr                string
	}{
		{
			"NoBackendRedisSecretFound", nil, systemRedisTestData(), systemDatabaseTestData(),
			fmt.Sprintf("\"%s\" not found", component.BackendSecretBackendRedisSecretName),
		},
		{
			"NoSystemRedisSecretFound", backendRedisTestData(), nil, systemDatabaseTestData(),
			fmt.Sprintf("\"%s\" not found", component.SystemSecretSystemRedisSecretName),
		},
		{
			"NoSystemDatabaseSecretFound", backendRedisTestData(), systemRedisTestData(), nil,
			fmt.Sprintf("\"%s\" not found", component.SystemSecretSystemDatabaseSecretName),
		},
		{
			"BackendRedisStorageURLMissing",
			func() map[string]string {
				data := backendRedisTestData()
				delete(data, component.BackendSecretBackendRedisStorageURLFieldName)
				return data
			}(),
			systemRedisTestData(), systemDatabaseTestData(), component.BackendSecretBackendRedisStorageURLFieldName,
		},
		{
			"BackendRedisQueueURLMissing",
			func() map[string]string {
				data := backendRedisTestData()
				delete(data, component.BackendSecretBackendRedisQueuesURLFieldName)
				return data
			}(),
			systemRedisTestData(), systemDatabaseTestData(), component.BackendSecretBackendRedisQueuesURLFieldName,
		},
		{
			"SystemRedisURLMissing",
			backendRedisTestData(),
			func() map[string]string {
				data := systemRedisTestData()
				delete(data, component.SystemSecretSystemRedisURLFieldName)
				return data
			}(),
			systemDatabaseTestData(), component.SystemSecretSystemRedisURLFieldName,
		},
		{
			"SystemRedisMessagebusURLMissing",
			backendRedisTestData(),
			func() map[string]string {
				data := systemRedisTestData()
				delete(data, component.SystemSecretSystemRedisMessageBusRedisURLFieldName)
				return data
			}(),
			systemDatabaseTestData(), component.SystemSecretSystemRedisMessageBusRedisURLFieldName,
		},
		{
			"SystemDatabaseURLMissing",
			backendRedisTestData(),
			systemRedisTestData(),
			func() map[string]string {
				data := systemDatabaseTestData()
				delete(data, component.SystemSecretSystemDatabaseURLFieldName)
				return data
			}(),
			component.SystemSecretSystemDatabaseURLFieldName,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
			if tc.backendRedisSecretData != nil {
				objs = append(objs, helper.GetTestSecret(namespace, component.BackendSecretBackendRedisSecretName, tc.backendRedisSecretData))
			}
			if tc.systemRedisSecretData != nil {
				objs = append(objs, helper.GetTestSecret(namespace, component.SystemSecretSystemRedisSecretName, tc.systemRedisSecretData))
			}
			if tc.systemDatabaseSecretData != nil {
				objs = append(objs, helper.GetTestSecret(namespace, component.SystemSecretSystemDatabaseSecretName, tc.systemDatabaseSecretData))
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
