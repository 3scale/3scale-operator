/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"reflect"

	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	OpenAPIKind = "OpenAPI"

	// OpenAPIInvalidConditionType represents that the combination of configuration
	// in the OpenAPISpec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the spec references invalid openapi spec
	OpenAPIInvalidConditionType common.ConditionType = "Invalid"

	// OpenAPIReadyConditionType indicates the openapi resource has been successfully reconciled.
	// Steady state
	OpenAPIReadyConditionType common.ConditionType = "Ready"

	// OpenAPIFailedConditionType indicates that an error occurred during reconcilliation.
	// The operator will retry.
	OpenAPIFailedConditionType common.ConditionType = "Failed"
)

// OpenAPIRefSpec Reference to the OpenAPI Specification
type OpenAPIRefSpec struct {
	// SecretRef refers to the secret object that contains the OpenAPI Document
	// +optional
	SecretRef *corev1.ObjectReference `json:"secretRef,omitempty"`

	// URL Remote URL from where to fetch the OpenAPI Document
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	// +optional
	URL *string `json:"url,omitempty"`
}

// OpenAPISpec defines the desired state of OpenAPI
type OpenAPISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// OpenAPIRef Reference to the OpenAPI Specification
	OpenAPIRef OpenAPIRefSpec `json:"openapiRef"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`

	// ProductionPublicBaseURL Custom public production URL
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	// +optional
	ProductionPublicBaseURL *string `json:"productionPublicBaseURL,omitempty"`

	// StagingPublicBaseURL Custom public staging URL
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	// +optional
	StagingPublicBaseURL *string `json:"stagingPublicBaseURL,omitempty"`

	// ProductSystemName 3scale product system name
	// +optional
	ProductSystemName *string `json:"productSystemName,omitempty"`

	// PrivateBaseURL Custom private base URL
	// +optional
	PrivateBaseURL *string `json:"privateBaseURL,omitempty"`

	// PrefixMatching Use prefix matching instead of strict matching on mapping rules derived from openapi operations
	// +optional
	PrefixMatching *bool `json:"prefixMatching,omitempty"`

	// PrivateAPIHostHeader Custom host header sent by the API gateway to the private API
	// +optional
	PrivateAPIHostHeader *string `json:"privateAPIHostHeader,omitempty"`

	// PrivateAPISecretToken Custom secret token sent by the API gateway to the private API
	// +optional
	PrivateAPISecretToken *string `json:"privateAPISecretToken,omitempty"`
}

// OpenAPIStatus defines the observed state of OpenAPI
type OpenAPIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ProviderAccountHost contains the 3scale account's provider URL
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// ProductResourceName references the managed 3scale product
	// +optional
	ProductResourceName *corev1.LocalObjectReference `json:"productResourceName,omitempty"`

	// BackendResourceNames contains a list of references to the managed 3scale backends
	// +optional
	BackendResourceNames []corev1.LocalObjectReference `json:"backendResourceNames,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Backend Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the openapi resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (o *OpenAPIStatus) Equals(other *OpenAPIStatus, logger logr.Logger) bool {
	if o.ProviderAccountHost != other.ProviderAccountHost {
		diff := cmp.Diff(o.ProviderAccountHost, other.ProviderAccountHost)
		logger.V(1).Info("ProviderAccountHost not equal", "difference", diff)
		return false
	}

	if o.ProductResourceName != other.ProductResourceName {
		diff := cmp.Diff(o.ProductResourceName, other.ProductResourceName)
		logger.V(1).Info("ProductResourceName not equal", "difference", diff)
		return false
	}

	if !reflect.DeepEqual(o.BackendResourceNames, other.BackendResourceNames) {
		diff := cmp.Diff(o.BackendResourceNames, other.BackendResourceNames)
		logger.V(1).Info("BackendResourceNames not equal", "difference", diff)
		return false
	}

	if o.ObservedGeneration != other.ObservedGeneration {
		diff := cmp.Diff(o.ObservedGeneration, other.ObservedGeneration)
		logger.V(1).Info("ObservedGeneration not equal", "difference", diff)
		return false
	}

	// Marshalling sorts by condition type
	currentMarshaledJSON, _ := o.Conditions.MarshalJSON()
	otherMarshaledJSON, _ := other.Conditions.MarshalJSON()
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		diff := cmp.Diff(string(currentMarshaledJSON), string(otherMarshaledJSON))
		logger.V(1).Info("Conditions not equal", "difference", diff)
		return false
	}

	return true
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OpenAPI is the Schema for the openapis API
// +kubebuilder:resource:path=openapis,scope=Namespaced
type OpenAPI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenAPISpec   `json:"spec,omitempty"`
	Status OpenAPIStatus `json:"status,omitempty"`
}

// SetDefaults set explicit defaults
func (o *OpenAPI) SetDefaults(logger logr.Logger) bool {
	updated := false

	// defaults
	if o.Spec.OpenAPIRef.SecretRef != nil && o.Spec.OpenAPIRef.SecretRef.Namespace == "" {
		o.Spec.OpenAPIRef.SecretRef.Namespace = o.GetNamespace()
		updated = true
	}

	return updated
}

func (o *OpenAPI) Validate() field.ErrorList {
	errors := field.ErrorList{}
	return errors
}

// +kubebuilder:object:root=true

// OpenAPIList contains a list of OpenAPI
type OpenAPIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenAPI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenAPI{}, &OpenAPIList{})
}
