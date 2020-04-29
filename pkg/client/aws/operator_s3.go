package aws

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type OperatorS3Client struct {
	s3Client S3Client
}

func NewOperatorS3Client(s3Client S3Client) *OperatorS3Client {
	return &OperatorS3Client{
		s3Client: s3Client,
	}
}

func (o *OperatorS3Client) HeadObject(bucket string, key string) error {
	input := &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	_, err := o.s3Client.HeadObject(input)
	return err
}

func (o *OperatorS3Client) GetObject(bucket string, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	res, err := o.s3Client.GetObject(input)
	if err != nil {
		return nil, err
	}

	body := res.Body
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (b *OperatorS3Client) objectExists(bucket string, key string) error {
	err := b.HeadObject(bucket, key)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// At the moment of writing this it seems that the HeadObject method for S3 does not return
			// ErrCodeNoSuchKey in case it not exists. It returns "NotFound" code. There is no
			// constant in the AWS Go SDK for "NotFound" in S3 so we have needed to hardcode it
			// (we can create our own constant if we end up needing to check that in several more places)
			// See related issue: https://github.com/aws/aws-sdk-go/issues/2095
			if awsErr.Code() != s3.ErrCodeNoSuchKey && awsErr.Code() != "NotFound" {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
