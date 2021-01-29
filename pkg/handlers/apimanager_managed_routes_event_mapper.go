package handlers

import (
	"context"

	appscommon "github.com/3scale/3scale-operator/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/go-logr/logr"
	appsv1 "github.com/openshift/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ handler.Mapper = &APIManagerRoutesEventMapper{}

// APIManagerRoutesEventMapper is an EventHandler that maps an existing OpenShift
// route to an APIManager. This handler should only be used on Route objects
// and when APIManager is used.
type APIManagerRoutesEventMapper struct {
	K8sClient client.Client
	Logger    logr.Logger
}

func (h *APIManagerRoutesEventMapper) Map(mapObject handler.MapObject) []reconcile.Request {
	var res []reconcile.Request
	apimanagerReconcileRequest := h.getAPIManagerOwnerReconcileRequest(mapObject.Meta)
	if apimanagerReconcileRequest != nil {
		res = append(res, *apimanagerReconcileRequest)
	}
	return res
}

func (h *APIManagerRoutesEventMapper) getAPIManagerOwnerReconcileRequest(object metav1.Object) *reconcile.Request {
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
		// or the name of the Zync Que DeploymentConfig.

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

		// If the OwnerReference of the received object is a DeploymentConfig and
		// its name is Zync Que's name then we fetch that Object and recursively
		// try to find an OwnerReference that is an APIManager. If it is found
		// we return it.
		zyncQueKind := "DeploymentConfig"
		zyncQueGroup := appsv1.GroupVersion.Group
		// An alternative to hardcode Zync-Que name would be just try to recurse
		// OwnerReferences until there are no more of them. That would be
		// potentially more costly.
		zyncQueDeploymentName := component.ZyncQueDeploymentName
		if ref.Kind == zyncQueKind && refGV.Group == zyncQueGroup && ref.Name == zyncQueDeploymentName {
			h.Logger.V(2).Info("OwnerReference to Zync-Que detected. Recursively looking for APIManager OwnerReferences...")
			existing := &appsv1.DeploymentConfig{}
			getErr := h.K8sClient.Get(context.Background(), types.NamespacedName{Name: ref.Name, Namespace: object.GetNamespace()}, existing)
			if getErr != nil {
				// If there's an error getting the object it might be due to
				// it might have been deleted already or any other kind of error.
				// In both cases we log it and ignore it and we continue the processing.
				h.Logger.Error(err, "Could not get object",
					"Kind", ref.Kind, "APIVersion", ref.APIVersion, "Name", ref.Name, "Namespace", object.GetNamespace())
			} else {
				// Recursively try to find an APIManager OwnerReference
				request := h.getAPIManagerOwnerReconcileRequest(existing)
				if request != nil {
					return request
				}
			}
		}
	}

	h.Logger.V(3).Info("No APIManager OwnerReference detected")

	return nil
}
