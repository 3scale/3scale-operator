package v1beta1

import (
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
	OpenapiKind = "Openapi"

	// OpenapiInvalidConditionType represents that the combination of configuration
	// in the OpenapiSpec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the spec references invalid openapi spec
	OpenapiInvalidConditionType common.ConditionType = "Invalid"

	// OpenapiReadyConditionType indicates the openapi resource has been successfully reconciled.
	// Steady state
	OpenapiReadyConditionType common.ConditionType = "Ready"

	// OpenapiFailedConditionType indicates that an error occurred during reconcilliation.
	// The operator will retry.
	OpenapiFailedConditionType common.ConditionType = "Failed"
)

// OpenAPIRefSpec Reference to the OpenAPI Specification
type OpenAPIRefSpec struct {
	// ConfigMapRef ConfigMap that contains the OpenAPI Document
	// +optional
	ConfigMapRef *corev1.ObjectReference `json:"configMapRef,omitempty"`

	// URL Remote URL from where to fetch the OpenAPI Document
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	// +optional
	URL *string `json:"url,omitempty"`
}

// OpenapiSpec defines the desired state of Openapi
type OpenapiSpec struct {
	// OpenAPIRef Reference to the OpenAPI Specification
	OpenAPIRef OpenAPIRefSpec `json:"openapiRef"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`

	// ProductionPublicBaseURL Custom public production URL
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	// +optional
	ProductionPublicBaseURL string `json:"productionPublicBaseURL,omitempty"`

	// StagingPublicBaseURL Custom public staging URL
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	// +optional
	StagingPublicBaseURL string `json:"stagingPublicBaseURL,omitempty"`

	// SkipOpenapiValidation Skip OpenAPI schema validation
	// +optional
	SkipOpenapiValidation bool `json:"skipOpenapiValidation,omitempty"`

	// ProductSystemName 3scale product system name
	// +optional
	ProductSystemName string `json:"productSystemName,omitempty"`
}

// OpenapiStatus defines the observed state of Openapi
type OpenapiStatus struct {
	// 3scale control plane host
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// 3scale control plane host
	// +optional
	ProductResourceName *corev1.LocalObjectReference `json:"productResourceName,omitempty"`

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

func (o *OpenapiStatus) Equals(other *OpenapiStatus, logger logr.Logger) bool {
	if o.ProviderAccountHost != other.ProviderAccountHost {
		diff := cmp.Diff(o.ProviderAccountHost, other.ProviderAccountHost)
		logger.V(1).Info("ProviderAccountHost not equal", "difference", diff)
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Openapi is the Schema for the openapis API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=openapis,scope=Namespaced
type Openapi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenapiSpec   `json:"spec,omitempty"`
	Status OpenapiStatus `json:"status,omitempty"`
}

// SetDefaults set explicit defaults
func (o *Openapi) SetDefaults(logger logr.Logger) bool {
	updated := false

	// defaults
	if o.Spec.OpenAPIRef.ConfigMapRef != nil && o.Spec.OpenAPIRef.ConfigMapRef.Namespace == "" {
		o.Spec.OpenAPIRef.ConfigMapRef.Namespace = o.GetNamespace()
		updated = true
	}

	return updated
}

func (o *Openapi) Validate() field.ErrorList {
	errors := field.ErrorList{}
	return errors
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenapiList contains a list of Openapi
type OpenapiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Openapi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Openapi{}, &OpenapiList{})
}
