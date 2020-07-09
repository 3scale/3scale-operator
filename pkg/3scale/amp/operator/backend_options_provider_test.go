package operator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	listenerReplicaCount int64 = 3
	workerReplicaCount   int64 = 4
	cronReplicaCount     int64 = 5
)

func testBackendCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "backend",
	}
}

func testBackendCommonListenerLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "listener",
	}
}

func testBackendCommonWorkerLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "worker",
	}
}

func testBackendCommonCronLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "cron",
	}
}

func testBackendListenerPodLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "listener",
		"com.redhat.component-name":    "backend-listener",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "nightly",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "backend-listener",
	}
}

func testBackendWorkerPodLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "worker",
		"com.redhat.component-name":    "backend-worker",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "nightly",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "backend-worker",
	}
}

func testBackendCronPodLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "cron",
		"com.redhat.component-name":    "backend-cron",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "nightly",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "backend-cron",
	}
}

func testBackendListenerAffinity() *v1.Affinity {
	return getTestAffinity("backend-listener")
}

func testBackendWorkerAffinity() *v1.Affinity {
	return getTestAffinity("backend-worker")
}

func testBackendCronAffinity() *v1.Affinity {
	return getTestAffinity("backend-cron")
}

func testBackendListenerTolerations() []v1.Toleration {
	return getTestTolerations("backend-listener")
}

func testBackendWorkerTolerations() []v1.Toleration {
	return getTestTolerations("backend-worker")
}

func testBackendCronTolerations() []v1.Toleration {
	return getTestTolerations("backend-cron")
}

func getInternalSecret() *v1.Secret {
	data := map[string]string{
		component.BackendSecretInternalApiUsernameFieldName: "someUserName",
		component.BackendSecretInternalApiPasswordFieldName: "somePasswd",
	}
	return GetTestSecret(namespace, component.BackendSecretInternalApiSecretName, data)
}

func getListenerSecret() *v1.Secret {
	data := map[string]string{
		component.BackendSecretBackendListenerServiceEndpointFieldName: "serviceValue",
		component.BackendSecretBackendListenerRouteEndpointFieldName:   "routeValue",
	}
	return GetTestSecret(namespace, component.BackendSecretBackendListenerSecretName, data)
}

func getRedisSecret() *v1.Secret {
	data := map[string]string{
		component.BackendSecretBackendRedisStorageURLFieldName:           "storageURLValue",
		component.BackendSecretBackendRedisQueuesURLFieldName:            "queueURLValue",
		component.BackendSecretBackendRedisStorageSentinelHostsFieldName: "storageSentinelHostsValue",
		component.BackendSecretBackendRedisStorageSentinelRoleFieldName:  "storageSentinelRoleValue",
		component.BackendSecretBackendRedisQueuesSentinelHostsFieldName:  "queueSentinelHostsValue",
		component.BackendSecretBackendRedisQueuesSentinelRoleFieldName:   "queueSentinelRoleValue",
	}
	return GetTestSecret(namespace, component.BackendSecretBackendRedisSecretName, data)
}
func basicApimanagerTestBackendOptions() *appsv1alpha1.APIManager {
	tmpListenerReplicaCount := listenerReplicaCount
	tmpWorkerReplicaCount := workerReplicaCount
	tmpCronReplicaCount := cronReplicaCount

	apimanager := basicApimanager()
	apimanager.Spec.Backend = &appsv1alpha1.BackendSpec{
		ListenerSpec: &appsv1alpha1.BackendListenerSpec{Replicas: &tmpListenerReplicaCount},
		WorkerSpec:   &appsv1alpha1.BackendWorkerSpec{Replicas: &tmpWorkerReplicaCount},
		CronSpec:     &appsv1alpha1.BackendCronSpec{Replicas: &tmpCronReplicaCount},
	}
	return apimanager
}

func defaultBackendOptions(opts *component.BackendOptions) *component.BackendOptions {
	return &component.BackendOptions{
		ServiceEndpoint:              component.DefaultBackendServiceEndpoint(),
		RouteEndpoint:                fmt.Sprintf("https://backend-%s.%s", tenantName, wildcardDomain),
		StorageURL:                   component.DefaultBackendRedisStorageURL(),
		QueuesURL:                    component.DefaultBackendRedisQueuesURL(),
		StorageSentinelHosts:         component.DefaultBackendStorageSentinelHosts(),
		StorageSentinelRole:          component.DefaultBackendStorageSentinelRole(),
		QueuesSentinelHosts:          component.DefaultBackendQueuesSentinelHosts(),
		QueuesSentinelRole:           component.DefaultBackendQueuesSentinelRole(),
		ListenerResourceRequirements: component.DefaultBackendListenerResourceRequirements(),
		WorkerResourceRequirements:   component.DefaultBackendWorkerResourceRequirements(),
		CronResourceRequirements:     component.DefaultCronResourceRequirements(),
		ListenerReplicas:             int32(listenerReplicaCount),
		WorkerReplicas:               int32(workerReplicaCount),
		CronReplicas:                 int32(cronReplicaCount),
		SystemBackendUsername:        component.DefaultSystemBackendUsername(),
		SystemBackendPassword:        opts.SystemBackendPassword,
		TenantName:                   tenantName,
		WildcardDomain:               wildcardDomain,
		ImageTag:                     product.ThreescaleRelease,
		CommonLabels:                 testBackendCommonLabels(),
		CommonListenerLabels:         testBackendCommonListenerLabels(),
		CommonWorkerLabels:           testBackendCommonWorkerLabels(),
		CommonCronLabels:             testBackendCommonCronLabels(),
		ListenerPodTemplateLabels:    testBackendListenerPodLabels(),
		WorkerPodTemplateLabels:      testBackendWorkerPodLabels(),
		CronPodTemplateLabels:        testBackendCronPodLabels(),
		WorkerMetrics:                true,
		ListenerMetrics:              true,
	}
}

func TestGetBackendOptionsProvider(t *testing.T) {
	falseValue := false

	cases := []struct {
		testName               string
		internalSecret         *v1.Secret
		listenerSecret         *v1.Secret
		redisSecret            *v1.Secret
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func(*component.BackendOptions) *component.BackendOptions
	}{
		{"Default", nil, nil, nil, basicApimanagerTestBackendOptions,
			func(opts *component.BackendOptions) *component.BackendOptions {
				return defaultBackendOptions(opts)
			},
		},
		{"WithoutResourceRequirements", nil, nil, nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestBackendOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			},
			func(in *component.BackendOptions) *component.BackendOptions {
				opts := defaultBackendOptions(in)
				opts.ListenerResourceRequirements = v1.ResourceRequirements{}
				opts.WorkerResourceRequirements = v1.ResourceRequirements{}
				opts.CronResourceRequirements = v1.ResourceRequirements{}
				return opts
			},
		},
		{"InternalSecret", getInternalSecret(), nil, nil, basicApimanagerTestBackendOptions,
			func(in *component.BackendOptions) *component.BackendOptions {
				opts := defaultBackendOptions(in)
				opts.SystemBackendUsername = "someUserName"
				opts.SystemBackendPassword = "somePasswd"
				return opts
			},
		},
		{"ListenerSecret", nil, getListenerSecret(), nil, basicApimanagerTestBackendOptions,
			func(in *component.BackendOptions) *component.BackendOptions {
				opts := defaultBackendOptions(in)
				opts.ServiceEndpoint = "serviceValue"
				opts.RouteEndpoint = "routeValue"
				return opts
			},
		},
		{"RedisSecret", nil, nil, getRedisSecret(), basicApimanagerTestBackendOptions,
			func(in *component.BackendOptions) *component.BackendOptions {
				opts := defaultBackendOptions(in)
				opts.StorageURL = "storageURLValue"
				opts.QueuesURL = "queueURLValue"
				opts.StorageSentinelHosts = "storageSentinelHostsValue"
				opts.StorageSentinelRole = "storageSentinelRoleValue"
				opts.QueuesSentinelHosts = "queueSentinelHostsValue"
				opts.QueuesSentinelRole = "queueSentinelRoleValue"
				return opts
			},
		},
		{"WithAffinity", nil, nil, nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestBackendOptions()
				apimanager.Spec.Backend.ListenerSpec.Affinity = testBackendListenerAffinity()
				apimanager.Spec.Backend.WorkerSpec.Affinity = testBackendWorkerAffinity()
				apimanager.Spec.Backend.CronSpec.Affinity = testBackendCronAffinity()
				return apimanager
			},
			func(in *component.BackendOptions) *component.BackendOptions {
				opts := defaultBackendOptions(in)
				opts.ListenerAffinity = testBackendListenerAffinity()
				opts.WorkerAffinity = testBackendWorkerAffinity()
				opts.CronAffinity = testBackendCronAffinity()
				return opts
			},
		},
		{"WithTolerations", nil, nil, nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestBackendOptions()
				apimanager.Spec.Backend.ListenerSpec.Tolerations = testBackendListenerTolerations()
				apimanager.Spec.Backend.WorkerSpec.Tolerations = testBackendWorkerTolerations()
				apimanager.Spec.Backend.CronSpec.Tolerations = testBackendCronTolerations()
				return apimanager
			},
			func(in *component.BackendOptions) *component.BackendOptions {
				opts := defaultBackendOptions(in)

				opts.ListenerTolerations = testBackendListenerTolerations()
				opts.WorkerTolerations = testBackendWorkerTolerations()
				opts.CronTolerations = testBackendCronTolerations()
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
			if tc.internalSecret != nil {
				objs = append(objs, tc.internalSecret)
			}
			if tc.listenerSecret != nil {
				objs = append(objs, tc.listenerSecret)
			}
			if tc.redisSecret != nil {
				objs = append(objs, tc.redisSecret)
			}

			cl := fake.NewFakeClient(objs...)
			optsProvider := NewOperatorBackendOptionsProvider(tc.apimanagerFactory(), namespace, cl)
			opts, err := optsProvider.GetBackendOptions()
			if err != nil {
				t.Error(err)
			}
			expectedOptions := tc.expectedOptionsFactory(opts)
			if !reflect.DeepEqual(expectedOptions, opts) {
				subT.Errorf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}
