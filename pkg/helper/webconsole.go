package helper

import (
	"context"
	"fmt"
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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
)

var logger = logs.GetLogger("openshift-webconsole")

// ConsoleLinkText is the text of the consoleLink shown on the webconsole
const ConsoleLinkText = "APIManager - 3scale Master Portal"

// ConsoleLinkNamePrefix is the prefix applied to Console link for system master
const ConsoleLinkNamePrefix = "system-master-link"

// ConsoleLinkSupported checks if a ConsoleLink CRD exists
func ConsoleLinkSupported() error {
	gvk := schema.GroupVersionKind{Group: "console.openshift.io", Version: "v1", Kind: "ConsoleLink"}
	return kubernetes.CustomResourceDefinitionExists(gvk)
}

// ReconcileConsoleLink creates/deletes/updates a ConsoleLink based on the Route
func ReconcileConsoleLink(ctx context.Context, c client.Client, req *reconcile.Request) error {
	route := getRoute(ctx, c, req)
	if route == nil {
		if err := removeConsoleLink(ctx, c, req); err != nil {
			return err
		}
	} else {
		if err := reconcileConsoleLink(ctx, c, req, route); err != nil {
			return err
		}
	}
	return nil
}

// getRoute retrieves all the routes in the current namespace and return the desired one.
func getRoute(ctx context.Context, c client.Client, req *reconcile.Request) *routev1.Route {
	reader := read.New(c).WithNamespace(req.Namespace)
	deployedRoutes, err := reader.List(&routev1.RouteList{})
	if err == nil {
		for _, route := range deployedRoutes { //capture targeted route object
			realr := route.(*routev1.Route)
			if strings.Compare(realr.Spec.To.Name, "system-master") == 0 {
				return realr
			}
		}
	}
	return nil
}

// reconcileConsoleLink create a ConsoleLink for the Route, or update the existing ConsoleLink if it exists.
// return error if update or creation could not be done
func reconcileConsoleLink(ctx context.Context, c client.Client, req *reconcile.Request, route *routev1.Route) error {
	consoleLinkName := fmt.Sprintf("%s-%s", ConsoleLinkNamePrefix, req.Namespace)
	consoleLink := &consolev1.ConsoleLink{}
	err := c.Get(ctx, types.NamespacedName{Name: consoleLinkName}, consoleLink)
	if err != nil && apierrors.IsNotFound(err) {
		consoleLink = getConsoleLink(consoleLinkName, route, req)
		if err := c.Create(ctx, consoleLink); err != nil {
			err = fmt.Errorf("failed to create consolelink %s %w", consoleLinkName, err)
			logger.Error(err)
			return err
		}
		logger.Info("consolelink has been created:", consoleLinkName)
	} else if err == nil && consoleLink != nil {
		if err = updateConsoleLink(ctx, route, consoleLink, c); err != nil {
			return err
		}
	} else {
		err = fmt.Errorf("failed to retrieve consolelink %s %w", consoleLinkName, err)
		logger.Error(err)
		return err
	}
	return nil
}

// removeConsoleLink removes a ConsoleLink object if it exists
// error is returned if the deletion is failed.
func removeConsoleLink(ctx context.Context, c client.Client, req *reconcile.Request) error {
	consoleLink := &consolev1.ConsoleLink{}
	consoleLinkName := fmt.Sprintf("%s-%s", ConsoleLinkNamePrefix, req.Namespace)
	err := c.Get(ctx, types.NamespacedName{Name: consoleLinkName}, consoleLink)
	if err == nil && consoleLink != nil {
		err = c.Delete(ctx, consoleLink)
		if err != nil {
			err = fmt.Errorf("failed to delete the consolelink %s %w", consoleLinkName, err)
			logger.Error(err)
			return err
		}
		logger.Info("deleted the consolelink:", consoleLinkName)
	}
	return nil
}

// updateConsoleLink update the ConsoleLink with properties from a Route
// error is returned if update is failed
func updateConsoleLink(ctx context.Context, route *routev1.Route, consoleLink *consolev1.ConsoleLink, c client.Client) error {
	url := "https://" + route.Spec.Host
	linkTxt := ConsoleLinkText
	if url != consoleLink.Spec.Href || linkTxt != consoleLink.Spec.Text {
		consoleLink.Spec.Href = url
		consoleLink.Spec.Text = linkTxt
		if err := c.Update(ctx, consoleLink); err != nil {
			err = fmt.Errorf("failed to update consolelink %s %w", consoleLink, err)
			logger.Error(err)
			return err
		}
	}
	return nil
}

func getConsoleLink(consoleLinkName string, route *routev1.Route, req *reconcile.Request) *consolev1.ConsoleLink {
	return &consolev1.ConsoleLink{
		ObjectMeta: metav1.ObjectMeta{
			Name: consoleLinkName,
			Labels: map[string]string{
				"3scale.net/route-name": req.Name,
			},
		},
		Spec: consolev1.ConsoleLinkSpec{
			Link: consolev1.Link{
				Text: ConsoleLinkText,
				Href: "https://" + route.Spec.Host,
			},
			Location: consolev1.NamespaceDashboard,
			NamespaceDashboard: &consolev1.NamespaceDashboardSpec{
				Namespaces: []string{req.Namespace},
			},
		},
	}
}
