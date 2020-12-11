package reconcilers

import (
	"k8s.io/api/core/v1"
)

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
