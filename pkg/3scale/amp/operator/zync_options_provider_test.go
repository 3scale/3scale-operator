package operator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
)

const (
	zyncReplica           int64 = 3
	zyncQueReplica        int64 = 4
	zyncSecretKeyBasename       = "someKeyBase"
	zyncDatabasePasswd          = "somePass3424"
	zyncAuthToken               = "someToken5252"
)

var zyncExternalDatabaseTestURL = fmt.Sprintf("postgresql://exampleuser:%s@databaseurl:5432/zync_production", zyncDatabasePasswd)

func testZyncCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "zync",
	}
}

func testZyncZyncCommonLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "zync",
		"threescale_component_element": "zync",
	}
}

func testZyncQueCommonLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "zync",
		"threescale_component_element": "zync-que",
	}
}

func testZyncDatabaseCommonLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "zync",
		"threescale_component_element": "database",
	}
}

func testZyncPodTemplateLabels() map[string]string {
	labels := map[string]string{
		"app":                          appLabel,
		"threescale_component":         "zync",
		"threescale_component_element": "zync",
		"deploymentConfig":             "zync",
	}
	addExpectedMeteringLabels(labels, "zync", helper.ApplicationType)

	return labels
}

func testZyncQuePodTemplateCommonLabels() map[string]string {
	labels := map[string]string{
		"app":                          appLabel,
		"threescale_component":         "zync",
		"threescale_component_element": "zync-que",
		"deploymentConfig":             "zync-que",
	}
	addExpectedMeteringLabels(labels, "zync-que", helper.ApplicationType)

	return labels
}

func testZyncDatabasePodTemplateCommonLabels() map[string]string {
	labels := map[string]string{
		"app":                          appLabel,
		"threescale_component":         "zync",
		"threescale_component_element": "database",
		"deploymentConfig":             "zync-database",
	}
	addExpectedMeteringLabels(labels, "zync-database", helper.ApplicationType)

	return labels
}

func testZyncAffinity() *v1.Affinity {
	return getTestAffinity("zync")
}

func testZyncQueAffinity() *v1.Affinity {
	return getTestAffinity("zync-que")
}

func testZyncDatabaseAffinity() *v1.Affinity {
	return getTestAffinity("zync-database")
}

func testZyncTolerations() []v1.Toleration {
	return getTestTolerations("zync")
}

func testZyncQueTolerations() []v1.Toleration {
	return getTestTolerations("znc-que")
}

func testZyncDatabaseTolerations() []v1.Toleration {
	return getTestTolerations("zync-database")
}

func testZyncCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("123m"),
			v1.ResourceMemory: resource.MustParse("456Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("789m"),
			v1.ResourceMemory: resource.MustParse("111Mi"),
		},
	}
}

func testQueZyncCustomResourceRequirements() *v1.ResourceRequirements {
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

func testZyncDatabaseCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("666m"),
			v1.ResourceMemory: resource.MustParse("777Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("888m"),
			v1.ResourceMemory: resource.MustParse("999Mi"),
		},
	}
}

func testZyncQueSACustomImagePullSecrets() []v1.LocalObjectReference {
	return []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "mysecret9"},
		v1.LocalObjectReference{Name: "mysecret14"},
	}
}

func getZyncSecret() *v1.Secret {
	data := map[string]string{
		component.ZyncSecretKeyBaseFieldName:             zyncSecretKeyBasename,
		component.ZyncSecretDatabasePasswordFieldName:    zyncDatabasePasswd,
		component.ZyncSecretAuthenticationTokenFieldName: zyncAuthToken,
	}
	return GetTestSecret(namespace, component.ZyncSecretName, data)
}

func getZyncSecretExternalDatabase(namespace string) *v1.Secret {
	data := map[string]string{
		component.ZyncSecretKeyBaseFieldName:             zyncSecretKeyBasename,
		component.ZyncSecretDatabasePasswordFieldName:    zyncDatabasePasswd,
		component.ZyncSecretDatabaseURLFieldName:         zyncExternalDatabaseTestURL,
		component.ZyncSecretAuthenticationTokenFieldName: zyncAuthToken,
	}
	return GetTestSecret(namespace, component.ZyncSecretName, data)
}

func basicApimanagerSpecTestZyncOptions() *appsv1alpha1.APIManager {
	tmpZyncReplicas := zyncReplica
	tmpZyncQueReplicas := zyncQueReplica

	apimanager := basicApimanager()
	apimanager.Spec.Zync = &appsv1alpha1.ZyncSpec{
		AppSpec: &appsv1alpha1.ZyncAppSpec{Replicas: &tmpZyncReplicas},
		QueSpec: &appsv1alpha1.ZyncQueSpec{Replicas: &tmpZyncQueReplicas},
	}
	return apimanager
}

func basicApimanagerWithExternalZyncDatabaseSpecTestZyncOptions() *appsv1alpha1.APIManager {
	trueVal := true
	apimanager := basicApimanagerSpecTestZyncOptions()
	apimanager.Spec.HighAvailability = &appsv1alpha1.HighAvailabilitySpec{
		Enabled:                     true,
		ExternalZyncDatabaseEnabled: &trueVal,
	}
	return apimanager
}

func defaultZyncOptions(opts *component.ZyncOptions) *component.ZyncOptions {
	expectedOpts := &component.ZyncOptions{
		ImageTag:                              product.ThreescaleRelease,
		DatabaseImageTag:                      product.ThreescaleRelease,
		ContainerResourceRequirements:         component.DefaultZyncContainerResourceRequirements(),
		QueContainerResourceRequirements:      component.DefaultZyncQueContainerResourceRequirements(),
		DatabaseContainerResourceRequirements: component.DefaultZyncDatabaseContainerResourceRequirements(),
		AuthenticationToken:                   opts.AuthenticationToken,
		DatabasePassword:                      opts.DatabasePassword,
		SecretKeyBase:                         opts.SecretKeyBase,
		ZyncReplicas:                          int32(zyncReplica),
		ZyncQueReplicas:                       int32(zyncQueReplica),
		CommonLabels:                          testZyncCommonLabels(),
		CommonZyncLabels:                      testZyncZyncCommonLabels(),
		CommonZyncQueLabels:                   testZyncQueCommonLabels(),
		CommonZyncDatabaseLabels:              testZyncDatabaseCommonLabels(),
		ZyncPodTemplateLabels:                 testZyncPodTemplateLabels(),
		ZyncQuePodTemplateLabels:              testZyncQuePodTemplateCommonLabels(),
		ZyncDatabasePodTemplateLabels:         testZyncDatabasePodTemplateCommonLabels(),
		ZyncMetrics:                           true,
		ZyncQueServiceAccountImagePullSecrets: component.DefaultZyncQueServiceAccountImagePullSecrets(),
		Namespace:                             opts.Namespace,
	}

	expectedOpts.DatabaseURL = component.DefaultZyncDatabaseURL(expectedOpts.DatabasePassword)

	return expectedOpts
}

func TestGetZyncOptionsProvider(t *testing.T) {
	falseValue := false

	cases := []struct {
		testName               string
		zyncSecret             *v1.Secret
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func(*component.ZyncOptions) *component.ZyncOptions
	}{
		{"Default", nil, basicApimanagerSpecTestZyncOptions,
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				return defaultZyncOptions(opts)
			},
		},
		{"WithoutResourceRequirements", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestZyncOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			},
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.ContainerResourceRequirements = v1.ResourceRequirements{}
				expectedOpts.QueContainerResourceRequirements = v1.ResourceRequirements{}
				expectedOpts.DatabaseContainerResourceRequirements = v1.ResourceRequirements{}
				return expectedOpts
			},
		},
		{"ZyncSecret", getZyncSecret(), basicApimanagerSpecTestZyncOptions,
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.SecretKeyBase = zyncSecretKeyBasename
				expectedOpts.DatabasePassword = zyncDatabasePasswd
				expectedOpts.AuthenticationToken = zyncAuthToken
				expectedOpts.DatabaseURL = component.DefaultZyncDatabaseURL(zyncDatabasePasswd)
				return opts
			},
		},
		{"ZyncSecretWithExternalZync", getZyncSecretExternalDatabase(namespace), basicApimanagerWithExternalZyncDatabaseSpecTestZyncOptions,
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.SecretKeyBase = zyncSecretKeyBasename
				expectedOpts.DatabasePassword = zyncDatabasePasswd
				expectedOpts.AuthenticationToken = zyncAuthToken
				expectedOpts.DatabaseURL = zyncExternalDatabaseTestURL
				return opts
			},
		},
		{"WithAffinity", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestZyncOptions()
				apimanager.Spec.Zync.AppSpec.Affinity = testZyncAffinity()
				apimanager.Spec.Zync.QueSpec.Affinity = testZyncQueAffinity()
				apimanager.Spec.Zync.DatabaseAffinity = testZyncDatabaseAffinity()
				return apimanager
			},
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.ZyncAffinity = testZyncAffinity()
				expectedOpts.ZyncQueAffinity = testZyncQueAffinity()
				expectedOpts.ZyncDatabaseAffinity = testZyncDatabaseAffinity()
				return expectedOpts
			},
		},
		{"WithTolerations", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestZyncOptions()
				apimanager.Spec.Zync.AppSpec.Tolerations = testZyncTolerations()
				apimanager.Spec.Zync.QueSpec.Tolerations = testZyncQueTolerations()
				apimanager.Spec.Zync.DatabaseTolerations = testZyncDatabaseTolerations()
				return apimanager
			},
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.ZyncTolerations = testZyncTolerations()
				expectedOpts.ZyncQueTolerations = testZyncQueTolerations()
				expectedOpts.ZyncDatabaseTolerations = testZyncDatabaseTolerations()
				return expectedOpts
			},
		},
		{"WithZyncCustomResourceRequirements", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestZyncOptions()
				apimanager.Spec.Zync.AppSpec.Resources = testZyncCustomResourceRequirements()
				apimanager.Spec.Zync.QueSpec.Resources = testQueZyncCustomResourceRequirements()
				apimanager.Spec.Zync.DatabaseResources = testZyncDatabaseCustomResourceRequirements()
				return apimanager
			},
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.ContainerResourceRequirements = *testZyncCustomResourceRequirements()
				expectedOpts.QueContainerResourceRequirements = *testQueZyncCustomResourceRequirements()
				expectedOpts.DatabaseContainerResourceRequirements = *testZyncDatabaseCustomResourceRequirements()
				return expectedOpts
			},
		},
		{"WithZyncCustomResourceRequirementsAndGlobalResourceRequirementsDisabled", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestZyncOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				apimanager.Spec.Zync.AppSpec.Resources = testZyncCustomResourceRequirements()
				apimanager.Spec.Zync.QueSpec.Resources = testQueZyncCustomResourceRequirements()
				apimanager.Spec.Zync.DatabaseResources = testZyncDatabaseCustomResourceRequirements()
				return apimanager
			},
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.ContainerResourceRequirements = *testZyncCustomResourceRequirements()
				expectedOpts.QueContainerResourceRequirements = *testQueZyncCustomResourceRequirements()
				expectedOpts.DatabaseContainerResourceRequirements = *testZyncDatabaseCustomResourceRequirements()
				return expectedOpts
			},
		},
		{"WithoutResourceRequirements", nil,
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerSpecTestZyncOptions()
				apimanager.Spec.ImagePullSecrets = testZyncQueSACustomImagePullSecrets()
				return apimanager
			},
			func(opts *component.ZyncOptions) *component.ZyncOptions {
				expectedOpts := defaultZyncOptions(opts)
				expectedOpts.ZyncQueServiceAccountImagePullSecrets = testZyncQueSACustomImagePullSecrets()
				return expectedOpts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{}
			if tc.zyncSecret != nil {
				objs = append(objs, tc.zyncSecret)
			}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewZyncOptionsProvider(tc.apimanagerFactory(), namespace, cl)
			opts, err := optsProvider.GetZyncOptions()
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
