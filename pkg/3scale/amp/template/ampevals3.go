package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpEvalS3TemplateFactory)
}

type AmpEvalS3TemplateAdapter struct {
}

func (e *AmpEvalS3TemplateAdapter) Adapt(template *templatev1.Template) {
	template.Name = "3scale-api-management-eval-s3"
	template.Annotations["description"] = "3scale API Management main system (Evaluation) with shared file storage in AWS S3."
}

type AmpEvalS3TemplateFactory struct {
}

func (f *AmpEvalS3TemplateFactory) Adapters() []adapters.Adapter {
	ampFactory := NewAmpTemplateFactory()
	evalAdapter := adapters.NewEvalAdapter()
	s3Adapter := adapters.NewS3Adapter()

	return append(ampFactory.Adapters(), evalAdapter, s3Adapter, &AmpEvalS3TemplateAdapter{})
}

func (f *AmpEvalS3TemplateFactory) Type() TemplateType {
	return "amp-eval-s3-template"
}

func NewAmpEvalS3TemplateFactory() TemplateFactory {
	return &AmpEvalS3TemplateFactory{}
}
