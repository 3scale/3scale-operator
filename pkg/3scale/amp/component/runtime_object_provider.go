package component

import "k8s.io/api/core/v1"

func createEnvVarFromConfigMap(envVarName string, configMapName, configMapKey string) v1.EnvVar {
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

func createEnvVarFromValue(name string, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func createEnvvarFromSecret(envVarName string, secretName, secretKey string) v1.EnvVar {
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
