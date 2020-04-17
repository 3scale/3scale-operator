package component

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HighAvailability struct {
	Options *HighAvailabilityOptions
}

var HighlyAvailableExternalDatabases = map[string]bool{
	"backend-redis": true,
	"system-redis":  true,
	"system-mysql":  true,
}

func NewHighAvailability(options *HighAvailabilityOptions) *HighAvailability {
	return &HighAvailability{Options: options}
}

func (ha *HighAvailability) SystemDatabaseSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: SystemSecretSystemDatabaseSecretName,
			Labels: map[string]string{
				"app":                  ha.Options.AppLabel,
				"threescale_component": "system",
			},
		},
		StringData: map[string]string{
			SystemSecretSystemDatabaseURLFieldName: ha.Options.SystemDatabaseURL,
		},
		Type: v1.SecretTypeOpaque,
	}
}
