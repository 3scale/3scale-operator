package operator

import (
	"fmt"
	"reflect"
	"strings"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
)

const (
	APIManagerSecretLabelPrefix = "secret.apimanager.apps.3scale.net/"
	APIManagerSecretLabelValue  = "true"
)

func apimanagerSecretLabelKey(uid string) string {
	return fmt.Sprintf("%s%s", APIManagerSecretLabelPrefix, uid)
}

func replaceAPIManagerSecretLabels(apimanager *appsv1alpha1.APIManager, desiredSecretUIDs []string) bool {

	existingLabels := apimanager.GetLabels()

	if existingLabels == nil {
		existingLabels = map[string]string{}
	}

	existingSecretLabels := map[string]string{}

	// existing Secret UIDs not included in desiredAPIUIDs are deleted
	for k := range existingLabels {
		if strings.HasPrefix(k, APIManagerSecretLabelPrefix) {
			existingSecretLabels[k] = APIManagerSecretLabelValue
			// it is safe to remove keys while looping in range
			delete(existingLabels, k)
		}
	}

	desiredSecretLabels := map[string]string{}
	for _, uid := range desiredSecretUIDs {
		desiredSecretLabels[apimanagerSecretLabelKey(uid)] = APIManagerSecretLabelValue
		existingLabels[apimanagerSecretLabelKey(uid)] = APIManagerSecretLabelValue
	}

	apimanager.SetLabels(existingLabels)

	return !reflect.DeepEqual(existingSecretLabels, desiredSecretLabels)
}
