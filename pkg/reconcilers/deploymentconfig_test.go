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
							corev1.Container{
								Name: "container1",
								Env:  envs,
							},
						},
					},
				},
			},
		}
	}

	sliceCopy := func(a []corev1.EnvVar) []corev1.EnvVar {
		return append(a[:0:0], a...)
	}

	envVarAB := []corev1.EnvVar{
		{
			Name:  "A",
			Value: "valueA",
		},
		{
			Name:  "B",
			Value: "valueB",
		},
	}

	envVarB := []corev1.EnvVar{
		{
			Name:  "B",
			Value: "valueB",
		},
	}

	envVarBA := []corev1.EnvVar{
		{
			Name:  "B",
			Value: "valueB",
		},
		{
			Name:  "A",
			Value: "valueA",
		},
	}

	envVarAB2 := []corev1.EnvVar{
		{
			Name:  "A",
			Value: "valueOther",
		},
		{
			Name:  "B",
			Value: "valueB",
		},
	}

	cases := []struct {
		testName          string
		existingEnvVar    []corev1.EnvVar
		desiredEnvVar     []corev1.EnvVar
		expectedResult    bool
		expectedNewEnvVar []corev1.EnvVar
	}{
		{"NothingToReconcile", sliceCopy(envVarAB), sliceCopy(envVarAB), false, sliceCopy(envVarAB)},
		{"MissingEnvVar", sliceCopy(envVarB), sliceCopy(envVarAB), true, sliceCopy(envVarBA)},
		{"UpdatedEnvVar", sliceCopy(envVarAB), sliceCopy(envVarAB2), true, sliceCopy(envVarAB2)},
		{"RemovedEnvVar", sliceCopy(envVarAB), sliceCopy(envVarB), true, sliceCopy(envVarB)},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingEnvVar)
			desired := dcFactory(tc.desiredEnvVar)
			update := DeploymentConfigEnvVarReconciler(desired, existing, "A")
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].Env, tc.expectedNewEnvVar) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.Containers[0].Env, tc.expectedNewEnvVar))
			}
		})
	}
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
			// It should be tested changes in triggers on image change only, but good enough for now
			if !reflect.DeepEqual(existing.Spec.Template.Labels, tc.expectedNewLabels) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Labels, tc.expectedNewLabels))
			}
		})
	}
}

func TestDeploymentConfigEnvVarSyncMutator(t *testing.T) {
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
		desiredEnvs     []corev1.EnvVar
		expectedResult  bool
		expectedNewEnvs []corev1.EnvVar
	}{
		{"NothingToReconcile", envsA, envsA, false, envsA},
		{"EnvsReconciled", envsB, envsA, true, envsA},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dcFactory(tc.existingEnvs)
			desired := dcFactory(tc.desiredEnvs)
			update, err := DeploymentConfigRemoveDuplicateEnvVarMutator(desired, existing)
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
