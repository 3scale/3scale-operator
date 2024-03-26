package controllers

import (
	"context"
	appscommon "github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// APIManagerConfigMapsEventMapper is an EventHandler that maps an existing OpenShift
// ConfigMaps to an APIManager. This handler should only be used on ConfigMaps objects
// and when APIManager is used.
type APIManagerConfigMapsEventMapper struct {
	Context   context.Context
	K8sClient client.Client
	Logger    logr.Logger
}

func (h *APIManagerConfigMapsEventMapper) Map(ctx context.Context, o client.Object) []reconcile.Request {
	var res []reconcile.Request
	apimanagerReconcileRequest := h.getAPIManagerOwnerReconcileRequest(o)
	if apimanagerReconcileRequest != nil {
		res = append(res, *apimanagerReconcileRequest)
	}
	return res
}

func (h *APIManagerConfigMapsEventMapper) getAPIManagerOwnerReconcileRequest(object metav1.Object) *reconcile.Request {
	h.Logger.V(2).Info("Processing meta object", "Name", object.GetName(), "Namespace", object.GetNamespace())

	for _, ref := range object.GetOwnerReferences() {
		refGV, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			h.Logger.Error(err, "Could not parse OwnerReference APIVersion",
				"api version", ref.APIVersion)
			return nil
		}

		h.Logger.V(2).Info("Evaluating OwnerReference", "GroupVersion", refGV, "Kind", ref.Kind, "Name", ref.Name)
		// Compare the OwnerReference Group and Kind against the APIManager Kind

		// If the OwnerReference of the received object is an APIManager we
		// return a reconcile Request using the name from the OwnerReference and the
		// Namespace from the object in the event
		apimanagerKind := appscommon.APIManagerKind
		apimanagerGroup := appsv1alpha1.GroupVersion.Group
		if ref.Kind == apimanagerKind && refGV.Group == apimanagerGroup {
			h.Logger.V(2).Info("APIManager OwnerReference detected. Reenqueuing as APIManager event", "APIManager name", ref.Name, "APIManager namespace", object.GetNamespace())
			// Match found - add a Request for the object referred to in the OwnerReference
			request := &reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      ref.Name,
				Namespace: object.GetNamespace(),
			}}
			return request
		}
	}

	h.Logger.V(3).Info("No APIManager OwnerReference detected")

	return nil
}
