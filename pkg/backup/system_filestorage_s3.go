package backup

type SystemS3FileStorage struct {
	Bucket         string                         `validate:"required"`
	Credentials    SystemS3FileStorageCredentials `validate:"required"`
	Region         string                         `valdate:"required"` // In the case of System S3 filestorage we require region because it should exist in the specified secret
	Endpoint       *string
	Path           *string
	ForcePathStyle *bool
	DisableSSL     *bool
}

type SystemS3FileStorageCredentials struct {
	AccessKeyID     string `validate:"required"`
	SecretAccessKey string `validate:"required"`
}
