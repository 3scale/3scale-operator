package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APIcastSpec defines the desired state of APIcast
// +k8s:openapi-gen=true
type APIcastSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Replicas    *int64                       `json:"replicas"`
	AdminPortal APIcastThreescaleAdminPortal `json:"adminPortal"`
	// +optional
	EnvironmentConfigurationSecretRef *v1.SecretEnvSource `json:"environmentConfigurationSecretRef,omitempty"`
	// +optional
	ServiceAccount *string `json:"serviceAccount,omitempty"`
	// +optional
	Image *string `json:"image,omitempty"`
	// +optional
	ExposedHostname *string `json:"exposedHostname,omitempty"`
}

type APIcastThreescaleAdminPortal struct {
	URLSecretKeyRef         v1.SecretKeySelector `json:"urlSecretKeyRef"`
	AccessTokenSecretKeyRef v1.SecretKeySelector `json:"accessTokenSecretKeyRef"`
}

type APIcastPodConfiguration struct {
}

// APIcastStatus defines the observed state of APIcast
// +k8s:openapi-gen=true
type APIcastStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Deployed bool `json:"deployed"`

	// ObservedGeneration reflects the generation of the most recently observed ReplicaSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents the latest available observations of a replica set's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []APIcastCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIcast is the Schema for the apicasts API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type APIcast struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APIcastSpec   `json:"spec,omitempty"`
	Status APIcastStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIcastList contains a list of APIcast
type APIcastList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIcast `json:"items"`
}

type APIcastCondition struct {
	// Type of replica set condition.
	Type APIcastConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type APIcastConditionType string

func init() {
	SchemeBuilder.Register(&APIcast{}, &APIcastList{})
}
