package controllers

import (
	"context"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
)

const zyncCreatedByLabel = "3scale.net/created-by"
const zyncCreatedByValue = "zync"

type ZyncRouteToAPIManagerMapper struct {
	Context   context.Context
	K8sClient client.Client
	Logger    logr.Logger
}

func (h *ZyncRouteToAPIManagerMapper) Map(ctx context.Context, o client.Object) []reconcile.Request {
	route, ok := o.(*routev1.Route)
	if !ok {
		return nil
	}

	apimanagerList := &appsv1alpha1.APIManagerList{}
	if err := h.K8sClient.List(ctx, apimanagerList, client.InNamespace(o.GetNamespace())); err != nil {
		h.Logger.Error(err, "failed to list APIManagers")
		return nil
	}

	var requests []reconcile.Request
	for i := range apimanagerList.Items {
		if routeHostMatchesAPIManager(route.Spec.Host, &apimanagerList.Items[i]) {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      apimanagerList.Items[i].Name,
					Namespace: apimanagerList.Items[i].Namespace,
				},
			})
		}
	}
	return requests
}

func routeHostMatchesAPIManager(routeHost string, apimanager *appsv1alpha1.APIManager) bool {
	for _, expectedHost := range apimanagerExpectedRouteHosts(apimanager) {
		if expectedHost == routeHost {
			return true
		}
	}
	return false
}
