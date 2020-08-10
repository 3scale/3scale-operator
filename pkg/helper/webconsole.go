package helper

import (
	"context"
	"fmt"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/RHsyseng/operator-utils/pkg/logs"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"github.com/RHsyseng/operator-utils/pkg/utils/kubernetes"
	consolev1 "github.com/openshift/api/console/v1"
	routev1 "github.com/openshift/api/route/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var logger = logs.GetLogger("openshift-webconsole")

// ConsoleLinkText is the text of the consoleLink shown on the webconsole
const ConsoleLinkText = "APIManager - 3scale"

// ConsoleLinkSupported checks if a ConsoleLink CRD exists
func ConsoleLinkSupported() error {
	gvk := schema.GroupVersionKind{Group: "console.openshift.io", Version: "v1", Kind: "ConsoleLink"}
	return kubernetes.CustomResourceDefinitionExists(gvk)
}

// CreateConsoleLink creates a ConsoleLink object if it doesn't already exist
func CreateConsoleLink(ctx context.Context, c client.Client, apimanager *appsv1alpha1.APIManager) {
	route := getRoute(ctx, c, apimanager)
	if route != nil {
		consoleLinkName := fmt.Sprintf("%s-%s", apimanager.ObjectMeta.Name, apimanager.Namespace)
		consoleLink := &consolev1.ConsoleLink{}
		err := c.Get(ctx, types.NamespacedName{Name: consoleLinkName}, consoleLink)
		if err != nil && apierrors.IsNotFound(err) {
			consoleLink = createConsoleLinkPointer(consoleLinkName, route, apimanager)
			if err := c.Create(ctx, consoleLink); err != nil {
				logger.Error(err, "Console link is not created.")
			} else {
				logger.Info("Console link has been created.", consoleLinkName)
			}
		} else if err == nil && consoleLink != nil {
			reconcileConsoleLink(ctx, route, consoleLink, c)
		}
	}
}

func getRoute(ctx context.Context, c client.Client, apimanager *appsv1alpha1.APIManager) *routev1.Route {
	reader := read.New(c).WithNamespace(apimanager.Namespace)
	deployedRoutes, err := reader.List(&routev1.RouteList{})
	if err != nil {
		return nil
	}
	for _, route := range deployedRoutes {	//look for 3scale admin routes, capture route object
		realr := route.(*routev1.Route)
		if strings.Compare(realr.Spec.To.Name, "system-provider") == 0 {
			return realr
		}
	}
	return nil
}

func reconcileConsoleLink(ctx context.Context, route *routev1.Route, consoleLink *consolev1.ConsoleLink, c client.Client) {
	url := "https://" + route.Spec.Host
	linkTxt := ConsoleLinkText
	if url != consoleLink.Spec.Href || linkTxt != consoleLink.Spec.Text {
		consoleLink.Spec.Href = url
		consoleLink.Spec.Text = linkTxt
		if err := c.Update(ctx, consoleLink); err != nil {
			logger.Error(err, "failed to reconcile Console Link", consoleLink)
		}
	}
}

func createConsoleLinkPointer(consoleLinkName string, route *routev1.Route, apimanager *appsv1alpha1.APIManager) *consolev1.ConsoleLink {
	return &consolev1.ConsoleLink{
		ObjectMeta: metav1.ObjectMeta{
			Name: consoleLinkName,
			Labels: map[string]string{
				"3scale.net/name": apimanager.ObjectMeta.Name,
			},
		},
		Spec: consolev1.ConsoleLinkSpec{
			Link: consolev1.Link{
				Text: ConsoleLinkText,
				Href: "https://" + route.Spec.Host,
			},
			Location: consolev1.NamespaceDashboard,
			NamespaceDashboard: &consolev1.NamespaceDashboardSpec{
				Namespaces: []string{apimanager.Namespace},
			},
		},
	}
}

// RemoveConsoleLink removes a ConsoleLink object if it exists
func RemoveConsoleLink(ctx context.Context, c client.Client, apimanager *appsv1alpha1.APIManager) {
	consoleLink := &consolev1.ConsoleLink{}
	consoleLinkName := fmt.Sprintf("%s-%s", apimanager.ObjectMeta.Name, apimanager.Namespace)
	err := c.Get(ctx, types.NamespacedName{Name: consoleLinkName}, consoleLink)
	if err == nil && consoleLink != nil {
		err = c.Delete(ctx, consoleLink)
		if err != nil {
			logger.Error(err, "Failed to delete the consolelink:", consoleLinkName)
		} else {
			logger.Info("Deleted the consolelink:", consoleLinkName)
		}
	}
}
