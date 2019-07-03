package template

import (
	templatev1 "github.com/openshift/api/template/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Basic3scaleTemplate creates base template for all 3scale templates
func Basic3scaleTemplate() *templatev1.Template {
	template := templatev1.Template{
		TypeMeta:   buildTemplateType(),
		ObjectMeta: buildTemplateMeta(),
		Message:    buildTemplateMessage(),
		Parameters: buildTemplateParameters(),
		Objects:    []runtime.RawExtension{},
	}
	return &template
}

func buildTemplateType() metav1.TypeMeta {
	typeMeta := metav1.TypeMeta{
		Kind:       "Template",
		APIVersion: "template.openshift.io/v1",
	}
	return typeMeta
}

// The 'metadata' part of a template
func buildTemplateMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        "3scale-api-management",
		Annotations: buildTemplateMetaAnnotations(),
	}
	return meta
}

func buildTemplateMetaAnnotations() map[string]string {
	annotations := map[string]string{
		"openshift.io/display-name":          "3scale API Management",
		"openshift.io/provider-display-name": "Red Hat, Inc.",
		"iconClass":                          "icon-3scale",
		"description":                        "3scale API Management main system",
		"tags":                               "integration, api management, 3scale", //TODO maybe refactor the tag values
	}
	return annotations
}

func buildTemplateMessage() string {
	return "Login on https://${TENANT_NAME}-admin.${WILDCARD_DOMAIN} as ${ADMIN_USERNAME}/${ADMIN_PASSWORD}"
}

func buildTemplateParameters() []templatev1.Parameter {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "AMP_RELEASE",
			Description: "AMP release tag.",
			Required:    true,
			Value:       "2.6.0",
		},
		templatev1.Parameter{
			Name:        "APP_LABEL",
			Description: "Used for object app labels",
			Required:    true,
			Value:       "3scale-api-management",
		},
		templatev1.Parameter{
			Name:        "TENANT_NAME",
			Description: "Tenant name under the root that Admin UI will be available with -admin suffix.",
			Required:    true,
			Value:       "3scale",
		},
		templatev1.Parameter{
			Name:        "RWX_STORAGE_CLASS",
			Description: "The Storage Class to be used by ReadWriteMany PVCs",
			Required:    false,
			Value:       "null",
			//Value: "null", //TODO this is incorrect. The value that should be used at the end is null and not "null", however template parameters only accept strings.
		},
	}
	return parameters
}
