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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

const (
	ProxyPromoteConfigKind = "ProxyPromoteConfig"

	// ProxyPromoteConfigReadyConditionType indicates the activedoc has been successfully synchronized.
	// Steady state
	ProxyPromoteConfigReadyConditionType common.ConditionType = "Ready"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProxyConfigPromoteSpec defines the desired state of ProxyConfigPromote
type ProxyConfigPromoteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// product CR metadata.name
	ProductCRName string `json:"productCRName"`

	// Environment you wish to promote to, if not present defaults to staging and if set to true promotes to production
	// +optional
	Production *bool `json:"production,omitempty"`

	// deleteCR  deletes this CR when it has successfully completed the promotion
	// +optional
	DeleteCR *bool `json:"deleteCR,omitempty"`
}

// ProxyConfigPromoteStatus defines the observed state of ProxyConfigPromote
type ProxyConfigPromoteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The id of the product that has been promoted
	//+optional
	ProductId string `json:"productId,omitempty"`

	// The latest Version in production
	//+optional
	LatestProductionVersion int `json:"latestProductionVersion,omitempty"`
	// The latest Version in staging
	//+optional
	LatestStagingVersion int `json:"latestStagingVersion,omitempty"`

	// Current state of the activedoc resource.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ProxyConfigPromote is the Schema for the proxyconfigpromotes API
type ProxyConfigPromote struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxyConfigPromoteSpec   `json:"spec,omitempty"`
	Status ProxyConfigPromoteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProxyConfigPromoteList contains a list of ProxyConfigPromote
type ProxyConfigPromoteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxyConfigPromote `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProxyConfigPromote{}, &ProxyConfigPromoteList{})
}

func (o *ProxyConfigPromoteStatus) Equals(other *ProxyConfigPromoteStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(o.ProductId, other.ProductId) {
		diff := cmp.Diff(o.ProductId, other.ProductId)
		logger.V(1).Info("ProductID not equal", "difference", diff)
		return false
	}

	if o.LatestProductionVersion != other.LatestProductionVersion {
		diff := cmp.Diff(o.LatestProductionVersion, other.LatestProductionVersion)
		logger.V(1).Info("LatestProductionVersion not equal", "difference", diff)
		return false
	}

	if o.LatestStagingVersion != other.LatestStagingVersion {
		diff := cmp.Diff(o.LatestStagingVersion, other.LatestStagingVersion)
		logger.V(1).Info("LatestStagingVersion not equal", "difference", diff)
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
