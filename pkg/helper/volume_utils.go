package helper

import (
	v1 "k8s.io/api/core/v1"
)

// FindVolumeByName returns the smallest index i at which x.Name == a[i].Name,
// or -1 if there is no such index.
func FindVolumeByName(a []v1.Volume, name string) int {
	for i, n := range a {
		if n.Name == name {
			return i
		}
	}
	return -1
}

func VolumeFromSecretEqual(a v1.Volume, b v1.Volume) bool {
	if a.Secret == nil || b.Secret == nil {
		return false
	}

	return a.Name == b.Name && a.Secret.SecretName == b.Secret.SecretName
}
