package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

type TemplateType string

const (
	AmpType           TemplateType = "amp-template"
	AmpS3Type         TemplateType = "amp-s3-template"
	AmpEvalType       TemplateType = "amp-eval-template"
	AmpEvalS3Type     TemplateType = "amp-eval-s3-template"
	AmpHAType         TemplateType = "amp-ha-template"
	AmpPostgresqlType TemplateType = "amp-postgresql-template"
)

// NewTemplate implements the main loop of adapters to generate template object
func NewTemplate(templateName string, componentOptions []string) *templatev1.Template {
	adapterList := FindTemplateAdapterList(templateName, componentOptions)
	tpl := Basic3scaleTemplate()
	for _, adapter := range adapterList {
		adapter.Adapt(tpl)
	}
	return tpl
}

func FindTemplateAdapterList(templateName string, componentOptions []string) []adapters.Adapter {
	switch TemplateType(templateName) {
	case AmpType:
		return AmpTemplateAdapters(componentOptions)
	case AmpS3Type:
		return AmpS3TemplateAdapters(componentOptions)
	case AmpEvalType:
		return AmpEvalTemplateAdapters(componentOptions)
	case AmpEvalS3Type:
		return AmpEvalS3TemplateAdapters(componentOptions)
	case AmpHAType:
		return AmpHATemplateAdapters(componentOptions)
	case AmpPostgresqlType:
		return AmpPostgresqlTemplateAdapters(componentOptions)
	}

	panic("Error: Template not recognized")
}
