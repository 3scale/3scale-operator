package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

func init() {
	// TemplateFactories is a list of template factories
	TemplateFactories = append(TemplateFactories, NewAmpS3TemplateFactory)
}

type AmpS3TemplateAdapter struct {
}

func (a *AmpS3TemplateAdapter) Adapt(template *templatev1.Template) {
	template.Name = "3scale-api-management-s3"
	template.Annotations["description"] = "3scale API Management main system with shared file storage in AWS S3."
}

type AmpS3TemplateFactory struct {
}

func (f *AmpS3TemplateFactory) Adapters() []adapters.Adapter {
	return []adapters.Adapter{
		adapters.NewImagesAdapter(),
		adapters.NewSystemMysqlImageAdapter(),
		adapters.NewRedisAdapter(),
		adapters.NewBackendAdapter(),
		adapters.NewMysqlAdapter(),
		adapters.NewMemcachedAdapter(),
		adapters.NewSystemAdapter(component.SystemFileStorageTypeS3),
		adapters.NewZyncAdapter(),
		adapters.NewApicastAdapter(),
		&AmpS3TemplateAdapter{},
	}
}

func (f *AmpS3TemplateFactory) Type() TemplateType {
	return "amp-s3-template"
}

func NewAmpS3TemplateFactory() TemplateFactory {
	return &AmpS3TemplateFactory{}
}
