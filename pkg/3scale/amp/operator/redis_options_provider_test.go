package operator

import (
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

func testRedisSystemCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "system",
	}
}

func testRedisSystemRedisLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "redis",
	}
}

func testRedisSystemRedisPodTemplateLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "redis",
		"com.redhat.component-name":    "system-redis",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "32",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "system-redis",
	}
}

func testRedisBackendCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "backend",
	}
}

func testRedisBackendRedisLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "redis",
	}
}

func testRedisBackendRedisPodTemplateLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "backend",
		"threescale_component_element": "redis",
		"com.redhat.component-name":    "backend-redis",
		"com.redhat.component-type":    "application",
		"com.redhat.component-version": "32",
		"com.redhat.product-name":      "3scale",
		"com.redhat.product-version":   "master",
		"deploymentConfig":             "backend-redis",
	}
}

func testBackendRedisAffinity() *v1.Affinity {
	return getTestAffinity("backend-redis")
}

func testSystemRedisAffinity() *v1.Affinity {
	return getTestAffinity("system-redis")
}

func testBackendRedisTolerations() []v1.Toleration {
	return getTestTolerations("backend-redis")
}

func testSystemRedisTolerations() []v1.Toleration {
	return getTestTolerations("system-redis")
}

func testBackendRedisCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("111m"),
			v1.ResourceMemory: resource.MustParse("222Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("333m"),
			v1.ResourceMemory: resource.MustParse("444Mi"),
		},
	}
}

func testSystemRedisCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("222m"),
			v1.ResourceMemory: resource.MustParse("333Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("444m"),
			v1.ResourceMemory: resource.MustParse("555Mi"),
		},
	}
}

func defaultRedisOptions() *component.RedisOptions {
	tmpInsecure := insecureImportPolicy
	return &component.RedisOptions{
		AmpRelease:      product.ThreescaleRelease,
		BackendImageTag: product.ThreescaleRelease,
		SystemImageTag:  product.ThreescaleRelease,
		BackendImage:    component.BackendRedisImageURL(),
		SystemImage:     component.SystemRedisImageURL(),
		BackendRedisContainerResourceRequirements: component.DefaultBackendRedisContainerResourceRequirements(),
		SystemRedisContainerResourceRequirements:  component.DefaultSystemRedisContainerResourceRequirements(),
		InsecureImportPolicy:                      &tmpInsecure,
		SystemCommonLabels:                        testRedisSystemCommonLabels(),
		SystemRedisLabels:                         testRedisSystemRedisLabels(),
		SystemRedisPodTemplateLabels:              testRedisSystemRedisPodTemplateLabels(),
		BackendCommonLabels:                       testRedisBackendCommonLabels(),
		BackendRedisLabels:                        testRedisBackendRedisLabels(),
		BackendRedisPodTemplateLabels:             testRedisBackendRedisPodTemplateLabels(),
		BackendStorageURL:                         component.DefaultBackendRedisStorageURL(),
		BackendQueuesURL:                          component.DefaultBackendRedisQueuesURL(),
		BackendRedisStorageSentinelHosts:          component.DefaultBackendStorageSentinelHosts(),
		BackendRedisStorageSentinelRole:           component.DefaultBackendStorageSentinelRole(),
		BackendRedisQueuesSentinelHosts:           component.DefaultBackendQueuesSentinelHosts(),
		BackendRedisQueuesSentinelRole:            component.DefaultBackendQueuesSentinelRole(),
		SystemRedisURL:                            component.DefaultSystemRedisURL(),
		SystemRedisMessageBusURL:                  component.DefaultSystemRedisMessageBusURL(),
		SystemRedisSentinelsHosts:                 component.DefaultSystemRedisSentinelHosts(),
		SystemRedisSentinelsRole:                  component.DefaultSystemRedisSentinelRole(),
		SystemMessageBusRedisSentinelsHosts:       component.DefaultSystemMessageBusRedisSentinelHosts(),
		SystemMessageBusRedisSentinelsRole:        component.DefaultSystemMessageBusRedisSentinelRole(),
		SystemMessageBusRedisNamespace:            component.DefaultSystemMessageBusRedisNamespace(),
		SystemRedisNamespace:                      component.DefaultSystemRedisNamespace(),
	}
}

func TestGetRedisOptionsProvider(t *testing.T) {
	tmpFalseValue := false
	backendRedisImageURL := "redis:backendCustomVersion"
	systemRedisImageURL := "redis:systemCustomVersion"
	backendRedisCustomStorageClass := "backendrediscustomstorageclass"
	systemRedisCustomStorageClass := "systemrediscustomstorageclass"

	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.RedisOptions
	}{
		{"Default", basicApimanager, defaultRedisOptions},
		{"WithoutResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &tmpFalseValue
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendRedisContainerResourceRequirements = &v1.ResourceRequirements{}
				opts.SystemRedisContainerResourceRequirements = &v1.ResourceRequirements{}
				return opts
			},
		},
		{"BackendRedisImageSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend = &appsv1alpha1.BackendSpec{
					RedisImage: &backendRedisImageURL,
				}
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendImage = backendRedisImageURL
				opts.BackendRedisPodTemplateLabels["com.redhat.component-version"] = "backendCustomVersion"
				return opts
			},
		},
		{"SystemRedisImageSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					RedisImage: &systemRedisImageURL,
				}
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.SystemImage = systemRedisImageURL
				opts.SystemRedisPodTemplateLabels["com.redhat.component-version"] = "systemCustomVersion"
				return opts
			},
		},
		{"SystemRedisOnlyPVCSpecSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					RedisPersistentVolumeClaimSpec: &appsv1alpha1.SystemRedisPersistentVolumeClaimSpec{},
				}
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				return opts
			},
		},
		{"BackendRedisOnlyPVCSpecSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend = &appsv1alpha1.BackendSpec{
					RedisPersistentVolumeClaimSpec: &appsv1alpha1.BackendRedisPersistentVolumeClaimSpec{},
				}
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				return opts
			},
		},
		{"BackendRedisStoragePVCStorageClassSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend = &appsv1alpha1.BackendSpec{
					RedisPersistentVolumeClaimSpec: &appsv1alpha1.BackendRedisPersistentVolumeClaimSpec{
						StorageClassName: &backendRedisCustomStorageClass,
					},
				}
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendRedisPVCStorageClass = &backendRedisCustomStorageClass
				return opts
			},
		},
		{"SystemRedisStoragePVCStorageClassSet",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					RedisPersistentVolumeClaimSpec: &appsv1alpha1.SystemRedisPersistentVolumeClaimSpec{
						StorageClassName: &systemRedisCustomStorageClass,
					},
				}
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.SystemRedisPVCStorageClass = &systemRedisCustomStorageClass
				return opts
			},
		},
		{"WithAffinity",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System.RedisAffinity = testSystemRedisAffinity()
				apimanager.Spec.Backend.RedisAffinity = testBackendRedisAffinity()
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.SystemRedisAffinity = testSystemRedisAffinity()
				opts.BackendRedisAffinity = testBackendRedisAffinity()
				return opts
			},
		},
		{"WithTolerations",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System.RedisTolerations = testSystemRedisTolerations()
				apimanager.Spec.Backend.RedisTolerations = testBackendRedisTolerations()
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.SystemRedisTolerations = testSystemRedisTolerations()
				opts.BackendRedisTolerations = testBackendRedisTolerations()
				return opts
			},
		},
		{"WithBackendRedisCustomResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend.RedisResources = testBackendRedisCustomResourceRequirements()
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendRedisContainerResourceRequirements = testBackendRedisCustomResourceRequirements()
				return opts
			},
		},
		{"WithBackendRedisCustomResourceRequirementsAndGlobalResourceRequirementsDisabled",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &tmpFalseValue
				apimanager.Spec.Backend.RedisResources = testBackendRedisCustomResourceRequirements()
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.SystemRedisContainerResourceRequirements = &v1.ResourceRequirements{}
				opts.BackendRedisContainerResourceRequirements = testBackendRedisCustomResourceRequirements()
				return opts
			},
		},
		{"WithSystemRedisCustomResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend.RedisResources = testSystemRedisCustomResourceRequirements()
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendRedisContainerResourceRequirements = testSystemRedisCustomResourceRequirements()
				return opts
			},
		},
		{"WithSystemRedisCustomResourceRequirementsAndGlobalResourceRequirementsDisabled",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &tmpFalseValue
				apimanager.Spec.System.RedisResources = testSystemRedisCustomResourceRequirements()
				return apimanager
			},
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendRedisContainerResourceRequirements = &v1.ResourceRequirements{}
				opts.SystemRedisContainerResourceRequirements = testSystemRedisCustomResourceRequirements()
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewRedisOptionsProvider(tc.apimanagerFactory(), namespace, cl)
			opts, err := optsProvider.GetRedisOptions()
			if err != nil {
				subT.Error(err)
			}
			expectedOptions := tc.expectedOptionsFactory()
			if !reflect.DeepEqual(expectedOptions, opts) {
				subT.Errorf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}
