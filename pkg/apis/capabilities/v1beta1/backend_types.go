package v1beta1

import (
	"reflect"
	"regexp"

	"github.com/3scale/3scale-operator/pkg/common"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	BackendKind = "Backend"

	// BackendInvalidConditionType represents that the combination of configuration
	// in the BackendSpec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the spec references non existing internal Metric reference
	BackendInvalidConditionType common.ConditionType = "Invalid"

	// BackendSyncedConditionType indicates the product has been successfully synchronized.
	// Steady state
	BackendSyncedConditionType common.ConditionType = "Synced"

	// BackendFailedConditionType indicates that an error occurred during synchronization.
	// The operator will retry.
	BackendFailedConditionType common.ConditionType = "Failed"
)

var (
	//
	backendSystemNameRegexp = regexp.MustCompile("[^a-zA-Z0-9]+")
)

// ProductStatusError represents that the combination of configuration in the BackendSpec
// is not supported by this cluster. This is not a transient error, but
// indicates a state that must be fixed before progress can be made.
// Example: the BackendSpec references non existing internal Metric refenrece
type BackendStatusError string

// BackendSpec defines the desired state of Backend
type BackendSpec struct {
	// Name is human readable name for the backend
	Name string `json:"name"`

	// SystemName identifies uniquely the product within the account provider
	// Default value will be sanitized Name
	// +optional
	SystemName string `json:"systemName,omitempty"`

	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	PrivateBaseURL string `json:"privateBaseURL,omitempty"`

	// +optional
	Description string `json:"description,omitempty"`

	// +optional
	MappingRules []MappingRuleSpec `json:"mappingRules,omitempty"`

	// Metrics
	// Map: system_name -> MetricSpec
	// system_name attr is unique for all metrics AND methods
	// In other words, if metric's system_name is A, there is no metric or method with system_name A.
	// +optional
	Metrics map[string]MetricSpec `json:"metrics,omitempty"`

	// Methods
	// Map: system_name -> MethodSpec
	// system_name attr is unique for all metrics AND methods
	// In other words, if metric's system_name is A, there is no metric or method with system_name A.
	// +optional
	Methods map[string]Methodpec `json:"methods,omitempty"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`
}

// BackendStatus defines the observed state of Backend
type BackendStatus struct {
	// +optional
	ID *int64 `json:"backendId,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed MachineSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// In the event that there is a terminal problem reconciling the
	// replicas, both ErrorReason and ErrorMessage will be set. ErrorReason
	// will be populated with a succinct value suitable for machine
	// interpretation, while ErrorMessage will contain a more verbose
	// string suitable for logging and human consumption.
	//
	// These fields should not be set for transitive errors that a
	// controller faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Backend's spec or the configuration of
	// the backend controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the backend controller, or the
	// responsible backend controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Backends
	// can be added as events to the Backend's object and/or logged in the
	// controller's output.
	// +optional
	ErrorReason *BackendStatusError `json:"errorReason,omitempty"`
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// Current state of the 3scale backend.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (b *BackendStatus) Equals(other *BackendStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(b.ID, other.ID) {
		diff := cmp.Diff(b.ID, other.ID)
		logger.V(1).Info("ID not equal", "difference", diff)
		return false
	}

	if b.ObservedGeneration != other.ObservedGeneration {
		diff := cmp.Diff(b.ObservedGeneration, other.ObservedGeneration)
		logger.V(1).Info("ObservedGeneration not equal", "difference", diff)
		return false
	}

	if !reflect.DeepEqual(b.ErrorReason, other.ErrorReason) {
		diff := cmp.Diff(b.ErrorReason, other.ErrorReason)
		logger.V(1).Info("ErrorReason not equal", "difference", diff)
		return false
	}

	if !reflect.DeepEqual(b.ErrorMessage, other.ErrorMessage) {
		diff := cmp.Diff(b.ErrorMessage, other.ErrorMessage)
		logger.V(1).Info("ErrorMessage not equal", "difference", diff)
		return false
	}

	// Marshalling sorts by condition type
	currentMarshaledJSON, _ := b.Conditions.MarshalJSON()
	otherMarshaledJSON, _ := other.Conditions.MarshalJSON()
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		diff := cmp.Diff(string(currentMarshaledJSON), string(otherMarshaledJSON))
		logger.V(1).Info("Conditions not equal", "difference", diff)
		return false
	}

	return true
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Backend is the Schema for the backends API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=backends,scope=Namespaced
type Backend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackendSpec   `json:"spec,omitempty"`
	Status BackendStatus `json:"status,omitempty"`
}

func (backend *Backend) SetDefaults() bool {
	updated := false

	// Respect 3scale API defaults
	if backend.Spec.SystemName == "" {
		backend.Spec.SystemName = backendSystemNameRegexp.ReplaceAllString(backend.Spec.Name, "")
		updated = true
	}

	return updated
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BackendList contains a list of Backend
type BackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backend{}, &BackendList{})
}
