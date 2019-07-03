package template

import "github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpS3TemplateFactory)
}

type AmpS3TemplateFactory struct {
}

func (f *AmpS3TemplateFactory) Adapters() []adapters.Adapter {
	ampFactory := NewAmpTemplateFactory()
	return append(ampFactory.Adapters(), adapters.NewS3Adapter())
}

func (f *AmpS3TemplateFactory) Type() TemplateType {
	return "amp-s3-template"
}

func NewAmpS3TemplateFactory() TemplateFactory {
	return &AmpS3TemplateFactory{}
}
