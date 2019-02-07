package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorS3OptionsProvider) GetS3Options() (*component.S3Options, error) {
	sob := component.S3OptionsBuilder{}
	sob.AwsAccessKeyId(*o.APIManagerSpec.AwsAccessKeyId)
	sob.AwsSecretAccessKey(*o.APIManagerSpec.AwsSecretAccessKey)
	sob.AwsRegion(*o.APIManagerSpec.AwsRegion)
	sob.AwsBucket(*o.APIManagerSpec.AwsBucket)
	sob.FileUploadStorage(*o.APIManagerSpec.FileUploadStorage)
	res, err := sob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create S3 Options - %s", err)
	}
	return res, nil
}
