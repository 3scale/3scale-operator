package helper

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/common"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConsoleLinkText is the text of the consoleLink shown on the webconsole
const ConsoleLinkText = "APIManager - 3scale"

// ConsoleLinkMasterNamePrefix is the prefix applied to Console link for system master
const ConsoleLinkMasterNamePrefix = "system-master-link"

// GenericConsoleLinkMutator performs the reconciliation for consolelink objects
func GenericConsoleLinkMutator(existingObj, desiredObj common.KubernetesObject) (bool, error) {
	existing, ok := existingObj.(*consolev1.ConsoleLink)
	if !ok {
		return false, fmt.Errorf("%T is not a *consolev1.ConsoleLink", existingObj)
	}
	desired, ok := desiredObj.(*consolev1.ConsoleLink)
	if !ok {
		return false, fmt.Errorf("%T is not a *consolev1.ConsoleLink", desiredObj)
	}

	update := false

	if existing.Spec.Href != desired.Spec.Href {
		existing.Spec.Href = desired.Spec.Href
		update = true
	}

	if existing.Spec.Text != desired.Spec.Text {
		existing.Spec.Text = desired.Spec.Text
		update = true
	}

	return update, nil
}

// GetMasterConsoleLink creates the consolelink obj for a target
func GetMasterConsoleLink(route *routev1.Route) *consolev1.ConsoleLink {
	return &consolev1.ConsoleLink{
		ObjectMeta: metav1.ObjectMeta{
			Name: GetMasterConsoleLinkName(route.Namespace),
			Labels: map[string]string{
				"3scale.net/route-name": route.Name,
			},
		},
		Spec: consolev1.ConsoleLinkSpec{
			Link: consolev1.Link{
				Text: ConsoleLinkText,
				Href: "https://" + route.Spec.Host,
			},
			Location: consolev1.NamespaceDashboard,
			NamespaceDashboard: &consolev1.NamespaceDashboardSpec{
				Namespaces: []string{route.Namespace},
			},
		},
	}
}

// GetMasterConsoleLinkName returns the consolelink name
func GetMasterConsoleLinkName(namespace string) string {
	return fmt.Sprintf("%s-%s", ConsoleLinkMasterNamePrefix, namespace)
}
