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
