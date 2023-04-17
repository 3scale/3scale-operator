package helper

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
)

func TestEnvVarReconciler(t *testing.T) {
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

	envVarA := []corev1.EnvVar{
		{
			Name:  "A",
			Value: "valueA",
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
		{"EmptyExisting", nil, sliceCopy(envVarA), true, sliceCopy(envVarA)},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := sliceCopy(tc.existingEnvVar)
			update := EnvVarReconciler(tc.desiredEnvVar, &existing, "A")
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing, tc.expectedNewEnvVar) {
				subT.Fatal(cmp.Diff(existing, tc.expectedNewEnvVar))
			}
		})
	}
}
