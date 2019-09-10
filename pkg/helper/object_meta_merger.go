package helper

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// From
// https://github.com/openshift/library-go/blob/master/pkg/operator/resource/resourcemerge/object_merger.go

// EnsureObjectMeta ensure Labels, Annotations
func EnsureObjectMeta(existing, desired *metav1.ObjectMeta) bool {
	updated := false

	MergeMapStringString(&updated, &existing.Labels, desired.Labels)
	MergeMapStringString(&updated, &existing.Annotations, desired.Annotations)

	return updated
}

func EnsureString(modified *bool, existing *string, required string) {
	if required != *existing {
		*existing = required
		*modified = true
	}
}

func MergeMapStringString(modified *bool, existing *map[string]string, desired map[string]string) {
	if *existing == nil {
		*existing = map[string]string{}
	}

	for k, v := range desired {
		if existingVal, ok := (*existing)[k]; !ok || v != existingVal {
			(*existing)[k] = v
			*modified = true
		}
	}
}
