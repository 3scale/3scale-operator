package operator

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	imagev1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator/resourcemerge"

	v1 "k8s.io/api/core/v1"
)

type AMPImagesReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &AMPImagesReconciler{}

func NewAMPImagesReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) AMPImagesReconciler {
	return AMPImagesReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *AMPImagesReconciler) Reconcile() (reconcile.Result, error) {
	ampImages, err := r.ampImages()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileBackendImageStream(ampImages.BackendImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileZyncImageStream(ampImages.ZyncImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileApicastImageStream(ampImages.APICastImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileSystemImageStream(ampImages.SystemImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileZyncDatabasePostgreSQLImageStream(ampImages.ZyncDatabasePostgreSQLImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileBackendRedisImageStream(ampImages.BackendRedisImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileSystemRedisImageStream(ampImages.SystemRedisImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileSystemMemcachedImageStream(ampImages.SystemMemcachedImageStream())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileDeploymentsServiceAccount(ampImages.DeploymentsServiceAccount())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// TODO should this be performed in another place
func (r *AMPImagesReconciler) ampImages() (*component.AmpImages, error) {
	optsProvider := OperatorAmpImagesOptionsProvider{APIManagerSpec: &r.apiManager.Spec}
	opts, err := optsProvider.GetAmpImagesOptions()
	if err != nil {
		return nil, err
	}
	return component.NewAmpImages(opts), nil
}

func (r *AMPImagesReconciler) reconcileImageStream(desiredImageStream *imagev1.ImageStream) error {
	desiredCopy := desiredImageStream.DeepCopy()
	err := r.InitializeAsAPIManagerObject(desiredCopy)
	if err != nil {
		return err
	}

	objectInfo := r.ObjectInfo(desiredCopy)
	existingImageStream := &imagev1.ImageStream{}
	err = r.Client().Get(context.TODO(), r.NamespacedNameWithAPIManagerNamespace(desiredCopy), existingImageStream)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredCopy)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureImageStream(existingImageStream, desiredCopy)
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

func (r *AMPImagesReconciler) ensureImageStream(updated, desired *imagev1.ImageStream) (bool, error) {
	changed := false

	objectMetaChanged, err := resourcemerge.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta, r.apiManager, r.Scheme())
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
func (r *AMPImagesReconciler) ensureImageTagReferences(updated, desired *imagev1.ImageStream) {
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

func (r *AMPImagesReconciler) ObjectInfo(obj common.KubernetesObject) string {
	return fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
}

func (r *AMPImagesReconciler) reconcileImageStreamTagReferencesGeneration(existingImageStream, desiredImageStream *imagev1.ImageStream) {
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

func (r *AMPImagesReconciler) reconcileBackendImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileZyncImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileApicastImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileSystemImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileZyncDatabasePostgreSQLImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileBackendRedisImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileSystemRedisImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileSystemMemcachedImageStream(desiredImageStream *imagev1.ImageStream) error {
	return r.reconcileImageStream(desiredImageStream)
}

func (r *AMPImagesReconciler) reconcileDeploymentsServiceAccount(desiredServiceAccount *v1.ServiceAccount) error {
	err := r.InitializeAsAPIManagerObject(desiredServiceAccount)
	if err != nil {
		return err
	}

	objectInfo := r.ObjectInfo(desiredServiceAccount)
	existingServiceAccount := &v1.ServiceAccount{}
	err = r.Client().Get(context.TODO(), r.NamespacedNameWithAPIManagerNamespace(desiredServiceAccount), existingServiceAccount)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredServiceAccount)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureDeploymentsServiceAccount(existingServiceAccount, desiredServiceAccount)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating ServiceAccount %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingServiceAccount)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating ServiceAccount %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *AMPImagesReconciler) ensureDeploymentsServiceAccount(updated, desired *v1.ServiceAccount) (bool, error) {
	changed := false

	objectMetaChanged, err := resourcemerge.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta, r.apiManager, r.Scheme())
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// We only reconcile ImagePullSecrets
	r.ensureServiceAccountImagePullSecrets(updated, desired)
	if !reflect.DeepEqual(updated.ImagePullSecrets, desired.ImagePullSecrets) {
		updated.ImagePullSecrets = desired.ImagePullSecrets
		changed = true
	}

	return changed, nil
}

// Merges existing serviceaccounts pullsecrets into the desired serviceaccounts
// This is because OpenShift creates additional ImagePullSecrets and we
// don't want to lose them
// TODO would it be better to just update the "updated" variable and just
// communicate that is has changed directly? This
func (r *AMPImagesReconciler) ensureServiceAccountImagePullSecrets(updated, desired *v1.ServiceAccount) {
	desiredServiceAccountImagePullSecretsMap := map[string]*v1.LocalObjectReference{}
	for idx := range desired.ImagePullSecrets {
		imagePullSecret := &desired.ImagePullSecrets[idx]
		desiredServiceAccountImagePullSecretsMap[imagePullSecret.Name] = imagePullSecret
	}

	newDesiredImagePullSecrets := []v1.LocalObjectReference{}
	for _, val := range desired.ImagePullSecrets {
		newDesiredImagePullSecrets = append(newDesiredImagePullSecrets, val)
	}

	for idx := range updated.ImagePullSecrets {
		updatedImagePullSecret := &updated.ImagePullSecrets[idx]
		if _, ok := desiredServiceAccountImagePullSecretsMap[updatedImagePullSecret.Name]; !ok {
			desiredServiceAccountImagePullSecretsMap[updatedImagePullSecret.Name] = updatedImagePullSecret
			newDesiredImagePullSecrets = append(newDesiredImagePullSecrets, *updatedImagePullSecret)
		}
	}

	desired.ImagePullSecrets = newDesiredImagePullSecrets

	sort.Slice(updated.ImagePullSecrets, func(i, j int) bool { return updated.ImagePullSecrets[i].Name < updated.ImagePullSecrets[j].Name })
	sort.Slice(desired.ImagePullSecrets, func(i, j int) bool { return desired.ImagePullSecrets[i].Name < desired.ImagePullSecrets[j].Name })
}
