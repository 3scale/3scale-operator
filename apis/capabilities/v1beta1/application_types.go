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
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

const (
	ApplicationReadyConditionType common.ConditionType = "Ready"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//AccountCRName name of account custom resource under which the application will be created
	AccountCR *corev1.LocalObjectReference `json:"accountCR"`

	//ProductCRName of product custom resource from which the application plan will be used
	ProductCR *corev1.LocalObjectReference `json:"productCR"`

	//ApplicationPlanName name of application plan that the application will use
	ApplicationPlanName string `json:"applicationPlanName"`

	//Name identifies the application uniquely within the account
	Name string `json:"name"`

	//Description human-readable text of the application
	Description string `json:"description"`

	//Suspend application if true suspends application, if false resumes application.
	//+optional
	Suspend bool `json:"suspend,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	ID *int64 `json:"applicationID,omitempty"`

	// 3scale control plane host
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// +optional
	State string `json:"state,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Application Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the 3scale application.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (b *ApplicationStatus) Equals(other *ApplicationStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(b.ID, other.ID) {
		diff := cmp.Diff(b.ID, other.ID)
		logger.V(1).Info("ID not equal", "difference", diff)
		return false
	}

	if b.ProviderAccountHost != other.ProviderAccountHost {
		diff := cmp.Diff(b.ProviderAccountHost, other.ProviderAccountHost)
		logger.V(1).Info("ProviderAccountHost not equal", "difference", diff)
		return false
	}

	if b.State != other.State {
		diff := cmp.Diff(b.State, other.State)
		logger.V(1).Info("State not equal", "difference", diff)
		return false
	}

	if b.ObservedGeneration != other.ObservedGeneration {
		diff := cmp.Diff(b.ObservedGeneration, other.ObservedGeneration)
		logger.V(1).Info("ObservedGeneration not equal", "difference", diff)
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

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
