package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/go-logr/logr"
)

// SecretToOpenAPIEventMapper is an EventHandler that maps an OAS source secret to it's corresponding OpenAPI CR
type SecretToOpenAPIEventMapper struct {
	Context   context.Context
	K8sClient client.Client
	Logger    logr.Logger
}

func (s *SecretToOpenAPIEventMapper) Map(ctx context.Context, obj client.Object) []reconcile.Request {
	openAPIList := &capabilitiesv1beta1.OpenAPIList{}

	// Filter by Secret UID
	opts := []client.ListOption{
		client.MatchingLabels{
			openAPISecretRefLabelKey: string(obj.GetUID()),
		},
	}

	err := s.K8sClient.List(ctx, openAPIList, opts...)
	if err != nil {
		s.Logger.Error(err, "failed to list OpenAPI resources")
		return nil
	}

	s.Logger.V(1).Info("Processing object", "key", client.ObjectKeyFromObject(obj), "accepted", len(openAPIList.Items) > 0)

	requests := []reconcile.Request{}
	for idx := range openAPIList.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      openAPIList.Items[idx].GetName(),
			Namespace: openAPIList.Items[idx].GetNamespace(),
		}})
	}

	return requests
}
