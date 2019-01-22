package apis

import (
	"github.com/3scale/3scale-operator/pkg/apis/amp/v1alpha1"

	appsv1 "github.com/openshift/api/apps/v1"
	authorizationv1 "github.com/openshift/api/authorization/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	registerOpenShiftTypes()
}

// TODO finish adding all necessary types
func registerOpenShiftTypes() {
	AddToSchemes = append(AddToSchemes,
		appsv1.Install,
		authorizationv1.Install,
		buildv1.Install,
		imagev1.Install,
		routev1.Install,
	)
}
