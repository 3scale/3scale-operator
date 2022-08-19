package helper

import v1 "k8s.io/api/core/v1"

func FindContainerPortByName(ports []v1.ContainerPort, name string) (v1.ContainerPort, bool) {
	for idx := range ports {
		if ports[idx].Name == name {
			return ports[idx], true
		}
	}

	return v1.ContainerPort{}, false
}
