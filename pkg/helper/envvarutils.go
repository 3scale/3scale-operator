package helper

import (
	"reflect"

	v1 "k8s.io/api/core/v1"
)

func EnvVarFromConfigMap(envVarName string, configMapName, configMapKey string) v1.EnvVar {
	return v1.EnvVar{
		Name: envVarName,
		ValueFrom: &v1.EnvVarSource{
			ConfigMapKeyRef: &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMapName,
				},
				Key: configMapKey,
			},
		},
	}
}

func EnvVarFromConfigMapOptional(envVarName string, configMapName, configMapKey string) v1.EnvVar {
	trueValue := true
	return v1.EnvVar{
		Name: envVarName,
		ValueFrom: &v1.EnvVarSource{
			ConfigMapKeyRef: &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configMapName,
				},
				Key:      configMapKey,
				Optional: &trueValue,
			},
		},
	}
}

func EnvVarFromValue(name string, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func EnvVarFromSecret(envVarName string, secretName, secretKey string) v1.EnvVar {
	return v1.EnvVar{
		Name: envVarName,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
}

func EnvVarFromSecretOptional(envVarName string, secretName, secretKey string) v1.EnvVar {
	trueValue := true
	return v1.EnvVar{
		Name: envVarName,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key:      secretKey,
				Optional: &trueValue,
			},
		},
	}
}

// FindEnvVar returns the smallest index i at which x == a[i],
// or -1 if there is no such index.
func FindEnvVar(a []v1.EnvVar, x string) int {
	for i, n := range a {
		if n.Name == x {
			return i
		}
	}
	return -1
}

// EnsureEnvVar updates existingEnvVars with desired, reconciling any
// possible differences. It returns whether existingEnvVars has been modified.
func EnsureEnvVar(desired v1.EnvVar, existingEnvVars *[]v1.EnvVar) bool {
	update := false
	envVarExists := false
	for idx := range *existingEnvVars {
		if (*existingEnvVars)[idx].Name == desired.Name {
			envVarExists = true
			if !reflect.DeepEqual((*existingEnvVars)[idx], desired) {
				(*existingEnvVars)[idx] = desired
				update = true
			}
			break
		}
	}

	if !envVarExists {
		*existingEnvVars = append(*existingEnvVars, desired)
		update = true
	}

	return update
}

// RemoveDuplicateEnvVars removes duplicate env vars by name from a slice
func RemoveDuplicateEnvVars(envVars []v1.EnvVar) []v1.EnvVar {
	set := map[string]v1.EnvVar{}
	var result []v1.EnvVar

	for idx := range envVars {
		if _, ok := set[envVars[idx].Name]; !ok {
			set[envVars[idx].Name] = envVars[idx]
			result = append(result, envVars[idx])
		}
	}

	return result
}

// EnvVarReconciler implements basic env var reconcilliation.
// Added when in desired and not in existing
// Updated when in desired and in existing but not equal
// Removed when not in desired and exists in existing DC
func EnvVarReconciler(desired []v1.EnvVar, existing *[]v1.EnvVar, envVar string) bool {
	update := false

	if existing == nil {
		*existing = make([]v1.EnvVar, 0)
	}

	desiredIdx := FindEnvVar(desired, envVar)
	existingIdx := FindEnvVar(*existing, envVar)

	if desiredIdx < 0 && existingIdx >= 0 {
		// env var exists in existing and does not exist in desired => Remove from the list
		// shift all of the elements at the right of the deleting index by one to the left
		*existing = append((*existing)[:existingIdx], (*existing)[existingIdx+1:]...)
		update = true
	} else if desiredIdx < 0 && existingIdx < 0 {
		// env var does not exist in existing and does not exist in desired => NOOP
	} else if desiredIdx >= 0 && existingIdx < 0 {
		// env var does not exist in existing and exists in desired => ADD it
		*existing = append(*existing, desired[desiredIdx])
		update = true
	} else {
		// env var exists in existing and exists in desired
		if !reflect.DeepEqual((*existing)[existingIdx], desired[desiredIdx]) {
			(*existing)[existingIdx] = desired[desiredIdx]
			update = true
		}
	}
	return update
}
