package operator

import (
	"fmt"
	"reflect"
	"strings"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
)

const (
	APIManagerSecretLabelPrefix = "secret.apimanager.apps.3scale.net/"
)

func apimanagerSecretLabelKey(uid string) string {
	return fmt.Sprintf("%s%s", APIManagerSecretLabelPrefix, uid)
}

func replaceAPIManagerSecretLabels(apimanager *appsv1alpha1.APIManager, desiredSecretUIDs map[string]string) bool {
	existingLabels := apimanager.GetLabels()

	if existingLabels == nil {
		existingLabels = map[string]string{}
	}

	existingSecretLabels := map[string]string{}

	// existing Secret UIDs not included in desiredAPIUIDs are deleted
	for key, value := range existingLabels {
		if strings.HasPrefix(key, APIManagerSecretLabelPrefix) {
			existingSecretLabels[key] = value
			// it is safe to remove keys while looping in range
			delete(existingLabels, key)
		}
	}

	desiredSecretLabels := map[string]string{}
	for uid, watchedByStatus := range desiredSecretUIDs {
		desiredSecretLabels[apimanagerSecretLabelKey(uid)] = watchedByStatus
		existingLabels[apimanagerSecretLabelKey(uid)] = watchedByStatus
	}

	apimanager.SetLabels(existingLabels)

	return !reflect.DeepEqual(existingSecretLabels, desiredSecretLabels)
}
