package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	templatev1 "github.com/openshift/api/template/v1"
)

type S3 struct {
}

func NewS3Adapter() Adapter {
	return &S3{}
}

func (s *S3) Adapt(template *templatev1.Template) {
	s3Options, err := s.options()
	if err != nil {
		panic(err)
	}
	s3Component := component.NewS3(s3Options)

	s.addParameters(template)
	s.addObjects(template, s3Component)
	s.postProcess(template, s3Component)

	// update metadata
	template.Name = "3scale-api-management-s3"
	template.ObjectMeta.Annotations["description"] = "3scale API Management main system with shared file storage in AWS S3."
}

func (s *S3) postProcess(template *templatev1.Template, s3Component *component.S3) {
	res := helper.UnwrapRawExtensions(template.Objects)
	res = s3Component.RemoveSystemStoragePVC(res)
	s3Component.RemoveSystemStorageReferences(res)
	s3Component.AddS3PostprocessOptionsToSystemEnvironmentCfgMap(helper.UnwrapRawExtensions(template.Objects))
	s3Component.AddCfgMapElemsToSystemBaseEnv(helper.UnwrapRawExtensions(template.Objects))
	s3Component.RemoveRWXStorageClassParameter(template)
	template.Objects = helper.WrapRawExtensions(res)
}

func (s *S3) addObjects(template *templatev1.Template, s3 *component.S3) {
	template.Objects = append(template.Objects, helper.WrapRawExtensions(s3.Objects())...)
}

func (s *S3) addParameters(template *templatev1.Template) {
	template.Parameters = append(template.Parameters, s.parameters()...)
}

func (s *S3) options() (*component.S3Options, error) {
	o := component.NewS3Options()
	o.AwsAccessKeyId = "${AWS_ACCESS_KEY_ID}"
	o.AwsSecretAccessKey = "${AWS_SECRET_ACCESS_KEY}"
	o.AwsRegion = "${AWS_REGION}"
	o.AwsBucket = "${AWS_BUCKET}"
	o.AwsProtocol = "${AWS_PROTOCOL}"
	o.AwsHostname = "${AWS_HOSTNAME}"
	o.AwsPathStyle = "${AWS_PATH_STYLE}"
	o.AwsCredentialsSecret = "aws-auth"

	err := o.Validate()
	return o, err
}

func (s3 *S3) parameters() []templatev1.Parameter {
	return []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "AWS_ACCESS_KEY_ID",
			Description: "AWS Access Key ID to use in S3 Storage for assets.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_SECRET_ACCESS_KEY",
			Description: "AWS Access Key Secret to use in S3 Storage for assets.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_BUCKET",
			Description: "AWS S3 Bucket Name to use in S3 Storage for assets.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_REGION",
			Description: "AWS Region to use in S3 Storage for assets.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_PROTOCOL",
			Description: "AWS S3 compatible provider endpoint protocol. HTTP or HTTPS.",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_HOSTNAME",
			Description: "AWS S3 compatible provider endpoint hostname",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "AWS_PATH_STYLE",
			Description: "When set to true, the bucket name is always left in the request URI and never moved to the host as a sub-domain",
			Required:    false,
		},
	}
}
