package reconcilers

import (
	"reflect"
	"testing"

	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestDeploymentReplicasMutator(t *testing.T) {
	numReplicas := int32(3)
	dFactory := func() *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDeployment",
				Namespace: "myNS",
			},
			Spec: k8sappsv1.DeploymentSpec{
				Replicas: &numReplicas,
			},
		}
	}

	cases := []struct {
		testName       string
		desired        func() *k8sappsv1.Deployment
		expectedResult bool
	}{
		{"NothingToReconcile", func() *k8sappsv1.Deployment { return dFactory() }, false},
		{
			"ReplicasReconcile",
			func() *k8sappsv1.Deployment {
				desired := dFactory()
				newNumReplicas := *desired.Spec.Replicas + int32(1000)
				desired.Spec.Replicas = &newNumReplicas
				return desired
			}, true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := dFactory()
			update, err := DeploymentReplicasMutator(tc.desired(), existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if *existing.Spec.Replicas != *tc.desired().Spec.Replicas {
				subT.Fatalf("replica reconciliation failed, existing: %d, desired: %d", existing.Spec.Replicas, tc.desired().Spec.Replicas)
			}
		})
	}
}

func TestDeploymentContainerResourcesMutator(t *testing.T) {
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
	dFactory := func(resources corev1.ResourceRequirements) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDeployment",
				Namespace: "myNS",
			},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
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
			existing := dFactory(tc.existingResources)
			desired := dFactory(tc.desiredResources)
			update, err := DeploymentContainerResourcesMutator(desired, existing)
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

func TestDeploymentAffinityMutator(t *testing.T) {
	testAffinity1 := &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchFields: []corev1.NodeSelectorRequirement{
							{
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
					{
						MatchFields: []corev1.NodeSelectorRequirement{
							{
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
	dFactory := func(affinity *corev1.Affinity) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDeployment",
				Namespace: "myNS",
			},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
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
			existing := dFactory(tc.existingAffinity)
			desired := dFactory(tc.desiredAffinity)
			update, err := DeploymentAffinityMutator(desired, existing)
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

func TestDeploymentTolerationsMutator(t *testing.T) {
	testTolerations1 := []corev1.Toleration{
		{
			Key:      "key1",
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOpEqual,
			Value:    "val1",
		},
		{
			Key:      "key2",
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOpEqual,
			Value:    "val2",
		},
	}
	testTolerations2 := []corev1.Toleration{
		{
			Key:      "key3",
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOpEqual,
			Value:    "val3",
		},
		{
			Key:      "key4",
			Effect:   corev1.TaintEffectNoExecute,
			Operator: corev1.TolerationOpEqual,
			Value:    "val4",
		},
	}
	dFactory := func(toleration []corev1.Toleration) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDeployment",
				Namespace: "myNS",
			},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
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
			existing := dFactory(tc.existingTolerations)
			desired := dFactory(tc.desiredTolerations)
			update, err := DeploymentTolerationsMutator(desired, existing)
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

func TestDeploymentEnvVarReconciler(t *testing.T) {
	t.Run("DifferentNumberOfContainers", func(subT *testing.T) {
		desired := &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDeployment", Namespace: "myNS"},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container1",
							},
						},
					},
				},
			},
		}
		existing := &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDeployment", Namespace: "myNS"},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container1",
							},
							{
								Name: "container2",
							},
						},
					},
				},
			},
		}

		update := DeploymentEnvVarReconciler(desired, existing, "A")
		if update {
			subT.Fatal("expected not to be updated")
		}
	})

	t.Run("DifferentNumberOfInitContainers", func(subT *testing.T) {
		desired := &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDeployment", Namespace: "myNS"},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container1",
							},
						},
						InitContainers: []corev1.Container{
							{
								Name: "initcontainer1",
							},
						},
					},
				},
			},
		}
		existing := &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDeployment", Namespace: "myNS"},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container1",
							},
						},
						InitContainers: []corev1.Container{
							{
								Name: "initcontainer1",
							},
							{
								Name: "initcontainer2",
							},
						},
					},
				},
			},
		}

		update := DeploymentEnvVarReconciler(desired, existing, "A")
		if update {
			subT.Fatal("expected not to be updated")
		}
	})

	t.Run("ContainersEnvVarReconciled", func(subT *testing.T) {
		desired := &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDeployment", Namespace: "myNS"},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container1",
								Env: []corev1.EnvVar{
									{Name: "A", Value: "valueA"},
								},
							},
							{
								Name: "container2",
								Env:  []corev1.EnvVar{},
							},
						},
					},
				},
			},
		}
		existing := &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDeployment", Namespace: "myNS"},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "container1",
								Env:  []corev1.EnvVar{},
							},
							{
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

		update := DeploymentEnvVarReconciler(desired, existing, "A")
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
		desired := &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDeployment", Namespace: "myNS"},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{
							{
								Name: "intcontainer1",
								Env: []corev1.EnvVar{
									{Name: "A", Value: "valueA"},
								},
							},
							{
								Name: "intcontainer2",
								Env:  []corev1.EnvVar{},
							},
						},
					},
				},
			},
		}
		existing := &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{Name: "myDeployment", Namespace: "myNS"},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{
							{
								Name: "initcontainer1",
								Env:  []corev1.EnvVar{},
							},
							{
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

		update := DeploymentEnvVarReconciler(desired, existing, "A")
		if !update {
			subT.Fatal("expected not be updated")
		}

		for i := range []int{0, 1} {
			if !reflect.DeepEqual(existing.Spec.Template.Spec.InitContainers[i].Env, desired.Spec.Template.Spec.InitContainers[i].Env) {
				subT.Fatal(cmp.Diff(existing.Spec.Template.Spec.InitContainers[i].Env, desired.Spec.Template.Spec.InitContainers[i].Env))
			}
		}
	})
}

func TestDeploymentPodTemplateLabelsMutator(t *testing.T) {
	dFactory := func(labels map[string]string) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDeployment",
				Namespace: "myNS",
			},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
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
			existing := dFactory(tc.existingLabels)
			desired := dFactory(tc.desiredLabels)
			update, err := DeploymentPodTemplateLabelsMutator(desired, existing)
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

func TestDeploymentRemoveDuplicateEnvVarMutator(t *testing.T) {
	dFactory := func(envs []corev1.EnvVar) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDeployment",
				Namespace: "myNS",
			},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
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
			existing := dFactory(tc.existingEnvs)
			update, err := DeploymentRemoveDuplicateEnvVarMutator(nil, existing)
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

func TestDeploymentTopologySpreadConstraintsMutator(t *testing.T) {
	testTopologySpreadConstraint1 := []corev1.TopologySpreadConstraint{
		{
			TopologyKey:       "topologyKey1",
			WhenUnsatisfiable: "DoNotSchedule",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "3scale-api-management"},
			},
		},
	}
	testTopologySpreadConstraint2 := []corev1.TopologySpreadConstraint{
		{
			TopologyKey:       "topologyKey2",
			WhenUnsatisfiable: "ScheduleAnyway",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "3scale-api-management", "threescale_component": "system"},
			},
		},
	}

	dFactory := func(topologySpreadConstraint []corev1.TopologySpreadConstraint) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDeployment",
				Namespace: "myNS",
			},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
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
			existing := dFactory(tc.existingTopologySpreadConstraints)
			desired := dFactory(tc.desiredTopologySpreadConstraints)
			update, err := DeploymentTopologySpreadConstraintsMutator(desired, existing)
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

func TestDeploymentPodTemplateAnnotationsMutator(t *testing.T) {
	dFactory := func(annotations map[string]string) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myDeployment",
				Namespace: "myNS",
			},
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
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
			existing := dFactory(tc.existingAnnotations)
			desired := dFactory(tc.desiredAnnotations)
			update, err := DeploymentPodTemplateAnnotationsMutator(desired, existing)
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

func TestDeploymentArgsMutator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
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
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Args: []string{"testArg"}},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
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
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Args: []string{"testArg1", "testArg2"}},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeploymentArgsMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentArgsMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentArgsMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeploymentProbesMutator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
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
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										LivenessProbe: &corev1.Probe{
											ProbeHandler: corev1.ProbeHandler{
												TCPSocket: &corev1.TCPSocketAction{
													Port: intstr.FromInt32(9306),
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
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
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
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										LivenessProbe: &corev1.Probe{
											ProbeHandler: corev1.ProbeHandler{
												TCPSocket: &corev1.TCPSocketAction{
													Port: intstr.FromInt32(9306),
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
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										LivenessProbe: &corev1.Probe{
											ProbeHandler: corev1.ProbeHandler{
												TCPSocket: &corev1.TCPSocketAction{
													Port: intstr.FromInt32(9306),
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
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										ReadinessProbe: &corev1.Probe{
											ProbeHandler: corev1.ProbeHandler{
												HTTPGet: &corev1.HTTPGetAction{
													Path: "/status",
													Port: intstr.IntOrString{
														Type:   intstr.Int,
														IntVal: 3000,
													},
												},
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
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
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
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										ReadinessProbe: &corev1.Probe{
											ProbeHandler: corev1.ProbeHandler{
												HTTPGet: &corev1.HTTPGetAction{
													Path: "/status",
													Port: intstr.IntOrString{
														Type:   intstr.Int,
														IntVal: 3000,
													},
												},
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
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										ReadinessProbe: &corev1.Probe{
											ProbeHandler: corev1.ProbeHandler{
												HTTPGet: &corev1.HTTPGetAction{
													Path: "/status",
													Port: intstr.IntOrString{
														Type:   intstr.Int,
														IntVal: 3000,
													},
												},
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
			got, err := DeploymentProbesMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentProbesMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentProbesMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeploymentAnnotationsMutator(t *testing.T) {
	dFactory := func(annotations map[string]string) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps.openshift.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        "myDeployment",
				Namespace:   "myNS",
				Annotations: annotations,
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
			existing := dFactory(tc.existingAnnotations)
			desired := dFactory(tc.desiredAnnotations)
			update, err := DeploymentAnnotationsMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.ObjectMeta.Annotations, tc.expectedNewAnnotations) {
				subT.Fatal(cmp.Diff(existing.ObjectMeta.Annotations, tc.expectedNewAnnotations))
			}
		})
	}
}

func TestDeploymentPodContainerImageMutator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "No Image Update Required",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "test-image"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "test-image"},
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
			name: "Image Update Required",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "test-image-desired"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "test-image-existing"},
								},
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeploymentPodContainerImageMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentPodContainerImageMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentPodContainerImageMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeploymentPodInitContainerImageMutator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "No Image Update Required",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "test-image"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{Image: "test-image"},
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
			name: "Image Update Required",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{
									{Image: "test-image-desired"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{
									{Image: "test-image-existing"},
								},
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeploymentPodInitContainerImageMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentPodInitContainerImageMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentPodInitContainerImageMutator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeploymentVolumeMuator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "VolumesMutator - no change",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
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
			name: "VolumesMutator - change - single volume",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol2"},
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
			name: "VolumesMutator - additional new volume",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
									{Name: "vol2"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
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
			name: "VolumesMutator - multiple volumes change",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
									{Name: "vol2"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
									{Name: "vol3"},
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
			name: "VolumesMutator - remove volumes",
			args: args{
				desired: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
									{Name: "vol2"},
								},
							},
						},
					},
				},
				existing: &k8sappsv1.Deployment{
					Spec: k8sappsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{Name: "vol1"},
									{Name: "vol2"},
									{Name: "vol3"},
								},
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeploymentVolumesMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentVolumesMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentVolumesMutator() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(tt.args.existing, tt.args.desired) {
				t.Errorf("DeploymentVolumesMutator() = %v\n, want %v", tt.args.existing, tt.args.desired)
			}
		})
	}
}

func TestInitContainerDeploymentVolumeMountsMuator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
	}

	deploymentFn := func(containers []corev1.Container) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: containers,
					},
				},
			},
		}
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "InitContainerVolumesMountsMutator - no change",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "InitContainerVolumesMountMutator - single container single volume mounts change",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol2", MountPath: "/newpath"}}},
				}),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "InitContainerVolumeMountMutator - no change single container multiple volume mounts",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{
						{Name: "vol1", MountPath: "/newpath"},
						{Name: "vol2", MountPath: "/newpath"},
					}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{
						{Name: "vol1", MountPath: "/newpath"},
						{Name: "vol2", MountPath: "/newpath"},
					}},
				}),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "InitContainerVolumeMountsMutator - single container multiple volume mounts change",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{
						{Name: "vol1", MountPath: "/newpath"},
						{Name: "vol3", MountPath: "/newpath"},
					}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{
						{Name: "vol1", MountPath: "/newpath"},
						{Name: "vol2", MountPath: "/newpath"},
					}},
				}),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "InitContainerVolumeMountsMutator - no change multiple containers single volume mounts",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "InitContainerVolumeMountsMutator - multiple containers single volume mounts change",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol2", MountPath: "/newpath"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3", MountPath: "/newpath"}}},
				}),
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeploymentInitContainerVolumeMountsMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentInitContainerVolumeMountsMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentInitContainerVolumeMountsMutator() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(tt.args.existing, tt.args.desired) {
				t.Error("DeploymentInitContainterVolumeMountsMutator: ", cmp.Diff(tt.args.existing, tt.args.desired))
			}
		})
	}
}

func TestContainerDeploymentVolumeMountsMuator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
	}

	deploymentFn := func(containers []corev1.Container) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: containers,
					},
				},
			},
		}
	}

	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "VolumeMountsMutator - no change",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "VolumeMountsMutator - single container single volume mounts change",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol2", MountPath: "/newpath"}}},
				}),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "VolumesMutator - no change single container multiple volume mounts",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{
						{Name: "vol1", MountPath: "/newpath"},
						{Name: "vol2", MountPath: "/newpath"},
					}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{
						{Name: "vol1", MountPath: "/newpath"},
						{Name: "vol2", MountPath: "/newpath"},
					}},
				}),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "VolumesMutator - single container multiple volume mounts change",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{
						{Name: "vol1", MountPath: "/newpath"},
						{Name: "vol3", MountPath: "/newpath"},
					}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{
						{Name: "vol1", MountPath: "/newpath"},
						{Name: "vol2", MountPath: "/newpath"},
					}},
				}),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "VolumesMutator - no change multiple containers single volume mounts",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
				}),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "VolumesMutator - multiple containers single volume mounts change",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol2", MountPath: "/newpath"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/newpath"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3", MountPath: "/newpath"}}},
				}),
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeploymentContainerVolumeMountsMutator(tt.args.desired, tt.args.existing)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeploymentContainerVolumeMountsMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeploymentContainerVolumeMountsMutator() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(tt.args.existing, tt.args.desired) {
				t.Error("DeploymentContainerVolumeMountsMutator: ", cmp.Diff(tt.args.existing, tt.args.desired))
			}
		})
	}
}

func TestWeakDeploymentVolumeMuator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
		volName  []string
	}

	deploymentFn := func(volumes []corev1.Volume) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Volumes: volumes,
					},
				},
			},
		}
	}

	tests := []struct {
		name     string
		args     args
		expected *k8sappsv1.Deployment
		want     bool
		wantErr  bool
	}{
		{
			name: "no change",
			args: args{
				desired:  deploymentFn([]corev1.Volume{{Name: "vol1"}}),
				existing: deploymentFn([]corev1.Volume{{Name: "vol1"}}),
				volName:  []string{"vol1"},
			},
			expected: deploymentFn([]corev1.Volume{{Name: "vol1"}}),
			want:     false,
			wantErr:  false,
		},
		{
			name: "add single volume",
			args: args{
				desired:  deploymentFn([]corev1.Volume{{Name: "vol2"}}),
				existing: deploymentFn([]corev1.Volume{{Name: "vol1"}}),
				volName:  []string{"vol2"},
			},
			expected: deploymentFn([]corev1.Volume{{Name: "vol1"}, {Name: "vol2"}}),
			want:     true,
			wantErr:  false,
		},
		{
			name: "add multiple volumes",
			args: args{
				desired:  deploymentFn([]corev1.Volume{{Name: "vol2"}, {Name: "vol3"}}),
				existing: deploymentFn([]corev1.Volume{{Name: "vol1"}}),
				volName:  []string{"vol2", "vol3"},
			},
			expected: deploymentFn([]corev1.Volume{{Name: "vol1"}, {Name: "vol2"}, {Name: "vol3"}}),
			want:     true,
			wantErr:  false,
		},
		{
			name: "update single volume",
			args: args{
				desired: deploymentFn([]corev1.Volume{{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secretB",
							},
						},
					},
				}}),
				existing: deploymentFn([]corev1.Volume{{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secretA",
							},
						},
					},
				}}),
				volName: []string{"vol1"},
			},
			expected: deploymentFn([]corev1.Volume{{
				Name: "vol1",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "secretB",
						},
					},
				},
			}}),
			want:    true,
			wantErr: false,
		},
		{
			name: "update multiple volumes",
			args: args{
				desired: deploymentFn([]corev1.Volume{
					{
						Name: "vol1",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secretC",
								},
							},
						},
					},
					{
						Name: "vol2",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secretD",
								},
							},
						},
					},
				}),
				existing: deploymentFn([]corev1.Volume{
					{
						Name: "vol1",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secretA",
								},
							},
						},
					},
					{
						Name: "vol2",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "secretB",
								},
							},
						},
					},
				}),
				volName: []string{"vol1", "vol2"},
			},
			expected: deploymentFn([]corev1.Volume{
				{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secretC",
							},
						},
					},
				},
				{
					Name: "vol2",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "secretD",
							},
						},
					},
				},
			}),
			want:    true,
			wantErr: false,
		},
		{
			name: "remove single volume",
			args: args{
				desired:  deploymentFn([]corev1.Volume{}),
				existing: deploymentFn([]corev1.Volume{{Name: "vol1"}, {Name: "vol2"}}),
				volName:  []string{"vol2"},
			},
			expected: deploymentFn([]corev1.Volume{{Name: "vol1"}}),
			want:     true,
			wantErr:  false,
		},
		{
			name: "remove multile volumes and existing volume",
			args: args{
				desired:  deploymentFn([]corev1.Volume{}),
				existing: deploymentFn([]corev1.Volume{{Name: "vol1"}, {Name: "vol2"}, {Name: "vol3"}}),
				volName:  []string{"vol2", "vol3"},
			},
			expected: deploymentFn([]corev1.Volume{{Name: "vol1"}}),
			want:     true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WeakDeploymentVolumesMutator(tt.args.desired, tt.args.existing, tt.args.volName)
			if (err != nil) != tt.wantErr {
				t.Errorf("WeakDeploymentVolumesMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WeakDeploymentVolumesMutator() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(tt.args.existing, tt.expected) {
				t.Fatal(cmp.Diff(tt.args.existing, tt.expected, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}

func TestWeakDeploymentInitContainerVolumeMountsMuator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
		volName  []string
	}

	deploymentFn := func(containers []corev1.Container) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						InitContainers: containers,
					},
				},
			},
		}
	}

	tests := []struct {
		name     string
		args     args
		expected *k8sappsv1.Deployment
		want     bool
		wantErr  bool
	}{
		{
			name: "no change",
			args: args{
				desired:  deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
				existing: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
				volName:  []string{"vol1"},
			},
			expected: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
			want:     false,
			wantErr:  false,
		},
		{
			name: "single container add volume mounts",
			args: args{
				desired:  deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}, {Name: "vol3"}}}}),
				existing: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
				volName:  []string{"vol2", "vol3"},
			},
			expected: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{
				{Name: "vol1"},
				{Name: "vol2"},
				{Name: "vol3"},
			}}}),
			want:    true,
			wantErr: false,
		},
		{
			name: "single container should not add volume mounts that not in the list",
			args: args{
				desired:  deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}, {Name: "vol3"}}}}),
				existing: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
				volName:  []string{"vol2"},
			},
			expected: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{
				{Name: "vol1"},
				{Name: "vol2"},
			}}}),
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple containers add volume mounts",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3"}, {Name: "vol4"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol5"}, {Name: "vol6"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}}},
				}),
				volName: []string{"vol3", "vol4", "vol5", "vol6"},
			},
			expected: deploymentFn([]corev1.Container{
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}, {Name: "vol3"}, {Name: "vol4"}}},
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}, {Name: "vol5"}, {Name: "vol6"}}},
			}),
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple containers should not add volume mounts that not in the list",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3"}, {Name: "vol4"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol5"}, {Name: "vol6"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}}},
				}),
				volName: []string{"vol3", "vol5"},
			},
			expected: deploymentFn([]corev1.Container{
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}, {Name: "vol3"}}},
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}, {Name: "vol5"}}},
			}),
			want:    true,
			wantErr: false,
		},
		{
			name: "single container update volume mounts",
			args: args{
				desired:  deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path2"}, {Name: "vol2"}}}}),
				existing: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path1"}}}}),
				volName:  []string{"vol1", "vol2"},
			},
			expected: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{
				{Name: "vol1", MountPath: "/path2"},
				{Name: "vol2"},
			}}}),
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple container update volume mounts",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path2"}, {Name: "vol2"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3", MountPath: "/path4"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path1"}, {Name: "vol2"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3", MountPath: "/path3"}}},
				}),
				volName: []string{"vol1", "vol2", "vol3"},
			},
			expected: deploymentFn([]corev1.Container{
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path2"}, {Name: "vol2"}}},
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol3", MountPath: "/path4"}}},
			}),
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WeakDeploymentInitContainerVolumeMountsMutator(tt.args.desired, tt.args.existing, tt.args.volName)
			if (err != nil) != tt.wantErr {
				t.Errorf("WeakDeploymentInitContainerVolumeMountsMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WeakDeploymentInitContainerVolumeMountsMutator() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(tt.args.existing, tt.expected) {
				t.Fatal(cmp.Diff(tt.args.existing, tt.expected, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}

func TestWeakDeploymentContainerVolumeMountsMuator(t *testing.T) {
	type args struct {
		desired  *k8sappsv1.Deployment
		existing *k8sappsv1.Deployment
		volName  []string
	}

	deploymentFn := func(containers []corev1.Container) *k8sappsv1.Deployment {
		return &k8sappsv1.Deployment{
			Spec: k8sappsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: containers,
					},
				},
			},
		}
	}

	tests := []struct {
		name     string
		args     args
		expected *k8sappsv1.Deployment
		want     bool
		wantErr  bool
	}{
		{
			name: "no change",
			args: args{
				desired:  deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
				existing: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
				volName:  []string{"vol1"},
			},
			expected: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
			want:     false,
			wantErr:  false,
		},
		{
			name: "single container add volume mounts",
			args: args{
				desired:  deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}, {Name: "vol3"}}}}),
				existing: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
				volName:  []string{"vol2", "vol3"},
			},
			expected: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{
				{Name: "vol1"},
				{Name: "vol2"},
				{Name: "vol3"},
			}}}),
			want:    true,
			wantErr: false,
		},
		{
			name: "single container should not add volume mounts that not in the list",
			args: args{
				desired:  deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}, {Name: "vol3"}}}}),
				existing: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}}}),
				volName:  []string{"vol2"},
			},
			expected: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{
				{Name: "vol1"},
				{Name: "vol2"},
			}}}),
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple containers add volume mounts",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3"}, {Name: "vol4"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol5"}, {Name: "vol6"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}}},
				}),
				volName: []string{"vol3", "vol4", "vol5", "vol6"},
			},
			expected: deploymentFn([]corev1.Container{
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}, {Name: "vol3"}, {Name: "vol4"}}},
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}, {Name: "vol5"}, {Name: "vol6"}}},
			}),
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple containers should not add volume mounts that not in the list",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3"}, {Name: "vol4"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol5"}, {Name: "vol6"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}}},
				}),
				volName: []string{"vol3", "vol5"},
			},
			expected: deploymentFn([]corev1.Container{
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol1"}, {Name: "vol3"}}},
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol2"}, {Name: "vol5"}}},
			}),
			want:    true,
			wantErr: false,
		},
		{
			name: "single container update volume mounts",
			args: args{
				desired:  deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path2"}, {Name: "vol2"}}}}),
				existing: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path1"}}}}),
				volName:  []string{"vol1", "vol2"},
			},
			expected: deploymentFn([]corev1.Container{{VolumeMounts: []corev1.VolumeMount{
				{Name: "vol1", MountPath: "/path2"},
				{Name: "vol2"},
			}}}),
			want:    true,
			wantErr: false,
		},
		{
			name: "multiple container update volume mounts",
			args: args{
				desired: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path2"}, {Name: "vol2"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3", MountPath: "/path4"}}},
				}),
				existing: deploymentFn([]corev1.Container{
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path1"}, {Name: "vol2"}}},
					{VolumeMounts: []corev1.VolumeMount{{Name: "vol3", MountPath: "/path3"}}},
				}),
				volName: []string{"vol1", "vol2", "vol3"},
			},
			expected: deploymentFn([]corev1.Container{
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol1", MountPath: "/path2"}, {Name: "vol2"}}},
				{VolumeMounts: []corev1.VolumeMount{{Name: "vol3", MountPath: "/path4"}}},
			}),
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WeakDeploymentContainerVolumeMountsMutator(tt.args.desired, tt.args.existing, tt.args.volName)
			if (err != nil) != tt.wantErr {
				t.Errorf("WeakDeploymentContainerVolumesMutator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WeakDeploymentContainerVolumesMutator() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(tt.args.existing, tt.expected) {
				t.Fatal(cmp.Diff(tt.args.existing, tt.expected, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}
