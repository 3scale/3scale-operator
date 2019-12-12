package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpEvalTemplateFactory)
}

type AmpEvalTemplateAdapter struct {
}

func (a *AmpEvalTemplateAdapter) Adapt(template *templatev1.Template) {
	template.Name = "3scale-api-management-eval"
	template.Annotations["description"] = "3scale API Management main system (Evaluation)"
}

type AmpEvalTemplateFactory struct {
}

func (f *AmpEvalTemplateFactory) Adapters() []adapters.Adapter {
	ampFactory := NewAmpTemplateFactory()
	return append(ampFactory.Adapters(), adapters.NewEvalAdapter(), &AmpEvalTemplateAdapter{})
}

func (f *AmpEvalTemplateFactory) Type() TemplateType {
	return "amp-eval-template"
}

func NewAmpEvalTemplateFactory() TemplateFactory {
	return &AmpEvalTemplateFactory{}
}
