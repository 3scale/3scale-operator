package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type S3Client interface {
	HeadObject(*s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)

	ListBuckets(*s3.ListBucketsInput) (*s3.ListBucketsOutput, error)
	ListObjects(*s3.ListObjectsInput) (*s3.ListObjectsOutput, error)
	ListObjectsV2(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	DeleteObjects(*s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error)

	WaitUntilObjectExists(*s3.HeadObjectInput) error
	WaitUntilObjectNotExists(*s3.HeadObjectInput) error
}

type s3Client struct {
	internalClient s3iface.S3API
}

func (s *s3Client) HeadObject(i *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	return s.internalClient.HeadObject(i)
}

func (s *s3Client) GetObject(i *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return s.internalClient.GetObject(i)
}

func (s *s3Client) PutObject(i *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return s.internalClient.PutObject(i)
}

func (s *s3Client) DeleteObject(i *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return s.internalClient.DeleteObject(i)
}

func (s *s3Client) ListBuckets(i *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return s.internalClient.ListBuckets(i)
}

// Added for backwards-compatibility with clients that do not support the V2 version of this method
func (s *s3Client) ListObjects(i *s3.ListObjectsInput) (*s3.ListObjectsOutput, error) {
	return s.internalClient.ListObjects(i)
}

func (s *s3Client) ListObjectsV2(i *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return s.internalClient.ListObjectsV2(i)
}

func (s *s3Client) DeleteObjects(i *s3.DeleteObjectsInput) (*s3.DeleteObjectsOutput, error) {
	return s.internalClient.DeleteObjects(i)
}

func (s *s3Client) WaitUntilObjectExists(i *s3.HeadObjectInput) error {
	return s.internalClient.WaitUntilObjectExists(i)
}

func (s *s3Client) WaitUntilObjectNotExists(i *s3.HeadObjectInput) error {
	return s.internalClient.WaitUntilObjectNotExists(i)
}

func NewS3ClientWithAWSConfig(awsConfig *aws.Config) (S3Client, error) {
	s, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	return NewS3ClientFromS3API(s3.New(s))
}

func NewS3Client(accessKeyID, secretAccessKey []byte, region string) (S3Client, error) {
	awsConfig := &aws.Config{}

	if region != "" {
		awsConfig.Region = &region
	}

	awsConfig.Credentials = credentials.NewStaticCredentials(
		string(accessKeyID), string(secretAccessKey), "")

	s, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	return NewS3ClientFromS3API(s3.New(s))
}

func NewS3ClientFromS3API(s3api s3iface.S3API) (S3Client, error) {
	return &s3Client{
		internalClient: s3api,
	}, nil
}
