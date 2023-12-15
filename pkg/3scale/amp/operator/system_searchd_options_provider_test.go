package operator

import (
	"fmt"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"reflect"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testSearchdBasicApimanager() *appsv1alpha1.APIManager {
	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apimanagerName,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				WildcardDomain: wildcardDomain,
				AppLabel:       &[]string{appLabel}[0],
			},
		},
	}

	_, err := apimanager.SetDefaults()
	if err != nil {
		panic(fmt.Errorf("Error creating Basic APIManager: %v", err))
	}
	return apimanager
}

func testSystemSearchdLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "searchd",
	}
}

func testSystemSearchdPodTemplateLabels() map[string]string {
	labels := map[string]string{
		"app":                               appLabel,
		"threescale_component":              "system",
		"threescale_component_element":      "searchd",
		reconcilers.DeploymentLabelSelector: "system-searchd",
	}
	addExpectedMeteringLabels(labels, "system-searchd", helper.ApplicationType)

	return labels
}
func testSystemSearchdPVCOptions() component.SearchdPVCOptions {
	return component.SearchdPVCOptions{
		StorageClass:    nil,
		VolumeName:      "",
		StorageRequests: resource.MustParse("1Gi"),
	}
}

func testSystemSearchdAffinity() *v1.Affinity {
	return getTestAffinity("system-searchd")
}

func testSystemSearchdTolerations() []v1.Toleration {
	return getTestTolerations("system-searchd")
}

func testSystemSearchdCustomResourceRequirements() *v1.ResourceRequirements {
	return &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("123m"),
			v1.ResourceMemory: resource.MustParse("456Mi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("789m"),
			v1.ResourceMemory: resource.MustParse("346Mi"),
		},
	}
}

func testDefaultExpectedSystemSearchdOptions() *component.SystemSearchdOptions {
	return &component.SystemSearchdOptions{
		ImageTag:                      product.ThreescaleRelease,
		ContainerResourceRequirements: component.DefaultSearchdContainerResourceRequirements(),
		Affinity:                      nil,
		Tolerations:                   nil,
		Labels:                        testSystemSearchdLabels(),
		PodTemplateLabels:             testSystemSearchdPodTemplateLabels(),
		PVCOptions:                    testSystemSearchdPVCOptions(),
	}
}

func TestGetSystemSearchdOptionsProvider(t *testing.T) {
	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.SystemSearchdOptions
	}{
		{"Default", testSearchdBasicApimanager, testDefaultExpectedSystemSearchdOptions},
		{"ResourceRequirementsToTrue",
			func() *appsv1alpha1.APIManager {
				apimanager := testSearchdBasicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &[]bool{true}[0]
				return apimanager
			}, testDefaultExpectedSystemSearchdOptions,
		},
		{"ResourceRequirementsToFalse",
			func() *appsv1alpha1.APIManager {
				apimanager := testSearchdBasicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &[]bool{false}[0]
				return apimanager
			},
			func() *component.SystemSearchdOptions {
				expectedOpts := testDefaultExpectedSystemSearchdOptions()
				expectedOpts.ContainerResourceRequirements = corev1.ResourceRequirements{}
				return expectedOpts
			},
		},
		{"WithCustomResourceRequirementsAndGlobalResourceRequirementsDisabled",
			func() *appsv1alpha1.APIManager {
				apimanager := testSearchdBasicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &[]bool{false}[0]
				apimanager.Spec.System.SearchdSpec.Resources = testSystemSearchdCustomResourceRequirements()
				return apimanager
			},
			func() *component.SystemSearchdOptions {
				expectedOpts := testDefaultExpectedSystemSearchdOptions()
				expectedOpts.ContainerResourceRequirements = *testSystemSearchdCustomResourceRequirements()
				return expectedOpts
			},
		},
		{"WithPVC",
			func() *appsv1alpha1.APIManager {
				apimanager := testSearchdBasicApimanager()
				apimanager.Spec.System.SearchdSpec.PVC = &appsv1alpha1.PVCGenericSpec{
					StorageClassName: &[]string{"mystorageclassname"}[0],
					Resources: &appsv1alpha1.PersistentVolumeClaimResources{
						Requests: resource.MustParse("666Mi"),
					},
					VolumeName: &[]string{"myvolume"}[0],
				}
				return apimanager
			},
			func() *component.SystemSearchdOptions {
				expectedOpts := testDefaultExpectedSystemSearchdOptions()
				expectedOpts.PVCOptions = component.SearchdPVCOptions{
					StorageClass:    &[]string{"mystorageclassname"}[0],
					StorageRequests: resource.MustParse("666Mi"),
					VolumeName:      "myvolume",
				}
				return expectedOpts
			},
		},
		{"WithAffinity",
			func() *appsv1alpha1.APIManager {
				apimanager := testSearchdBasicApimanager()
				apimanager.Spec.System.SearchdSpec.Affinity = testSystemSearchdAffinity()
				return apimanager
			},
			func() *component.SystemSearchdOptions {
				expectedOpts := testDefaultExpectedSystemSearchdOptions()
				expectedOpts.Affinity = testSystemSearchdAffinity()
				return expectedOpts
			},
		},
		{"WithTolerations",
			func() *appsv1alpha1.APIManager {
				apimanager := testSearchdBasicApimanager()
				apimanager.Spec.System.SearchdSpec.Tolerations = testSystemSearchdTolerations()
				return apimanager
			},
			func() *component.SystemSearchdOptions {
				expectedOpts := testDefaultExpectedSystemSearchdOptions()
				expectedOpts.Tolerations = testSystemSearchdTolerations()
				return expectedOpts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewSystemSearchdOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetOptions()
			if err != nil {
				subT.Fatal(err)
			}
			expectedOptions := tc.expectedOptionsFactory()
			if !reflect.DeepEqual(expectedOptions, opts) {
				subT.Fatalf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}
