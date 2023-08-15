package controllers

import (
	"fmt"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"reflect"
	"strings"
)

const (
	OpenApiSecretLabelPrefix = "secret.openapi.apps.3scale.net/"
	OpenApiSecretLabelValue  = "true"
)

func openApiSecretLabelKey(uid string) string {
	return fmt.Sprintf("%s%s", OpenApiSecretLabelPrefix, uid)
}

func openapiSecretLabelKey(uid string) string {
	return fmt.Sprintf("%s%s", OpenApiSecretLabelPrefix, uid)
}

func replaceOpenAPISecretLabels(openapi *capabilitiesv1beta1.OpenAPI, secretUID string) bool {
	existingLabels := openapi.GetLabels()
	if existingLabels == nil {
		existingLabels = map[string]string{}
	}

	existingSecretLabels := map[string]string{}
	// existing Secret UIDs not included in desiredAPIUIDs are deleted
	for k := range existingLabels {
		if strings.HasPrefix(k, OpenApiSecretLabelPrefix) {
			existingSecretLabels[k] = OpenApiSecretLabelValue
			// it is safe to remove keys while looping in range
			delete(existingLabels, k)
		}
	}

	desiredSecretLabels := map[string]string{}
	desiredSecretLabels[openapiSecretLabelKey(secretUID)] = OpenApiSecretLabelValue
	existingLabels[openapiSecretLabelKey(secretUID)] = OpenApiSecretLabelValue

	openapi.SetLabels(existingLabels)

	return !reflect.DeepEqual(existingSecretLabels, desiredSecretLabels)
}
