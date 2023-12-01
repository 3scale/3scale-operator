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
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDeploymentConfigReplicasMutator(t *testing.T) {
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
			update, err := DeploymentConfigReplicasMutator(tc.desired(), existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if existing.Spec.Replicas != tc.desired().Spec.Replicas {
				subT.Fatalf("replica reconciliation failed, existing: %d, desired: %d", existing.Spec.Replicas, tc.desired().Spec.Replicas)
			}

		})
	}
}

func TestDeploymentConfigContainerResourcesMutator(t *testing.T) {
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
			update, err := DeploymentConfigContainerResourcesMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !helper.CmpResources(&existing.Spec.Template.Spec.Containers[0].Resources, &desired.Spec.Template.Spec.Containers[0].Resources) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Containers[0].Resources, desired.Spec.Template.Spec.Containers[0].Resources, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}

func TestDeploymentConfigAffinityMutator(t *testing.T) {
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
			update, err := DeploymentConfigAffinityMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity))
			}
		})
	}
}

func TestDeploymentConfigTolerationsMutator(t *testing.T) {
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
			update, err := DeploymentConfigTolerationsMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations))
			}
		})
	}

}

func TestDeploymentConfigEnvVarReconciler(t *testing.T) {
	t.Run("DifferentNumberOfContainers", func(subT *testing.T) {
		desired := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "container1",
							},
						},
					},
				},
			},
		}
		existing := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "container1",
							},
							corev1.Container{
								Name: "container2",
							},
						},
					},
				},
			},
		}

		update := DeploymentConfigEnvVarReconciler(desired, existing, "A")
		if update {
			subT.Fatal("expected not to be updated")
		}
	})

	t.Run("DifferentNumberOfInitContainers", func(subT *testing.T) {
		desired := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "container1",
							},
						},
						InitContainers: []corev1.Container{
							corev1.Container{
								Name: "initcontainer1",
							},
						},
					},
				},
			},
		}
		existing := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "container1",
							},
						},
						InitContainers: []corev1.Container{
							corev1.Container{
								Name: "initcontainer1",
							},
							corev1.Container{
								Name: "initcontainer2",
							},
						},
					},
				},
			},
		}

		update := DeploymentConfigEnvVarReconciler(desired, existing, "A")
		if update {
			subT.Fatal("expected not to be updated")
		}
	})

	t.Run("ContainersEnvVarReconciled", func(subT *testing.T) {
		desired := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "container1",
								Env: []corev1.EnvVar{
									{Name: "A", Value: "valueA"},
								},
							},
							corev1.Container{
								Name: "container2",
								Env:  []corev1.EnvVar{},
							},
						},
					},
				},
			},
		}
		existing := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							corev1.Container{
								Name: "container1",
								Env:  []corev1.EnvVar{},
							},
							corev1.Container{
								Name: "container2",
								Env: []corev1.EnvVar{
									{Name: "A", Value: "valueA"},
								},
							},
						},
					},
				},
			},
		}

		update := DeploymentConfigEnvVarReconciler(desired, existing, "A")
		if !update {
			subT.Fatal("expected not be updated")
		}

		for i := range []int{0, 1} {
			if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[i].Env, desired.Spec.Template.Spec.Containers[i].Env) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Containers[i].Env, desired.Spec.Template.Spec.Containers[i].Env))
			}
		}

	})

	t.Run("InitContainersEnvVarReconciled", func(subT *testing.T) {
		desired := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{
							corev1.Container{
								Name: "intcontainer1",
								Env: []corev1.EnvVar{
									{Name: "A", Value: "valueA"},
								},
							},
							corev1.Container{
								Name: "intcontainer2",
								Env:  []corev1.EnvVar{},
							},
						},
					},
				},
			},
		}
		existing := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{
							corev1.Container{
								Name: "initcontainer1",
								Env:  []corev1.EnvVar{},
							},
							corev1.Container{
								Name: "initcontainer2",
								Env: []corev1.EnvVar{
									{Name: "A", Value: "valueA"},
								},
							},
						},
					},
				},
			},
		}

		update := DeploymentConfigEnvVarReconciler(desired, existing, "A")
		if !update {
			subT.Fatal("expected not be updated")
		}

		for i := range []int{0, 1} {
			if !reflect.DeepEqual(existing.Spec.Template.Spec.InitContainers[i].Env, desired.Spec.Template.Spec.InitContainers[i].Env) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.InitContainers[i].Env, desired.Spec.Template.Spec.InitContainers[i].Env))
			}
		}
	})

	t.Run("PreHookEnvVarReconciled", func(subT *testing.T) {
		desired := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{},
				},
				Strategy: appsv1.DeploymentStrategy{
					RollingParams: &appsv1.RollingDeploymentStrategyParams{
						Pre: &appsv1.LifecycleHook{
							ExecNewPod: &appsv1.ExecNewPodHook{
								Env: []corev1.EnvVar{
									{Name: "A", Value: "valueA"},
								},
							},
						},
					},
				},
			},
		}
		existing := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{},
				},
				Strategy: appsv1.DeploymentStrategy{
					RollingParams: &appsv1.RollingDeploymentStrategyParams{
						Pre: &appsv1.LifecycleHook{
							ExecNewPod: &appsv1.ExecNewPodHook{
								Env: []corev1.EnvVar{},
							},
						},
					},
				},
			},
		}

		update := DeploymentConfigEnvVarReconciler(desired, existing, "A")
		if !update {
			subT.Fatal("expected not be updated")
		}

		if !reflect.DeepEqual(existing.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env, desired.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env) {
			subT.Fatal(cmp.Diff(existing.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env, desired.Spec.Strategy.RollingParams.Pre.ExecNewPod.Env))
		}
	})

	t.Run("PostHookEnvVarReconciled", func(subT *testing.T) {
		desired := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{},
				},
				Strategy: appsv1.DeploymentStrategy{
					RollingParams: &appsv1.RollingDeploymentStrategyParams{
						Post: &appsv1.LifecycleHook{
							ExecNewPod: &appsv1.ExecNewPodHook{
								Env: []corev1.EnvVar{
									{Name: "A", Value: "valueA"},
								},
							},
						},
					},
				},
			},
		}
		existing := &appsv1.DeploymentConfig{
			TypeMeta: metav1.TypeMeta{
				Kind:       "DeploymentConfig",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDC", Namespace: "myNS"},
			Spec: appsv1.DeploymentConfigSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{},
				},
				Strategy: appsv1.DeploymentStrategy{
					RollingParams: &appsv1.RollingDeploymentStrategyParams{
						Post: &appsv1.LifecycleHook{
							ExecNewPod: &appsv1.ExecNewPodHook{
								Env: []corev1.EnvVar{},
							},
						},
					},
				},
			},
		}

		update := DeploymentConfigEnvVarReconciler(desired, existing, "A")
		if !update {
			subT.Fatal("expected not be updated")
		}

		if !reflect.DeepEqual(existing.Spec.Strategy.RollingParams.Post.ExecNewPod.Env, desired.Spec.Strategy.RollingParams.Post.ExecNewPod.Env) {
			subT.Fatal(cmp.Diff(existing.Spec.Strategy.RollingParams.Post.ExecNewPod.Env, desired.Spec.Strategy.RollingParams.Post.ExecNewPod.Env))
		}
	})
}

func TestDeploymentConfigImageChangeTriggerMutator(t *testing.T) {
	dcFactory := func(triggers []appsv1.DeploymentTriggerPolicy) *appsv1.DeploymentConfig {
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
				Triggers: triggers,
			},
		}
	}

	sliceCopy := func(a []appsv1.DeploymentTriggerPolicy) []appsv1.DeploymentTriggerPolicy {
		return append(a[:0:0], a...)
	}

	triggersA := []appsv1.DeploymentTriggerPolicy{
		{
			Type: appsv1.DeploymentTriggerOnImageChange,
			ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
				From: corev1.ObjectReference{
					Name: "imagestreamA",
				},
			},
		},
	}

	triggersB := []appsv1.DeploymentTriggerPolicy{
		{
			Type: appsv1.DeploymentTriggerOnImageChange,
			ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
				From: corev1.ObjectReference{
					Name: "imagestreamB",
				},
			},
		},
	}

	cases := []struct {
		testName         string
		existingTriggers []appsv1.DeploymentTriggerPolicy
		desiredTriggers  []appsv1.DeploymentTriggerPolicy
		expectedResult   bool
	}{
		{"NothingToReconcile", sliceCopy(triggersA), sliceCopy(triggersA), false},
		{"DifferentName", sliceCopy(triggersA), sliceCopy(triggersB), true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingTriggers)
			desired := dcFactory(tc.desiredTriggers)
			update, err := DeploymentConfigImageChangeTriggerMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			// It should be tested changes in triggers on image change only, but good enough for now
			if !reflect.DeepEqual(existing.Spec.Triggers, desired.Spec.Triggers) {
				subT.Fatal(cmp.Diff(existing.Spec.Triggers, desired.Spec.Triggers))
			}
		})
	}
}

func TestDeploymentConfigPodTemplateLabelsMutator(t *testing.T) {
	dcFactory := func(labels map[string]string) *appsv1.DeploymentConfig {
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
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
				},
			},
		}
	}

	mapCopy := func(originalMap map[string]string) map[string]string {
		// Create the target map
		targetMap := make(map[string]string)

		// Copy from the original map to the target map
		for key, value := range originalMap {
			targetMap[key] = value
		}

		return targetMap
	}

	labelsA := map[string]string{"a": "1", "a2": "2"}
	labelsB := map[string]string{"a": "other", "b": "1"}

	cases := []struct {
		testName          string
		existingLabels    map[string]string
		desiredLabels     map[string]string
		expectedResult    bool
		expectedNewLabels map[string]string
	}{
		{"NothingToReconcile", mapCopy(labelsA), mapCopy(labelsA), false, mapCopy(labelsA)},
		{"LabelsReconciled", mapCopy(labelsB), mapCopy(labelsA), true, map[string]string{
			"a": "1", "a2": "2", "b": "1",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingLabels)
			desired := dcFactory(tc.desiredLabels)
			update, err := DeploymentConfigPodTemplateLabelsMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Labels, tc.expectedNewLabels) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Labels, tc.expectedNewLabels))
			}
		})
	}
}

func TestDeploymentConfigRemoveDuplicateEnvVarMutator(t *testing.T) {
	dcFactory := func(envs []corev1.EnvVar) *appsv1.DeploymentConfig {
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
							{
								Name: "container1",
								Env:  envs,
							},
							{
								Name: "container2",
								Env:  envs,
							},
						},
					},
				},
			},
		}
	}

	envsA := []corev1.EnvVar{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}}
	envsB := []corev1.EnvVar{{Name: "a", Value: "1"}, {Name: "a", Value: "1"}, {Name: "b", Value: "2"}}

	cases := []struct {
		testName        string
		existingEnvs    []corev1.EnvVar
		expectedResult  bool
		expectedNewEnvs []corev1.EnvVar
	}{
		{"NothingToReconcile", envsA, false, envsA},
		{"EnvsReconciled", envsB, true, envsA},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingEnvs)
			update, err := DeploymentConfigRemoveDuplicateEnvVarMutator(nil, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			for idx := range existing.Spec.Template.Spec.Containers {
				if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[idx].Env, tc.expectedNewEnvs) {
					subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Containers[idx].Env, tc.expectedNewEnvs))
				}
			}
		})
	}
}

func TestDeploymentConfigTopologySpreadConstraintsMutator(t *testing.T) {
	testTopologySpreadConstraint1 := []corev1.TopologySpreadConstraint{
		corev1.TopologySpreadConstraint{
			TopologyKey:       "topologyKey1",
			WhenUnsatisfiable: "DoNotSchedule",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "3scale-api-management"},
			},
		},
	}
	testTopologySpreadConstraint2 := []corev1.TopologySpreadConstraint{
		corev1.TopologySpreadConstraint{
			TopologyKey:       "topologyKey2",
			WhenUnsatisfiable: "ScheduleAnyway",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "3scale-api-management", "threescale_component": "system"},
			},
		},
	}

	dcFactory := func(topologySpreadConstraint []corev1.TopologySpreadConstraint) *appsv1.DeploymentConfig {
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
						TopologySpreadConstraints: topologySpreadConstraint,
					},
				},
			},
		}
	}

	cases := []struct {
		testName                          string
		existingTopologySpreadConstraints []corev1.TopologySpreadConstraint
		desiredTopologySpreadConstraints  []corev1.TopologySpreadConstraint
		expectedResult                    bool
	}{
		{"NothingToReconcile", nil, nil, false},
		{"EqualTopologies", testTopologySpreadConstraint1, testTopologySpreadConstraint1, false},
		{"DifferentTopologie", testTopologySpreadConstraint1, testTopologySpreadConstraint2, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingTopologySpreadConstraints)
			desired := dcFactory(tc.desiredTopologySpreadConstraints)
			update, err := DeploymentConfigTopologySpreadConstraintsMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints))
			}
		})
	}

}

func TestDeploymentConfigPodTemplateAnnotationsMutator(t *testing.T) {
	dcFactory := func(annotations map[string]string) *appsv1.DeploymentConfig {
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
					ObjectMeta: metav1.ObjectMeta{
						Annotations: annotations,
					},
				},
			},
		}
	}

	mapCopy := func(originalMap map[string]string) map[string]string {
		// Create the target map
		targetMap := make(map[string]string)

		// Copy from the original map to the target map
		for key, value := range originalMap {
			targetMap[key] = value
		}

		return targetMap
	}

	annotationsA := map[string]string{"a": "1", "a2": "2"}
	annotationsB := map[string]string{"a": "other", "b": "1"}

	cases := []struct {
		testName               string
		existingAnnotations    map[string]string
		desiredAnnotations     map[string]string
		expectedResult         bool
		expectedNewAnnotations map[string]string
	}{
		{"NothingToReconcile", mapCopy(annotationsA), mapCopy(annotationsA), false, mapCopy(annotationsA)},
		{"AnnotationsReconciled", mapCopy(annotationsB), mapCopy(annotationsA), true, map[string]string{
			"a": "1", "a2": "2", "b": "1",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingAnnotations)
			desired := dcFactory(tc.desiredAnnotations)
			update, err := DeploymentConfigPodTemplateAnnotationsMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Annotations, tc.expectedNewAnnotations) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Annotations, tc.expectedNewAnnotations))
			}
		})
	}
}

func TestDeploymentConfigArgsMutator(t *testing.T) {
	type args struct {
		desired  *appsv1.DeploymentConfig
		existing *appsv1.DeploymentConfig
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "No Args Update Required",
			args: args{
				desired: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Args: []string{"testArg"}},
								},
							},
						},
					},
				},
				existing: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Args: []string{"testArg"}},
								},
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Args Update Required",
			args: args{
				desired: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Args: []string{"testArg1", "testArg2"}},
								},
							},
						},
					},
				},
				existing: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Args: []string{"testArg1"}},
								},
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeploymentConfigArgsMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentConfigArgsMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentConfigArgsMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeploymentConfigProbesMutator(t *testing.T) {
	type args struct {
		desired  *appsv1.DeploymentConfig
		existing *appsv1.DeploymentConfig
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Liveness Probe Updated",
			args: args{
				desired: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										LivenessProbe: &v1.Probe{
											ProbeHandler: v1.ProbeHandler{
												TCPSocket: &v1.TCPSocketAction{
													Port: intstr.FromInt(9306),
												},
											},
											InitialDelaySeconds: 60,
											PeriodSeconds:       10,
										},
									},
								},
							},
						},
					},
				},
				existing: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										LivenessProbe: nil,
									},
								},
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Liveness Probe Not Updated",
			args: args{
				desired: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										LivenessProbe: &v1.Probe{
											ProbeHandler: v1.ProbeHandler{
												TCPSocket: &v1.TCPSocketAction{
													Port: intstr.FromInt(9306),
												},
											},
											InitialDelaySeconds: 60,
											PeriodSeconds:       10,
										},
									},
								},
							},
						},
					},
				},
				existing: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										LivenessProbe: &v1.Probe{
											ProbeHandler: v1.ProbeHandler{
												TCPSocket: &v1.TCPSocketAction{
													Port: intstr.FromInt(9306),
												},
											},
											InitialDelaySeconds: 60,
											PeriodSeconds:       10,
										},
									},
								},
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Readiness Probe Updated",
			args: args{
				desired: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										ReadinessProbe: &v1.Probe{
											ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
												Path: "/status",
												Port: intstr.IntOrString{
													Type:   intstr.Type(intstr.Int),
													IntVal: 3000}},
											},
											InitialDelaySeconds: 30,
											TimeoutSeconds:      5,
										},
									},
								},
							},
						},
					},
				},
				existing: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										ReadinessProbe: nil,
									},
								},
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Readiness Probe Not Updated",
			args: args{
				desired: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										ReadinessProbe: &v1.Probe{
											ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
												Path: "/status",
												Port: intstr.IntOrString{
													Type:   intstr.Type(intstr.Int),
													IntVal: 3000}},
											},
											InitialDelaySeconds: 30,
											TimeoutSeconds:      5,
										},
									},
								},
							},
						},
					},
				},
				existing: &appsv1.DeploymentConfig{
					Spec: appsv1.DeploymentConfigSpec{
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										ReadinessProbe: &v1.Probe{
											ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
												Path: "/status",
												Port: intstr.IntOrString{
													Type:   intstr.Type(intstr.Int),
													IntVal: 3000}},
											},
											InitialDelaySeconds: 30,
											TimeoutSeconds:      5,
										},
									},
								},
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeploymentConfigProbesMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentConfigProbesMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentConfigProbesMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}
