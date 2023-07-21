package controllers

import (
	"context"
	"fmt"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	APImanagerSecretLabelPrefix = "secret.apimanager.apps.3scale.net/"
)

// SecretToApimanagerEventMapper is an EventHandler that maps secret object to apimanager CR's
type SecretToApimanagerEventMapper struct {
	K8sClient client.Client
	Logger    logr.Logger
	Namespace string
}

func apimanagerSecretLabelKey(uid string) string {
	return fmt.Sprintf("%s%s", APImanagerSecretLabelPrefix, uid)
}

func (s *SecretToApimanagerEventMapper) Map(obj client.Object) []reconcile.Request {

	apimanagerList := &appsv1alpha1.APIManagerList{}

	// filter by Secret UID
	opts := []client.ListOption{client.HasLabels{apimanagerSecretLabelKey(string(obj.GetUID()))}}

	// Support namespace scope or cluster scoped
	if s.Namespace != "" {
		opts = append(opts, client.InNamespace(s.Namespace))
	}

	err := s.K8sClient.List(context.Background(), apimanagerList, opts...)
	if err != nil {
		s.Logger.Error(err, "reading apimanager list")
		return nil
	}

	s.Logger.V(1).Info("Processing object", "key", client.ObjectKeyFromObject(obj), "accepted", len(apimanagerList.Items) > 0)

	requests := []reconcile.Request{}
	for idx := range apimanagerList.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      apimanagerList.Items[idx].GetName(),
			Namespace: apimanagerList.Items[idx].GetNamespace(),
		}})
	}

	return requests
}
