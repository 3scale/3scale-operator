package controllers

import (
	"context"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ConfigMapToApimanagerEventMapper struct {
	K8sClient client.Client
	Logger    logr.Logger
	Namespace string
}

func (s *ConfigMapToApimanagerEventMapper) Map(obj client.Object) []reconcile.Request {
	objectName := obj.GetName()
	if objectName != helper.OperatorRequirementsConfigMapName {
		return nil
	}

	apimanagerList := &appsv1alpha1.APIManagerList{}

	// filter by Secret UID
	opts := []client.ListOption{}

	// Support namespace scope or cluster scoped
	if s.Namespace != "" {
		opts = append(opts, client.InNamespace(s.Namespace))
	}

	err := s.K8sClient.List(context.Background(), apimanagerList, opts...)
	if err != nil {
		s.Logger.Error(err, "reading apimanager list")
		return nil
	}

	requests := []reconcile.Request{}
	for idx := range apimanagerList.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      apimanagerList.Items[idx].GetName(),
			Namespace: apimanagerList.Items[idx].GetNamespace(),
		}})
	}

	return requests
}
