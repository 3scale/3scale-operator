package component

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	AwsBucket          = "AWS_BUCKET"
	AwsRegion          = "AWS_REGION"
	AwsProtocol        = "AWS_PROTOCOL"
	AwsHostname        = "AWS_HOSTNAME"
	AwsPathStyle       = "AWS_PATH_STYLE"
)

type S3 struct {
	Options *S3Options
}

func NewS3(options *S3Options) *S3 {
	return &S3{Options: options}
}

// TODO Template-only object. Decide what to do with this component.
// It seems that now only the field names of the secret configuration and the
// options themselves are needed (to create this secret)
func (s3 *S3) S3AWSSecret() *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: s3.Options.AwsCredentialsSecret,
		},
		StringData: map[string]string{
			AwsAccessKeyID:     s3.Options.AwsAccessKeyId,
			AwsSecretAccessKey: s3.Options.AwsSecretAccessKey,
			AwsRegion:          s3.Options.AwsRegion,
			AwsBucket:          s3.Options.AwsBucket,
			AwsProtocol:        s3.Options.AwsProtocol,
			AwsHostname:        s3.Options.AwsHostname,
			AwsPathStyle:       s3.Options.AwsPathStyle,
		},
		Type: v1.SecretTypeOpaque,
	}
}
