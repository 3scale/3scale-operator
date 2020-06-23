package reconcilers

import (
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentConfigReplicasReconciler(t *testing.T) {
	dcFactory := func() *appsv1.DeploymentConfig {
		return &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDC",
				Namespace: "myNS",
			},
			Spec: appsv1.DeploymentConfigSpec{
				Replicas: 3,
			},
		}
	}

	cases := []struct {
		testName       string
		desired        func() *appsv1.DeploymentConfig
		expectedResult bool
	}{
		{"NothingToReconcile", func() *appsv1.DeploymentConfig { return dcFactory() }, false},
		{"ReplicasReconcile",
			func() *appsv1.DeploymentConfig {
				desired := dcFactory()
				desired.Spec.Replicas = desired.Spec.Replicas + 1000
				return desired
			}, true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory()
			update := DeploymentConfigReplicasReconciler(tc.desired(), existing)
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if existing.Spec.Replicas != tc.desired().Spec.Replicas {
				subT.Fatalf("replica reconciliation failed, existing: %d, desired: %d", existing.Spec.Replicas, tc.desired().Spec.Replicas)
			}

		})
	}
}

func TestDeploymentConfigContainerResourcesReconciler(t *testing.T) {
	emptyResourceRequirements := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}
	notEmptyResources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("110Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("200m"),
			corev1.ResourceMemory: resource.MustParse("220Mi"),
		},
	}
	dcFactory := func(resources corev1.ResourceRequirements) *appsv1.DeploymentConfig {
		return &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDC",
				Namespace: "myNS",
			},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name:      "container1",
								Resources: resources,
							},
						},
					},
				},
			},
		}
	}

	cases := []struct {
		testName          string
		existingResources corev1.ResourceRequirements
		desiredResources  corev1.ResourceRequirements
		expectedResult    bool
	}{
		{"NothingToReconcile", emptyResourceRequirements, emptyResourceRequirements, false},
		{"NothingToReconcileWithResources", notEmptyResources, notEmptyResources, false},
		{"AddResources", emptyResourceRequirements, notEmptyResources, true},
		{"RemoveResources", notEmptyResources, emptyResourceRequirements, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingResources)
			desired := dcFactory(tc.desiredResources)
			update := DeploymentConfigContainerResourcesReconciler(desired, existing)
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !helper.CmpResources(&existing.Spec.Template.Spec.Containers[0].Resources, &desired.Spec.Template.Spec.Containers[0].Resources) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Containers[0].Resources, desired.Spec.Template.Spec.Containers[0].Resources, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}

func TestDeploymentConfigAffinityReconciler(t *testing.T) {
	testAffinity1 := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					corev1.NodeSelectorTerm{
						MatchFields: []corev1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      "key1",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"val1"},
							},
						},
					},
				},
			},
		},
	}
	testAffinity2 := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					corev1.NodeSelectorTerm{
						MatchFields: []corev1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      "key2",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"val2"},
							},
						},
					},
				},
			},
		},
	}
	dcFactory := func(affinity *corev1.Affinity) *appsv1.DeploymentConfig {
		return &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDC",
				Namespace: "myNS",
			},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Affinity: affinity,
					},
				},
			},
		}
	}

	cases := []struct {
		testName         string
		existingAffinity *corev1.Affinity
		desiredAffinity  *corev1.Affinity
		expectedResult   bool
	}{
		{"NothingToReconcile", nil, nil, false},
		{"EqualAffinities", testAffinity1, testAffinity1, false},
		{"DifferentAffinities", testAffinity1, testAffinity2, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingAffinity)
			desired := dcFactory(tc.desiredAffinity)
			update := DeploymentConfigAffinityReconciler(desired, existing)
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity))
			}
		})
	}
}

func TestDeploymentConfigTolerationsReconciler(t *testing.T) {
	testTolerations1 := []corev1.Toleration{
		corev1.Toleration{
			Key:      "key1",
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOpEqual,
			Value:    "val1",
		},
		corev1.Toleration{
			Key:      "key2",
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOpEqual,
			Value:    "val2",
		},
	}
	testTolerations2 := []corev1.Toleration{
		corev1.Toleration{
			Key:      "key3",
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOpEqual,
			Value:    "val3",
		},
		corev1.Toleration{
			Key:      "key4",
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOpEqual,
			Value:    "val4",
		},
	}
	dcFactory := func(toleration []corev1.Toleration) *appsv1.DeploymentConfig {
		return &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDC",
				Namespace: "myNS",
			},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Tolerations: toleration,
					},
				},
			},
		}
	}

	cases := []struct {
		testName            string
		existingTolerations []corev1.Toleration
		desiredTolerations  []corev1.Toleration
		expectedResult      bool
	}{
		{"NothingToReconcile", nil, nil, false},
		{"EqualAffinities", testTolerations1, testTolerations1, false},
		{"DifferentAffinities", testTolerations1, testTolerations2, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingTolerations)
			desired := dcFactory(tc.desiredTolerations)
			update := DeploymentConfigTolerationsReconciler(desired, existing)
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations))
			}
		})
	}

}
