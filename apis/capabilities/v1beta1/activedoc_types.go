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
	"regexp"
	"strings"

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
	ActiveDocKind = "ActiveDoc"

	// ActiveDocInvalidConditionType represents that the combination of configuration
	// in the ActiveDocSpec is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	ActiveDocInvalidConditionType common.ConditionType = "Invalid"

	// ActiveDocOrphanConditionType represents that the configuration in the ActiveDocSpec
	// contains reference to non existing resource.
	// This is (should be) a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the ActiveDocSpec references non existing product resource
	ActiveDocOrphanConditionType common.ConditionType = "Orphan"

	// ActiveDocReadyConditionType indicates the activedoc has been successfully synchronized.
	// Steady state
	ActiveDocReadyConditionType common.ConditionType = "Ready"

	// ActiveDocFailedConditionType indicates that an error occurred during synchronization.
	// The operator will retry.
	ActiveDocFailedConditionType common.ConditionType = "Failed"
)

var (
	//
	activeDocSystemNameRegexp = regexp.MustCompile("[^a-zA-Z0-9]+")
)

// ActiveDocOpenAPIRefSpec Reference to the OpenAPI Specification
type ActiveDocOpenAPIRefSpec struct {
	// SecretRef refers to the secret object that contains the OpenAPI Document
	// +optional
	SecretRef *corev1.ObjectReference `json:"secretRef,omitempty"`

	// URL Remote URL from where to fetch the OpenAPI Document
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	// +optional
	URL *string `json:"url,omitempty"`
}

// ActiveDocSpec defines the desired state of ActiveDoc
type ActiveDocSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`

	// Name is human readable name for the activedoc
	Name string `json:"name"`

	// SystemName identifies uniquely the activedoc within the account provider
	// Default value will be sanitized Name
	// +kubebuilder:validation:Pattern=`^[a-z0-9]+$`
	// +optional
	SystemName *string `json:"systemName,omitempty"`

	// Description is a human readable text of the activedoc
	// +optional
	Description *string `json:"description,omitempty"`

	// ActiveDocOpenAPIRef Reference to the OpenAPI Specification
	ActiveDocOpenAPIRef ActiveDocOpenAPIRefSpec `json:"activeDocOpenAPIRef"`

	// ProductSystemName identifies uniquely the product
	// +optional
	ProductSystemName *string `json:"productSystemName,omitempty"`

	// Published switch to publish the activedoc
	// +optional
	Published *bool `json:"published,omitempty"`

	// SkipSwaggerValidations switch to skip OpenAPI validation
	// +optional
	SkipSwaggerValidations *bool `json:"skipSwaggerValidations,omitempty"`
}

// ActiveDocStatus defines the observed state of ActiveDoc
type ActiveDocStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	ID *int64 `json:"activeDocId,omitempty"`

	// ProviderAccountHost contains the 3scale account's provider URL
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// ProductResourceName references the managed 3scale product
	// +optional
	ProductResourceName *corev1.LocalObjectReference `json:"productResourceName,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Backend Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the activedoc resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (o *ActiveDocStatus) Equals(other *ActiveDocStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(o.ID, other.ID) {
		diff := cmp.Diff(o.ID, other.ID)
		logger.V(1).Info("ID not equal", "difference", diff)
		return false
	}

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
// +kubebuilder:printcolumn:JSONPath=".status.providerAccountHost",name="Provider Account",type=string
// +kubebuilder:printcolumn:JSONPath=".status.conditions[?(@.type=='Ready')].status",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".status.activeDocId",name="3scale ID",type=integer

// ActiveDoc is the Schema for the activedocs API
type ActiveDoc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ActiveDocSpec   `json:"spec,omitempty"`
	Status ActiveDocStatus `json:"status,omitempty"`
}

func (a *ActiveDoc) SetDefaults(logger logr.Logger) bool {
	updated := false

	// Respect 3scale API defaults
	// CRD OpenAPI validation ensures systemName is not empty and it is lowercase
	if a.Spec.SystemName == nil {
		tmp := activeDocSystemNameRegexp.ReplaceAllString(a.Spec.Name, "")
		// 3scale API ignores case of the system name field
		tmp = strings.ToLower(tmp)
		a.Spec.SystemName = &tmp
		updated = true
	}

	if a.Spec.ActiveDocOpenAPIRef.SecretRef != nil && a.Spec.ActiveDocOpenAPIRef.SecretRef.Namespace == "" {
		a.Spec.ActiveDocOpenAPIRef.SecretRef.Namespace = a.GetNamespace()
		updated = true
	}

	return updated
}

func (a *ActiveDoc) Validate() field.ErrorList {
	errors := field.ErrorList{}
	return errors
}

// +kubebuilder:object:root=true

// ActiveDocList contains a list of ActiveDoc
type ActiveDocList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ActiveDoc `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ActiveDoc{}, &ActiveDocList{})
}
