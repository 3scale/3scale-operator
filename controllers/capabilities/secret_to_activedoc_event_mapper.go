package controllers

import (
	"context"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SecretToActiveDocEventMapper is an EventHandler that maps OpenApi secret object to ActiveDoc CR's
type SecretToActiveDocEventMapper struct {
	K8sClient client.Client
	Logger    logr.Logger
	Namespace string
}

func (s *SecretToActiveDocEventMapper) Map(obj client.Object) []reconcile.Request {

	activeDocList := &capabilitiesv1beta1.ActiveDocList{}

	// filter by Secret UID
	opts := []client.ListOption{client.HasLabels{openApiSecretLabelKey(string(obj.GetUID()))}}

	// Support namespace scope or cluster scoped
	if s.Namespace != "" {
		opts = append(opts, client.InNamespace(s.Namespace))
	}

	err := s.K8sClient.List(context.Background(), activeDocList, opts...)
	if err != nil {
		s.Logger.Error(err, "reading ActiveDoc list")
		return nil
	}

	s.Logger.V(1).Info("Processing object", "key", client.ObjectKeyFromObject(obj), "accepted", len(activeDocList.Items) > 0)

	requests := []reconcile.Request{}
	for idx := range activeDocList.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      activeDocList.Items[idx].GetName(),
			Namespace: activeDocList.Items[idx].GetNamespace(),
		}})
	}

	return requests
}
