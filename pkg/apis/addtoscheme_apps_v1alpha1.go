package apis

import (
	"github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	consolev1 "github.com/openshift/api/console/v1"

)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		routev1.Install,
		consolev1.Install,
		v1alpha1.SchemeBuilder.AddToScheme)
}
