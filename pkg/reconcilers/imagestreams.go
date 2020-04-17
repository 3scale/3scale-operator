package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"
	imagev1 "github.com/openshift/api/image/v1"
)

func GenericImagestreamMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*imagev1.ImageStream)
	if !ok {
		return false, fmt.Errorf("%T is not a *imagev1.ImageStream", existingObj)
	}
	desired, ok := desiredObj.(*imagev1.ImageStream)
	if !ok {
		return false, fmt.Errorf("%T is not a *imagev1.ImageStream", desiredObj)
	}

	// merging approach will be implemented
	// spec.tags tagrefences in desired must exist in existing.
	// If element does not exist, append
	// If exists (by name), ensure From field and ImportPolicy fields are equal.
	updated := false

	findTagReference := func(tagRefName string, tagRefS []imagev1.TagReference) int {
		for i := range tagRefS {
			if tagRefS[i].Name == tagRefName {
				return i
			}
		}
		return -1
	}

	for idx := range desired.Spec.Tags {
		if existingIdx := findTagReference(desired.Spec.Tags[idx].Name, existing.Spec.Tags); existingIdx < 0 {
			// does not exist, append
			existing.Spec.Tags = append(existing.Spec.Tags, desired.Spec.Tags[idx])
			updated = true
		} else {
			// exists, reconcile
			tmpUpdated := imageStreamReconcile(existing.Spec.Tags, existingIdx, desired.Spec.Tags, idx)
			updated = updated || tmpUpdated
		}
	}

	return updated, nil
}

func imageStreamReconcile(existingTags []imagev1.TagReference, existingIdx int, desiredTags []imagev1.TagReference, desiredIdx int) bool {
	// From and ImportPolicy fields are equal.
	updated := false

	if !reflect.DeepEqual(existingTags[existingIdx].From, desiredTags[desiredIdx].From) {
		existingTags[existingIdx].From = desiredTags[desiredIdx].From
		updated = true
	}

	if !reflect.DeepEqual(existingTags[existingIdx].ImportPolicy, desiredTags[desiredIdx].ImportPolicy) {
		existingTags[existingIdx].ImportPolicy = desiredTags[desiredIdx].ImportPolicy
		updated = true
	}

	return updated
}
