package operator

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
)

// TODO probably this could be in only one type where the different
// OptionsProvider interfaces are implemented. For the moment we
// duplicate the data to allow having possible future differences in
// the required data
type OperatorAmpImagesOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorRedisOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorBackendOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorMysqlOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorMemcachedOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorSystemOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorZyncOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorApicastOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorWildcardRouterOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorProductizedOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorS3OptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}

type OperatorHighAvailabilityOptionsProvider struct {
	APIManagerSpec *appsv1alpha1.APIManagerSpec
}
