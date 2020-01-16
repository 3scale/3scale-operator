package helper

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetTestSecret(namespace, secretName string, data map[string]string) *v1.Secret {
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		StringData: data,
		Type:       v1.SecretTypeOpaque,
	}
	secret.Data = GetSecretDataFromStringData(secret.StringData)
	return secret
}
