package component

import v1 "k8s.io/api/core/v1"

func envVarFromConfigMap(envVarName string, configMapName, configMapKey string) v1.EnvVar {
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

func envVarFromConfigMapOptional(envVarName string, configMapName, configMapKey string) v1.EnvVar {
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

func envVarFromValue(name string, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func envVarFromSecret(envVarName string, secretName, secretKey string) v1.EnvVar {
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

func envVarFromSecretOptional(envVarName string, secretName, secretKey string) v1.EnvVar {
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
