package operator

import (
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"
)

const (
	productionReplicaCount int64 = 3
	stagingReplicaCount    int64 = 4
	apicastManagementAPI         = "disabled"
	openSSLVerify                = false
	responseCodes                = true
)

func testApicastCommonLabels() map[string]string {
	return map[string]string{
		"app":                  appLabel,
		"threescale_component": "apicast",
	}
}

func testApicastStagingLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "apicast",
		"threescale_component_element": "staging",
	}
}

func testApicastProductionLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "apicast",
		"threescale_component_element": "production",
	}
}

func testApicastStagingPodLabels() map[string]string {
	labels := map[string]string{
		"app":                               appLabel,
		"threescale_component":              "apicast",
		"threescale_component_element":      "staging",
		reconcilers.DeploymentLabelSelector: "apicast-staging",
	}
	addExpectedMeteringLabels(labels, "apicast-staging", helper.ApplicationType)

	return labels
}

func testApicastProductionPodLabels() map[string]string {
	labels := map[string]string{
		"app":                               appLabel,
		"threescale_component":              "apicast",
		"threescale_component_element":      "production",
		reconcilers.DeploymentLabelSelector: "apicast-production",
	}
	addExpectedMeteringLabels(labels, "apicast-production", helper.ApplicationType)

	return labels
}

func testApicastStagingAffinity() *v1.Affinity {
	return getTestAffinity("apicast-staging")
}

func testApicastProductionAffinity() *v1.Affinity {
	return getTestAffinity("apicast-production")
}

func testApicastStagingTolerations() []v1.Toleration {
	return getTestTolerations("apicast-staging")
}

func testApicastProductionTolerations() []v1.Toleration {
	return getTestTolerations("apicast-production")
}

func testApicastStagingCustomResourceRequirements() *v1.ResourceRequirements {
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

func testApicastProductionCustomResourceRequirements() *v1.ResourceRequirements {
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

func basicApimanagerTestApicastOptions() *appsv1alpha1.APIManager {
	tmpApicastManagementAPI := apicastManagementAPI
	tmpOpenSSLVerify := openSSLVerify
	tmpResponseCodes := responseCodes
	tmpProductionReplicaCount := productionReplicaCount
	tmpStagingReplicaCount := stagingReplicaCount

	apimanager := basicApimanager()
	apimanager.Spec.Apicast = &appsv1alpha1.ApicastSpec{
		ApicastManagementAPI: &tmpApicastManagementAPI,
		OpenSSLVerify:        &tmpOpenSSLVerify,
		IncludeResponseCodes: &tmpResponseCodes,
		StagingSpec: &appsv1alpha1.ApicastStagingSpec{
			Replicas:      &tmpStagingReplicaCount,
			OpenTelemetry: &appsv1alpha1.OpenTelemetrySpec{},
		},
		ProductionSpec: &appsv1alpha1.ApicastProductionSpec{
			Replicas:      &tmpProductionReplicaCount,
			OpenTelemetry: &appsv1alpha1.OpenTelemetrySpec{},
		},
	}
	return apimanager
}

func defaultApicastOptions() *component.ApicastOptions {
	return &component.ApicastOptions{
		ManagementAPI:                      apicastManagementAPI,
		OpenSSLVerify:                      strconv.FormatBool(openSSLVerify),
		ResponseCodes:                      strconv.FormatBool(responseCodes),
		ImageTag:                           product.ThreescaleRelease,
		ExtendedMetrics:                    true,
		ProductionResourceRequirements:     component.DefaultProductionResourceRequirements(),
		StagingResourceRequirements:        component.DefaultStagingResourceRequirements(),
		ProductionReplicas:                 int32(productionReplicaCount),
		StagingReplicas:                    int32(stagingReplicaCount),
		CommonLabels:                       testApicastCommonLabels(),
		CommonStagingLabels:                testApicastStagingLabels(),
		CommonProductionLabels:             testApicastProductionLabels(),
		StagingPodTemplateLabels:           testApicastStagingPodLabels(),
		ProductionPodTemplateLabels:        testApicastProductionPodLabels(),
		Namespace:                          namespace,
		ProductionTracingConfig:            &component.APIcastTracingConfig{TracingLibrary: apps.APIcastDefaultTracingLibrary},
		StagingTracingConfig:               &component.APIcastTracingConfig{TracingLibrary: apps.APIcastDefaultTracingLibrary},
		StagingOpentelemetry:               component.OpentelemetryConfig{},
		ProductionOpentelemetry:            component.OpentelemetryConfig{},
		StagingAdditionalPodAnnotations:    map[string]string{APIcastEnvironmentCMAnnotation: "788712912"},
		ProductionAdditionalPodAnnotations: map[string]string{APIcastEnvironmentCMAnnotation: "788712912"},
	}
}

func otlpSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-secret-name",
			Namespace:       "someNS",
			ResourceVersion: "999",
		},
		Data: map[string][]byte{
			"config.json": []byte(`
			  exporter = "otlp"
			  processor = "simple"
			  [exporters.otlp]
			  # Alternatively the OTEL_EXPORTER_OTLP_ENDPOINT environment variable can also be used.
			  host = "jaeger"
			  port = 4317
			  # Optional: enable SSL, for endpoints that support it
			  # use_ssl = true
			  # Optional: set a filesystem path to a pem file to be used for SSL encryption
			  # (when use_ssl = true)
			  # ssl_cert_path = "/path/to/cert.pem"
			  [processors.batch]
			  max_queue_size = 2048
			  schedule_delay_millis = 5000
			  max_export_batch_size = 512
			  [service]
			  name = "apicast" # Opentelemetry resource name,
			  `),
		},
	}
}

func TestGetApicastOptionsProvider(t *testing.T) {
	falseValue := false

	cases := []struct {
		name                   string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.ApicastOptions
	}{
		{"Default", basicApimanagerTestApicastOptions, defaultApicastOptions},
		{"WithGlobalResourceRequirementsDisabled",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionResourceRequirements = v1.ResourceRequirements{}
				opts.StagingResourceRequirements = v1.ResourceRequirements{}
				return opts
			},
		},
		{"WithAffinity",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.Apicast.ProductionSpec.Affinity = testApicastProductionAffinity()
				apimanager.Spec.Apicast.StagingSpec.Affinity = testApicastStagingAffinity()
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionAffinity = testApicastProductionAffinity()
				opts.StagingAffinity = testApicastStagingAffinity()
				return opts
			},
		},
		{"WithTolerations",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.Apicast.ProductionSpec.Tolerations = testApicastProductionTolerations()
				apimanager.Spec.Apicast.StagingSpec.Tolerations = testApicastStagingTolerations()
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionTolerations = testApicastProductionTolerations()
				opts.StagingTolerations = testApicastStagingTolerations()
				return opts
			},
		},
		{"WithAPIcastCustomResourceRequirements",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.Apicast.ProductionSpec.Resources = testApicastProductionCustomResourceRequirements()
				apimanager.Spec.Apicast.StagingSpec.Resources = testApicastStagingCustomResourceRequirements()
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionResourceRequirements = *testApicastProductionCustomResourceRequirements()
				opts.StagingResourceRequirements = *testApicastStagingCustomResourceRequirements()
				return opts
			},
		},
		{"WithAPIcastCustomResourceRequirementsAndGlobalResourceRequirementsDisabled",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				apimanager.Spec.ResourceRequirementsEnabled = &falseValue
				apimanager.Spec.Apicast.ProductionSpec.Resources = testApicastProductionCustomResourceRequirements()
				apimanager.Spec.Apicast.StagingSpec.Resources = testApicastStagingCustomResourceRequirements()
				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()
				opts.ProductionResourceRequirements = *testApicastProductionCustomResourceRequirements()
				opts.StagingResourceRequirements = *testApicastStagingCustomResourceRequirements()
				return opts
			},
		},
		{"WithServiceCacheSize",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()
				var stagingCacheSize int32 = 10
				var productionCacheSize int32 = 20

				apimanager.Spec.Apicast.ProductionSpec.ServiceCacheSize = &productionCacheSize
				apimanager.Spec.Apicast.StagingSpec.ServiceCacheSize = &stagingCacheSize

				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()

				var stagingCacheSize int32 = 10
				var productionCacheSize int32 = 20

				opts.ProductionServiceCacheSize = &productionCacheSize
				opts.StagingServiceCacheSize = &stagingCacheSize

				return opts
			},
		},
		{"WithApicastStagingTelemtryConfigurationWithCustomMountPath",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()

				var trueOpenTelemetry = true
				var opentelemtryKey = "some-key"

				apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.Enabled = &trueOpenTelemetry
				apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretRef = &v1.LocalObjectReference{
					Name: "my-secret-name",
				}
				apimanager.Spec.Apicast.StagingSpec.OpenTelemetry.TracingConfigSecretKey = &opentelemtryKey

				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()

				var trueOpenTelemetry bool = true
				var opentelemtryKey string = component.OpentelemetryConfigMountBasePath + "/some-key"

				opts.StagingOpentelemetry.Enabled = trueOpenTelemetry
				opts.StagingOpentelemetry.Secret = *otlpSecret()
				opts.StagingOpentelemetry.ConfigFile = opentelemtryKey

				return opts
			},
		},
		{"WithApicastProductionTelemtryConfigurationWithDefaultMountPath",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanagerTestApicastOptions()

				var trueOpenTelemetry = true

				apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.Enabled = &trueOpenTelemetry
				apimanager.Spec.Apicast.ProductionSpec.OpenTelemetry.TracingConfigSecretRef = &v1.LocalObjectReference{
					Name: "my-secret-name",
				}

				return apimanager
			},
			func() *component.ApicastOptions {
				opts := defaultApicastOptions()

				var trueOpenTelemetry bool = true
				var opentelemtryKey string = component.OpentelemetryConfigMountBasePath + "/config.json"

				opts.ProductionOpentelemetry.Enabled = trueOpenTelemetry
				opts.ProductionOpentelemetry.Secret = *otlpSecret()
				opts.ProductionOpentelemetry.ConfigFile = opentelemtryKey

				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			objs := []runtime.Object{
				otlpSecret(),
			}
			cl := fake.NewFakeClient(objs...)
			optsProvider := NewApicastOptionsProvider(tc.apimanagerFactory(), cl)
			opts, err := optsProvider.GetApicastOptions()
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
