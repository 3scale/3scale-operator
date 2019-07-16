package operator

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type ServiceAccountReconciler struct {
	BaseReconciler
	ObjectMetaMerger ObjectMetaMerger
}

func NewServiceAccountReconciler(baseReconciler BaseReconciler, objectMetaMerger ObjectMetaMerger) ServiceAccountReconciler {
	return ServiceAccountReconciler{
		BaseReconciler:   baseReconciler,
		ObjectMetaMerger: objectMetaMerger,
	}
}

func (r *ServiceAccountReconciler) Reconcile(desiredServiceAccount *v1.ServiceAccount) error {
	objectInfo := ObjectInfo(desiredServiceAccount)
	existingServiceAccount := &v1.ServiceAccount{}
	err := r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredServiceAccount.Name, Namespace: desiredServiceAccount.Namespace}, existingServiceAccount)
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

func (r *ServiceAccountReconciler) ensureDeploymentsServiceAccount(updated, desired *v1.ServiceAccount) (bool, error) {
	changed := false

	objectMetaChanged, err := r.ObjectMetaMerger.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta)
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
func (r *ServiceAccountReconciler) ensureServiceAccountImagePullSecrets(updated, desired *v1.ServiceAccount) {
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
