package template

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/template/adapters"
	templatev1 "github.com/openshift/api/template/v1"
)

type EvalS3Adapter struct {
}

func (e *EvalS3Adapter) Adapt(template *templatev1.Template) {
	template.Name = "3scale-api-management-eval-s3"
	template.ObjectMeta.Annotations["description"] = "3scale API Management main system (Evaluation) with shared file storage in AWS S3."
}

// AmpEvalS3TemplateAdapters defines the list of adapters to build the template
func AmpEvalS3TemplateAdapters(options []string) []adapters.Adapter {
	adapterList := AmpTemplateAdapters(options)
	evalAdapter := adapters.NewEvalAdapter(options)
	s3Adapter := adapters.NewS3Adapter(options)

	return append(adapterList, evalAdapter, s3Adapter, &EvalS3Adapter{})
}
