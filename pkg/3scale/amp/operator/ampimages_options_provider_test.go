package operator

import (
	"reflect"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/version"
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
	systemSearchdImage   = "quay.io/3scale/searchd:mytag"
)

func defaultAmpImageOptions() *component.AmpImagesOptions {
	return &component.AmpImagesOptions{
		AppLabel:                    appLabel,
		AmpRelease:                  version.ThreescaleVersionMajorMinor(),
		ApicastImage:                ApicastImageURL(),
		BackendImage:                BackendImageURL(),
		SystemImage:                 SystemImageURL(),
		ZyncImage:                   ZyncImageURL(),
		ZyncDatabasePostgreSQLImage: component.ZyncPostgreSQLImageURL(),
		SystemMemcachedImage:        SystemMemcachedImageURL(),
		SystemSearchdImage:          SystemSearchdImageURL(),
		ImagePullSecrets:            component.AmpImagesDefaultImagePullSecrets(),
	}
}

func testAmpImagesCustomImagePullSecrets() []v1.LocalObjectReference {
	return []v1.LocalObjectReference{
		{Name: "mysecret1"},
		{Name: "mysecret5"},
	}
}

func TestGetAmpImagesOptionsProvider(t *testing.T) {
	tmpApicastImage := apicastImage
	tmpBackendImage := backendImage
	tmpSystemImage := systemImage
	tmpZyncImage := zyncImage
	tmpZyncPostgresqlImage := zyncPostgresqlImage
	tmpSystemMemcachedImage := systemMemcachedImage
	tmpSystemSearchdImage := systemSearchdImage

	cases := []struct {
		name                   string
		apimanagerFactory      func() *appsv1alpha1.APIManager
		expectedOptionsFactory func() *component.AmpImagesOptions
	}{
		{
			"Default", basicApimanager, defaultAmpImageOptions,
		},
		{
			"apicastImage",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Apicast = &appsv1alpha1.ApicastSpec{Image: &tmpApicastImage}
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
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Backend = &appsv1alpha1.BackendSpec{Image: &tmpBackendImage}
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
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{Image: &tmpSystemImage}
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
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Zync = &appsv1alpha1.ZyncSpec{Image: &tmpZyncImage}
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
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.Zync = &appsv1alpha1.ZyncSpec{PostgreSQLImage: &tmpZyncPostgresqlImage}
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
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{MemcachedImage: &tmpSystemMemcachedImage}
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.SystemMemcachedImage = tmpSystemMemcachedImage
				return opts
			},
		},
		{
			"systemSearchdImage",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System.SearchdSpec = &appsv1alpha1.SystemSearchdSpec{Image: &tmpSystemSearchdImage}
				return apimanager
			},
			func() *component.AmpImagesOptions {
				opts := defaultAmpImageOptions()
				opts.SystemSearchdImage = tmpSystemSearchdImage
				return opts
			},
		},
		{
			"custom image pull secrets",
			func() *appsv1alpha1.APIManager {
				apimanager := basicApimanager()
				apimanager.Spec.System = &appsv1alpha1.SystemSpec{MemcachedImage: &tmpSystemMemcachedImage}
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
