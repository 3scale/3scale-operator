package controllers

import (
	"context"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SecretToOpenApiEventMapper is an EventHandler that maps secret object to openApi CR's
type SecretToOpenApiEventMapper struct {
	K8sClient client.Client
	Logger    logr.Logger
	Namespace string
}

func (s *SecretToOpenApiEventMapper) Map(obj client.Object) []reconcile.Request {

	openApiList := &capabilitiesv1beta1.OpenAPIList{}

	// filter by Secret UID
	opts := []client.ListOption{client.HasLabels{openApiSecretLabelKey(string(obj.GetUID()))}}

	// Support namespace scope or cluster scoped
	if s.Namespace != "" {
		opts = append(opts, client.InNamespace(s.Namespace))
	}

	err := s.K8sClient.List(context.Background(), openApiList, opts...)
	if err != nil {
		s.Logger.Error(err, "reading openApi list")
		return nil
	}

	s.Logger.V(1).Info("Processing object", "key", client.ObjectKeyFromObject(obj), "accepted", len(openApiList.Items) > 0)

	requests := []reconcile.Request{}
	for idx := range openApiList.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      openApiList.Items[idx].GetName(),
			Namespace: openApiList.Items[idx].GetNamespace(),
		}})
	}

	return requests
}
