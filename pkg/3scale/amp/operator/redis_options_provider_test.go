package operator

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"

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
	labels := map[string]string{
		"app":                               appLabel,
		"threescale_component":              "system",
		"threescale_component_element":      "redis",
		reconcilers.DeploymentLabelSelector: "system-redis",
	}
	addExpectedMeteringLabels(labels, "system-redis", helper.ApplicationType)

	return labels
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
	labels := map[string]string{
		"app":                               appLabel,
		"threescale_component":              "backend",
		"threescale_component_element":      "redis",
		reconcilers.DeploymentLabelSelector: "backend-redis",
	}
	addExpectedMeteringLabels(labels, "backend-redis", helper.ApplicationType)

	return labels
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

func testBackendRedisSecret() *v1.Secret {
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

func testSystemRedisSecret() *v1.Secret {
	data := map[string]string{
		component.SystemSecretSystemRedisURLFieldName:  "redis://system1:6379",
		component.SystemSecretSystemRedisSentinelHosts: "someHosts1",
		component.SystemSecretSystemRedisSentinelRole:  "someRole1",
	}
	return GetTestSecret(namespace, component.SystemSecretSystemRedisSecretName, data)
}

func testBackendRedisPodTemplateAnnotations() map[string]string {
	return map[string]string{"redisConfigMapResourceVersion": "999"}
}

func testSystemRedisPodTemplateAnnotations() map[string]string {
	return map[string]string{"redisConfigMapResourceVersion": "999"}
}

func defaultRedisOptions() *component.RedisOptions {
	return &component.RedisOptions{
		AmpRelease:      version.ThreescaleVersionMajorMinor(),
		BackendImageTag: version.ThreescaleVersionMajorMinor(),
		SystemImageTag:  version.ThreescaleVersionMajorMinor(),
		BackendImage:    component.BackendRedisImageURL(),
		SystemImage:     component.SystemRedisImageURL(),
		BackendRedisContainerResourceRequirements: component.DefaultBackendRedisContainerResourceRequirements(),
		SystemRedisContainerResourceRequirements:  component.DefaultSystemRedisContainerResourceRequirements(),
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
		SystemRedisSentinelsHosts:                 component.DefaultSystemRedisSentinelHosts(),
		SystemRedisSentinelsRole:                  component.DefaultSystemRedisSentinelRole(),
		SystemRedisPodTemplateAnnotations:         testSystemRedisPodTemplateAnnotations(),
		BackendRedisPodTemplateAnnotations:        testBackendRedisPodTemplateAnnotations(),
		SystemRedisSSL:                            component.DefaultSystemRedisSSL(),
		BackendConfigSSL:                          component.DefaultBackendConfigSSL(),
		BackendConfigQueuesSSL:                    component.DefaultBackendConfigQueuesSSL(),
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
		backendRedisSecret     *v1.Secret
		systemRedisSecret      *v1.Secret
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.RedisOptions
	}{
		{"Default", nil, nil, basicApimanager, defaultRedisOptions},
		{"WithoutResourceRequirements", nil, nil,
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
		{"BackendRedisImageSet", nil, nil,
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
				return opts
			},
		},
		{"SystemRedisImageSet", nil, nil,
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
				return opts
			},
		},
		{"SystemRedisOnlyPVCSpecSet", nil, nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{
					RedisPersistentVolumeClaimSpec: &appsv1alpha1.SystemRedisPersistentVolumeClaimSpec{},
				}
				return apimanager
			}, defaultRedisOptions,
		},
		{"BackendRedisOnlyPVCSpecSet", nil, nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend = &appsv1alpha1.BackendSpec{
					RedisPersistentVolumeClaimSpec: &appsv1alpha1.BackendRedisPersistentVolumeClaimSpec{},
				}
				return apimanager
			}, defaultRedisOptions,
		},
		{"BackendRedisStoragePVCStorageClassSet", nil, nil,
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
		{"SystemRedisStoragePVCStorageClassSet", nil, nil,
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
		{"WithAffinity", nil, nil,
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
		{"WithTolerations", nil, nil,
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
		{"WithBackendRedisCustomResourceRequirements", nil, nil,
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
		{"WithBackendRedisCustomResourceRequirementsAndGlobalResourceRequirementsDisabled", nil, nil,
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
		{"WithSystemRedisCustomResourceRequirements", nil, nil,
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
		{"WithSystemRedisCustomResourceRequirementsAndGlobalResourceRequirementsDisabled", nil, nil,
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
		{"WithBackendRedisSecret", testBackendRedisSecret(), nil, basicApimanager,
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.BackendStorageURL = "storageURLValue"
				opts.BackendQueuesURL = "queueURLValue"
				opts.BackendRedisStorageSentinelHosts = "storageSentinelHostsValue"
				opts.BackendRedisStorageSentinelRole = "storageSentinelRoleValue"
				opts.BackendRedisQueuesSentinelHosts = "queueSentinelHostsValue"
				opts.BackendRedisQueuesSentinelRole = "queueSentinelRoleValue"
				return opts
			},
		},
		{"WithSystemRedisSecret", nil, testSystemRedisSecret(), basicApimanager,
			func() *component.RedisOptions {
				opts := defaultRedisOptions()
				opts.SystemRedisURL = "redis://system1:6379"
				opts.SystemRedisSentinelsHosts = "someHosts1"
				opts.SystemRedisSentinelsRole = "someRole1"
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			configMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      "redis-config",
				},
				Data: map[string]string{},
			}
			objs := []runtime.Object{}
			if tc.backendRedisSecret != nil {
				objs = append(objs, tc.backendRedisSecret)
			}
			if tc.systemRedisSecret != nil {
				objs = append(objs, tc.systemRedisSecret)
			}
			objs = append(objs, configMap)
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
