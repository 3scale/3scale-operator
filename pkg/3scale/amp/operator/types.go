package operator

import (
	ampv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/amp/v1alpha1"
)

// TODO probably this could be in only one type where the different
// OptionsProvider interfaces are implemented. For the moment we
// duplicate the data to allow having possible future differences in
// the required data
type OperatorAmpImagesOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}

type OperatorRedisOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}

type OperatorBackendOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}

type OperatorMysqlOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}

type OperatorMemcachedOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}

type OperatorSystemOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}

type OperatorZyncOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}

type OperatorApicastOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}

type OperatorWildcardRouterOptionsProvider struct {
	AmpSpec *ampv1alpha1.AMPSpec
}
