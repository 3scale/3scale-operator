package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APISpec defines the desired state of API
type APISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	APIBase
	APISelectors
}

type APIBase struct {
	Description       string            `json:"description"`
	IntegrationMethod IntegrationMethod `json:"integrationMethod"`
}

type APISelectors struct {
	PlanSelector   metav1.LabelSelector `json:"planSelector"`
	MetricSelector metav1.LabelSelector `json:"metricSelector"`
}

// APIStatus defines the observed state of API
type APIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// API is the Schema for the apis API
// +k8s:openapi-gen=true
type API struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APISpec   `json:"spec,omitempty"`
	Status APIStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIList contains a list of API
type APIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []API `json:"items"`
}

func init() {
	SchemeBuilder.Register(&API{}, &APIList{})
}

type IntegrationMethod struct {
	ApicastOnPrem *ApicastOnPrem `json:"apicastOnPrem"`
	CodePlugin    *CodePlugin    `json:"codePlugin"`
	ApicastHosted *ApicastHosted `json:"apicastHosted"`
}

func (api *API) getIntegrationMethodType() string {
	if api.Spec.IntegrationMethod.ApicastHosted != nil {
		return "ApicastHosted"
	} else if api.Spec.IntegrationMethod.ApicastOnPrem != nil {
		return "ApicastOnPrem"
	} else if api.Spec.IntegrationMethod.CodePlugin != nil {
		return "CodePlugin"
	}
	return ""
}

type ApicastHosted struct {
	APIcastBaseOptions
	APIcastBaseSelectors
}

type APIcastBaseOptions struct {
	PrivateBaseURL         string                        `json:"privateBaseURL"`
	APITestGetRequest      string                        `json:"apiTestGetRequest"`
	AuthenticationSettings ApicastAuthenticationSettings `json:"authenticationSettings"`
}

type APIcastBaseSelectors struct {
	MappingRulesSelector metav1.LabelSelector `json:"mappingRulesSelector"`
	PoliciesSelector     metav1.LabelSelector `json:"policiesSelector"`
}

type ApicastOnPrem struct {
	APIcastBaseOptions
	StagingPublicBaseURL    string `json:"stagingPublicBaseURL"`
	ProductionPublicBaseURL string `json:"productionPublicBaseURL"`
	APIcastBaseSelectors
}

type ApicastAuthenticationSettings struct {
	HostHeader  string                 `json:"hostHeader"`
	SecretToken string                 `json:"secretToken"`
	Credentials IntegrationCredentials `json:"credentials"`
	Errors      Errors                 `json:"errors"`
}

type APIKey struct {
	AuthParameterName   string `json:"authParameterName"`
	CredentialsLocation string `json:"credentialsLocation"`
}

type AppID struct {
	AppIDParameterName  string `json:"appIDParameterName"`
	AppKeyParameterName string `json:"appKeyParameterName"`
	CredentialsLocation string `json:"credentialsLocation"`
}

type Errors struct {
	AuthenticationFailed  Authentication `json:"authenticationFailed"`
	AuthenticationMissing Authentication `json:"authenticationMissing"`
}

type Authentication struct {
	ResponseCode int64  `json:"responseCode"`
	ContentType  string `json:"contentType"`
	ResponseBody string `json:"responseBody"`
}

type MatchLabels struct {
	API string `json:"api"`
}

type IntegrationCredentials struct {
	APIKey          *APIKey          `json:"apiKey"`
	AppID           *AppID           `json:"appID"`
	OpenIDConnector *OpenIDConnector `json:"openIDConnector"`
}

type OpenIDConnector struct {
	Issuer              string `json:"issuer"`
	CredentialsLocation string `json:"credentialsLocation"`
}

type CodePlugin struct {
	AuthenticationSettings CodePluginAuthenticationSettings `json:"authenticationSettings"`
}

type CodePluginAuthenticationSettings struct {
	Credentials IntegrationCredentials `json:"credentials"`
}
