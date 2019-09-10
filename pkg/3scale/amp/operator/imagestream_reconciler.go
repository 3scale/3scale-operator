package operator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/3scale/3scale-operator/pkg/helper"
	imagev1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type ImageStreamReconciler interface {
	IsUpdateNeeded(desired, existing *imagev1.ImageStream) bool
}

type ImageStreamBaseReconciler struct {
	BaseAPIManagerLogicReconciler
	reconciler ImageStreamReconciler
}

func NewImageStreamBaseReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler, reconciler ImageStreamReconciler) *ImageStreamBaseReconciler {
	return &ImageStreamBaseReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
		reconciler:                    reconciler,
	}
}

func (r *ImageStreamBaseReconciler) Reconcile(desired *imagev1.ImageStream) error {
	objectInfo := ObjectInfo(desired)
	existing := &imagev1.ImageStream{}
	err := r.Client().Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: r.apiManager.GetNamespace()},
		existing)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.createResource(desired)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			return nil
		}
		return err
	}

	update, err := r.isUpdateNeeded(desired, existing)
	if err != nil {
		return err
	}

	if update {
		return r.updateResource(existing)
	}

	return nil
}

func (r *ImageStreamBaseReconciler) isUpdateNeeded(desired, existing *imagev1.ImageStream) (bool, error) {
	updated := helper.EnsureObjectMeta(&existing.ObjectMeta, &desired.ObjectMeta)

	updatedTmp, err := r.ensureOwnerReference(existing)
	if err != nil {
		return false, nil
	}
	updated = updated || updatedTmp

	updatedTmp = r.reconciler.IsUpdateNeeded(desired, existing)
	updated = updated || updatedTmp

	return updated, nil
}

type ImageStreamGenericReconciler struct {
}

func NewImageStreamGenericReconciler() *ImageStreamGenericReconciler {
	return &ImageStreamGenericReconciler{}
}

func (r *ImageStreamGenericReconciler) IsUpdateNeeded(desired, existing *imagev1.ImageStream) bool {
	// merging approach will be implemented
	// spec.tags tagrefences in desired must exist in existing.
	// If element does not exist, append
	// If exists (by name), ensure From field and ImportPolicy fields are equal.
	updated := false

	findTagReference := func(tagRefName string, tagRefS []imagev1.TagReference) int {
		for i, _ := range tagRefS {
			if tagRefS[i].Name == tagRefName {
				return i
			}
		}
		return -1
	}

	for idx, _ := range desired.Spec.Tags {
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

	return updated
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
