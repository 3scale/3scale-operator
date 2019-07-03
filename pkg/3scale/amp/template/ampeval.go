package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpEvalTemplateFactory)
}

type AmpEvalTemplateFactory struct {
}

func (f *AmpEvalTemplateFactory) Adapters() []adapters.Adapter {
	ampFactory := NewAmpTemplateFactory()
	return append(ampFactory.Adapters(), adapters.NewEvalAdapter())
}

func (f *AmpEvalTemplateFactory) Type() TemplateType {
	return "amp-eval-template"
}

func NewAmpEvalTemplateFactory() TemplateFactory {
	return &AmpEvalTemplateFactory{}
}
