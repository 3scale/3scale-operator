package helper

import "github.com/3scale/3scale-operator/pkg/common"

// From
// https://github.com/openshift/library-go/blob/master/pkg/operator/resource/resourcemerge/object_merger.go

// EnsureObjectMeta ensure Labels, Annotations
func EnsureObjectMeta(existing, desired common.KubernetesObject) bool {
	updated := false

	existingLabels := existing.GetLabels()
	existingAnnotations := existing.GetAnnotations()

	MergeMapStringString(&updated, &existingLabels, desired.GetLabels())
	MergeMapStringString(&updated, &existingAnnotations, desired.GetAnnotations())

	existing.SetLabels(existingLabels)
	existing.SetAnnotations(existingAnnotations)

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
