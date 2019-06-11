package product

import "fmt"

type Version string

const (
	ProductUpstream    Version = "upstream"
	ProductRelease_2_5 Version = "2.5"
)

func NewImageProvider(productVersion Version) (ImageProvider, error) {
	switch productVersion {
	case ProductRelease_2_5:
		return &release_2_5{}, nil
	case ProductUpstream:
		return &upstream{}, nil
	default:
		return nil, fmt.Errorf("Product version '%s' is not a valid product version", productVersion)
	}
}

type ImageProvider interface {
	GetApicastImage() string
	GetBackendImage() string
	GetBackendRedisImage() string
	GetSystemImage() string
	GetSystemRedisImage() string
	GetSystemMySQLImage() string
	GetSystemPostgreSQLImage() string
	GetSystemMemcachedImage() string
	GetWildcardRouterImage() string
	GetZyncImage() string
	GetZyncPostgreSQLImage() string
}

func IsProductizedVersion(productVersion Version) bool {
	return productVersion != ProductUpstream
}

func CurrentProductVersion() Version {
	return ProductRelease_2_5
}
