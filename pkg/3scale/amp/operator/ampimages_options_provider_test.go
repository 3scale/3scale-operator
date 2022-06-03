package operator

import (
	"reflect"
	"testing"

	appsv1beta1 "github.com/3scale/3scale-operator/apis/apps/v1beta1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	apicastImage         = "quay.io/3scale/apicast:mytag"
	backendImage         = "quay.io/3scale/backend:mytag"
	systemImage          = "quay.io/3scale/backend:mytag"
	zyncImage            = "quay.io/3scale/zync:mytag"
	zyncPostgresqlImage  = "postgresql-10:mytag"
	systemMemcachedImage = "memcached:mytag"
)

func defaultAmpImageOptions() *component.AmpImagesOptions {
	return &component.AmpImagesOptions{
		AppLabel:                    appLabel,
		AmpRelease:                  product.ThreescaleRelease,
		ApicastImage:                ApicastImageURL(),
		BackendImage:                BackendImageURL(),
		SystemImage:                 SystemImageURL(),
		ZyncImage:                   ZyncImageURL(),
		ZyncDatabasePostgreSQLImage: component.ZyncPostgreSQLImageURL(),
		SystemMemcachedImage:        SystemMemcachedImageURL(),
		InsecureImportPolicy:        insecureImportPolicy,
		ImagePullSecrets:            component.AmpImagesDefaultImagePullSecrets(),
	}
}

func testAmpImagesCustomImagePullSecrets() []v1.LocalObjectReference {
	return []v1.LocalObjectReference{
		v1.LocalObjectReference{Name: "mysecret1"},
		v1.LocalObjectReference{Name: "mysecret5"},
	}
}

func TestGetAmpImagesOptionsProvider(t *testing.T) {
	tmpApicastImage := apicastImage
	tmpBackendImage := backendImage
	tmpSystemImage := systemImage
	tmpZyncImage := zyncImage
	tmpZyncPostgresqlImage := zyncPostgresqlImage
	tmpSystemMemcachedImage := systemMemcachedImage

	cases := []struct {
		name                   string
		apimanagerFactory      func() *appsv1beta1.APIManager
		expectedOptionsFactory func() *component.AmpImagesOptions
	}{
		{
			"Default", basicApimanager, defaultAmpImageOptions,
		},
		{
			"apicastImage",
			func() *appsv1beta1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Apicast = &appsv1beta1.ApicastSpec{Image: &tmpApicastImage}
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.ApicastImage = tmpApicastImage
				return opts
			},
		},
		{
			"backendImage",
			func() *appsv1beta1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend = &appsv1beta1.BackendSpec{Image: &tmpBackendImage}
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.BackendImage = tmpBackendImage
				return opts
			},
		},
		{
			"systemImage",
			func() *appsv1beta1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1beta1.SystemSpec{Image: &tmpSystemImage}
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.SystemImage = tmpSystemImage
				return opts
			},
		},
		{
			"zyncImage",
			func() *appsv1beta1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Zync = &appsv1beta1.ZyncSpec{Image: &tmpZyncImage}
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.ZyncImage = tmpZyncImage
				return opts
			},
		},
		{
			"zyncPostgresqlImage",
			func() *appsv1beta1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Zync = &appsv1beta1.ZyncSpec{PostgreSQLImage: &tmpZyncPostgresqlImage}
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.ZyncDatabasePostgreSQLImage = tmpZyncPostgresqlImage
				return opts
			},
		},
		{
			"systemMemcachedImage",
			func() *appsv1beta1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1beta1.SystemSpec{MemcachedImage: &tmpSystemMemcachedImage}
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.SystemMemcachedImage = tmpSystemMemcachedImage
				return opts
			},
		},
		{
			"custom image pull secrets",
			func() *appsv1beta1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1beta1.SystemSpec{MemcachedImage: &tmpSystemMemcachedImage}
				apimanager.Spec.ImagePullSecrets = testAmpImagesCustomImagePullSecrets()
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.ImagePullSecrets = testAmpImagesCustomImagePullSecrets()
				opts.SystemMemcachedImage = tmpSystemMemcachedImage
				return opts
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			optsProvider := NewAmpImagesOptionsProvider(tc.apimanagerFactory())
			opts, err := optsProvider.GetAmpImagesOptions()
			if err != nil {
				subT.Error(err)
			}
			expectedOptions := tc.expectedOptionsFactory()
			if !reflect.DeepEqual(expectedOptions, opts) {
				subT.Errorf("Resulting expected options differ: %s", cmp.Diff(expectedOptions, opts, cmpopts.IgnoreUnexported(resource.Quantity{})))
			}
		})
	}
}
