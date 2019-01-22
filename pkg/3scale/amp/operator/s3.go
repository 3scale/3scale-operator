package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func (o *OperatorS3OptionsProvider) GetS3Options() (*component.S3Options, error) {
	sob := component.S3OptionsBuilder{}
	sob.AwsAccessKeyId(*o.AmpSpec.AwsAccessKeyId)
	sob.AwsSecretAccessKey(*o.AmpSpec.AwsSecretAccessKey)
	sob.AwsRegion(*o.AmpSpec.AwsRegion)
	sob.AwsBucket(*o.AmpSpec.AwsBucket)
	sob.FileUploadStorage(*o.AmpSpec.FileUploadStorage)
	res, err := sob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create S3 Options - %s", err)
	}
	return res, nil
}
