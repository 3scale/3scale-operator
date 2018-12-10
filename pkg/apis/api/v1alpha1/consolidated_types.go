package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConsolidatedSpec defines the desired state of Consolidated
type ConsolidatedSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Tenants []InternalTenant `json:"tenants"`
	APIs    []InternalAPI    `json:"APIs"`
}

// ConsolidatedStatus defines the observed state of Consolidated
type ConsolidatedStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Consolidated is the Schema for the consolidateds API
// +k8s:openapi-gen=true
type Consolidated struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsolidatedSpec   `json:"spec,omitempty"`
	Status ConsolidatedStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConsolidatedList contains a list of Consolidated
type ConsolidatedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Consolidated `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Consolidated{}, &ConsolidatedList{})
}

type InternalAPI struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Credentials InternalCredential `json:"credentials"`
	Integration Integration      `json:"integration,omitempty"`
	Metrics     []InternalMetric `json:"metrics,omitempty"`
	Plans       []InternalPlan   `json:"Plans,omitempty"`
}

type Integration struct {
	ApicastOnPrem InternalApicastOnPrem `json:"apicastOnPrem"`
}

type InternalApicastOnPrem struct {
	PrivateBaseURL          string                         `json:"privateBaseURL"`
	StagingPublicBaseURL    string                         `json:"stagingPublicBaseURL"`
	ProductionPublicBaseURL string                         `json:"productionPublicBaseURL"`
	APITestGetRequest       string                         `json:"apiTestGetRequest"`
	AuthenticationSettings  InternalAuthenticationSettings `json:"authenticationSettings"`
	Errors                  InternalErrors                 `json:"errors"`
	MappingRules            []InternalMappingRule          `json:"mappingRules"`
}

type InternalAuthenticationSettings struct {
	HostHeader  string                         `json:"hostHeader"`
	SecretToken string                         `json:"secretToken"`
	Credentials InternalIntegrationCredentials `json:"credentials"`
}

type InternalIntegrationCredentials struct {
	APIKey InternalAPIKey `json:"apiKey"`
}

type InternalAPIKey struct {
	AuthParameterName   string `json:"authParameterName"`
	CredentialsLocation string `json:"credentialsLocation"`
}

type InternalErrors struct {
	AuthenticationFailed  InternalAuthentication `json:"authenticationFailed"`
	AuthenticationMissing InternalAuthentication `json:"authenticationMissing"`
}

type InternalAuthentication struct {
	ResponseCode int64  `json:"responseCode"`
	ContentType  string `json:"contentType"`
	ResponseBody string `json:"responseBody"`
}

type InternalMappingRule struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	Increment int64  `json:"increment"`
	Metric    string `json:"metric"`
}

type InternalMetric struct {
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description"`
}

type InternalPlan struct {
	Name             string          `json:"name"`
	TrialPeriodDays  int64           `json:"trialPeriodDays"`
	ApprovalRequired bool            `json:"approvalRequired"`
	Costs            PlanCost       `json:"costs"`
	Limits           []InternalLimit `json:"limits"`
}

type InternalLimit struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Period      string `json:"period"`
	MaxValue    int64  `json:"maxValue"`
	Metric      string `json:"metric"`
}

type InternalTenant struct {
	Name        string               `json:"name"`
	Credentials []InternalCredential `json:"credentials"`
}

type InternalCredential struct {
	AccessToken string `json:"accessToken"`
	AdminURL    string `json:"adminURL"`
}
