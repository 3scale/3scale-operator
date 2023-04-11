package operator

import (
	"fmt"
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

func testSphinxBasicApimanager() *appsv1alpha1.APIManager {
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

func testSystemSphinxLabels() map[string]string {
	return map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "sphinx",
	}
}

func testSystemSphinxPodTemplateLabels() map[string]string {
	labels := map[string]string{
		"app":                          appLabel,
		"threescale_component":         "system",
		"threescale_component_element": "sphinx",
		"deploymentConfig":             "system-sphinx",
	}
	addExpectedMeteringLabels(labels, "system-sphinx", helper.ApplicationType)

	return labels
}
func testSystemSphinxPVCOptions() component.SphinxPVCOptions {
	return component.SphinxPVCOptions{
		StorageClass:    nil,
		VolumeName:      "",
		StorageRequests: resource.MustParse("1Gi"),
	}
}

func testSystemSphinxAffinity() *v1.Affinity {
	return getTestAffinity("system-sphinx")
}

func testSystemSphinxTolerations() []v1.Toleration {
	return getTestTolerations("system-sphinx")
}

func testSystemSphinxCustomResourceRequirements() *v1.ResourceRequirements {
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

func testDefaultExpectedSystemSphinxOptions() *component.SystemSphinxOptions {
	return &component.SystemSphinxOptions{
		ImageTag:                      product.ThreescaleRelease,
		ContainerResourceRequirements: component.DefaultSphinxContainerResourceRequirements(),
		Affinity:                      nil,
		Tolerations:                   nil,
		Labels:                        testSystemSphinxLabels(),
		PodTemplateLabels:             testSystemSphinxPodTemplateLabels(),
		PVCOptions:                    testSystemSphinxPVCOptions(),
	}
}

func TestGetSystemSphinxOptionsProvider(t *testing.T) {
	cases := []struct {
		testName               string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.SystemSphinxOptions
	}{
		{"Default", testSphinxBasicApimanager, testDefaultExpectedSystemSphinxOptions},
		{"ResourceRequirementsToTrue",
			func() *appsv1alpha1.APIManager {
				apimanager := testSphinxBasicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &[]bool{true}[0]
				return apimanager
			}, testDefaultExpectedSystemSphinxOptions,
		},
		{"ResourceRequirementsToFalse",
			func() *appsv1alpha1.APIManager {
				apimanager := testSphinxBasicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &[]bool{false}[0]
				return apimanager
			},
			func() *component.SystemSphinxOptions {
				expectedOpts := testDefaultExpectedSystemSphinxOptions()
				expectedOpts.ContainerResourceRequirements = corev1.ResourceRequirements{}
				return expectedOpts
			},
		},
		{"WithCustomResourceRequirementsAndGlobalResourceRequirementsDisabled",
			func() *appsv1alpha1.APIManager {
				apimanager := testSphinxBasicApimanager()
				apimanager.Spec.ResourceRequirementsEnabled = &[]bool{false}[0]
				apimanager.Spec.System.SphinxSpec.Resources = testSystemSphinxCustomResourceRequirements()
				return apimanager
			},
			func() *component.SystemSphinxOptions {
				expectedOpts := testDefaultExpectedSystemSphinxOptions()
				expectedOpts.ContainerResourceRequirements = *testSystemSphinxCustomResourceRequirements()
				return expectedOpts
			},
		},
		{"WithPVC",
			func() *appsv1alpha1.APIManager {
				apimanager := testSphinxBasicApimanager()
				apimanager.Spec.System.SphinxSpec.PVC = &appsv1alpha1.SystemSphinxPVCSpec{
					StorageClassName: &[]string{"mystorageclassname"}[0],
					Resources: &appsv1alpha1.PersistentVolumeClaimResources{
						Requests: resource.MustParse("666Mi"),
					},
					VolumeName: &[]string{"myvolume"}[0],
				}
				return apimanager
			},
			func() *component.SystemSphinxOptions {
				expectedOpts := testDefaultExpectedSystemSphinxOptions()
				expectedOpts.PVCOptions = component.SphinxPVCOptions{
					StorageClass:    &[]string{"mystorageclassname"}[0],
					StorageRequests: resource.MustParse("666Mi"),
					VolumeName:      "myvolume",
				}
				return expectedOpts
			},
		},
		{"WithAffinity",
			func() *appsv1alpha1.APIManager {
				apimanager := testSphinxBasicApimanager()
				apimanager.Spec.System.SphinxSpec.Affinity = testSystemSphinxAffinity()
				return apimanager
			},
			func() *component.SystemSphinxOptions {
				expectedOpts := testDefaultExpectedSystemSphinxOptions()
				expectedOpts.Affinity = testSystemSphinxAffinity()
				return expectedOpts
			},
		},
		{"WithTolerations",
			func() *appsv1alpha1.APIManager {
				apimanager := testSphinxBasicApimanager()
				apimanager.Spec.System.SphinxSpec.Tolerations = testSystemSphinxTolerations()
				return apimanager
			},
			func() *component.SystemSphinxOptions {
				expectedOpts := testDefaultExpectedSystemSphinxOptions()
				expectedOpts.Tolerations = testSystemSphinxTolerations()
				return expectedOpts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			optsProvider := NewSystemSphinxOptionsProvider(tc.apimanagerFactory())
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
