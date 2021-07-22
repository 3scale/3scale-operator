package helper

import (
	v1 "k8s.io/api/core/v1"
)

// FindVolumeMountByMountPath returns the smallest index i at which x.MountPath == a[i].MountPath,
// or -1 if there is no such index.
func FindVolumeMountByMountPath(a []v1.VolumeMount, x v1.VolumeMount) int {
	for i, n := range a {
		if n.MountPath == x.MountPath {
			return i
		}
	}
	return -1
}

// FindVolumeMountByMountPath returns the smallest index i at which x.MountPath == a[i].MountPath,
// or -1 if there is no such index.
func FindVolumeMountByName(a []v1.VolumeMount, name string) int {
	for i, n := range a {
		if n.Name == name {
			return i
		}
	}
	return -1
}
