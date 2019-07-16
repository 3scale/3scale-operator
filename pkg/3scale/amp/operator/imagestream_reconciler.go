package operator

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	imagev1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type ImageStreamReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewImageStreamReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) ImageStreamReconciler {
	return ImageStreamReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *ImageStreamReconciler) Reconcile(desiredImageStream *imagev1.ImageStream) error {
	objectInfo := ObjectInfo(desiredImageStream)
	existingImageStream := &imagev1.ImageStream{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredImageStream.Name, Namespace: desiredImageStream.Namespace}, existingImageStream)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredImageStream)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureImageStream(existingImageStream, desiredImageStream)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating ImageStream %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingImageStream)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating ImageStream %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *ImageStreamReconciler) ensureImageStream(updated, desired *imagev1.ImageStream) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	r.ensureImageTagReferences(updated, desired)
	if !reflect.DeepEqual(updated.Spec, desired.Spec) {
		updated.Spec = desired.Spec
		changed = true
	}

	return changed, nil
}

// Sets the generation field of the desired ImageStreamTags with the value
// of the equivalent ImageStreamTag in the existing ImageStreamTag
// That's because the Generation field in the ImageStream TagReferences
// is filled by OpenShift after deploying so that would
// make the comparison on that field always with a result of being
// unequal when comparing the generation field between existing and
// desired
// It also sets the ReferencePolicyType in the desired in case it is empty
// because OpenShift fills it with a value when not defined
// The arrays are sorted because there could be the same tags
// but in different order and the comparison should be performed
// independently of the order of the arrays
func (r *ImageStreamReconciler) ensureImageTagReferences(updated, desired *imagev1.ImageStream) {
	updatedImageStreamTagReferenceMap := map[string]*imagev1.TagReference{}
	for idx := range updated.Spec.Tags {
		tagref := &updated.Spec.Tags[idx]
		updatedImageStreamTagReferenceMap[tagref.Name] = tagref
	}

	for idx := range desired.Spec.Tags {
		desiredTagRef := &desired.Spec.Tags[idx]
		if updatedTagRef, ok := updatedImageStreamTagReferenceMap[desiredTagRef.Name]; ok {
			desiredTagRef.Generation = updatedTagRef.Generation

			if desiredTagRef.ReferencePolicy.Type == "" {
				desiredTagRef.ReferencePolicy.Type = updatedTagRef.ReferencePolicy.Type
			}
		}
	}

	sort.Slice(updated.Spec.Tags, func(i, j int) bool { return updated.Spec.Tags[i].Name < updated.Spec.Tags[j].Name })
	sort.Slice(desired.Spec.Tags, func(i, j int) bool { return desired.Spec.Tags[i].Name < desired.Spec.Tags[j].Name })

	// if len(updated.Spec.Tags) != len(desired.Spec.Tags) {
	// 	updated.Spec.Tags = desired.Spec.Tags
	// 	return
	// }

	// for idx := range desired.Spec.Tags {
	// 	if desired.Spec.Tags[idx].Name != updated.Spec.Tags[idx].Name {
	// 		updated.Spec.Tags = desired.Spec.Tags
	// 		return
	// 	}
	// }

}

func (r *ImageStreamReconciler) reconcileImageStreamTagReferencesGeneration(existingImageStream, desiredImageStream *imagev1.ImageStream) {
	existingImageStreamTagReferenceMap := map[string]*imagev1.TagReference{}
	for idx := range existingImageStream.Spec.Tags {
		tagref := &existingImageStream.Spec.Tags[idx]
		existingImageStreamTagReferenceMap[tagref.Name] = tagref
	}

	for idx := range desiredImageStream.Spec.Tags {
		desiredTagRef := &desiredImageStream.Spec.Tags[idx]
		if existingTagRef, ok := existingImageStreamTagReferenceMap[desiredTagRef.Name]; ok {
			desiredTagRef.Generation = existingTagRef.Generation
		}
	}
}
