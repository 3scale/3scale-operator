package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProductStatusError represents that the combination of configuration in the ProductSpec
// is not supported by this cluster. This is not a transient error, but
// indicates a state that must be fixed before progress can be made.
// Example: the ProductSpec references non existing internal Metric refenrece
type ProductStatusError string

const (
	InvalidInternalMetricReferenceError ProductStatusError = "InvalidInternalMetricReferenceError"
)

// LimitSpec define the maximum value a metric can take on a contract before the user is no longer authorized to use resources.
// Once a limit has been passed in a given period, reject messages will be issued if the service is accessed under this contract.
type LimitSpec struct {
	// +kubebuilder:validation:Enum=eternity;year;month;week;day;hour;minute
	Period          string `json:"period,omitempty"`
	Value           int64  `json:"value,omitempty"`
	MetricMethodRef string `json:"metricMethodRef,omitempty"`
}

// PricingRuleSpec defines the desired state of Application Plan's Pricing Rule
type PricingRuleSpec struct {
	From            int    `json:"from,omitempty"`
	To              int    `json:"to,omitempty"`
	MetricMethodRef string `json:"metricMethodRef,omitempty"`
	// Price per unit (USD)
	// +kubebuilder:validation:Pattern=^\d+.?\d{2}$
	PricePerUnit string `json:"pricePerUnit,omitempty"`
}

// ApplicationPlanSpec defines the desired state of Product's Application Plan
type ApplicationPlanSpec struct {
	// +optional
	FriendlyName *string `json:"friendlyName,omitempty"`

	// Set whether or not applications can be created on demand
	// or if approval is required from you before they are activated.
	// +optional
	AppsRequireApproval *bool `json:"appsRequireApproval,omitempty"`

	// Trial Period (days)
	// +kubebuilder:validation:Minimum=0
	// +optional
	TrialPeriod *int `json:"trialPeriod,omitempty"`

	// Setup fee (USD)
	// +kubebuilder:validation:Pattern=^\d+.?\d{2}$
	// +optional
	SetupFee *string `json:"setupFee,omitempty"`

	// Cost per Month (USD)
	// +kubebuilder:validation:Pattern=^\d+.?\d{2}$
	// +optional
	CostMonth *string `json:"costMonth,omitempty"`

	// +optional
	// Pricing Rules
	PricingRules []PricingRuleSpec `json:"pricingRules,omitempty"`

	// +optional
	// Limits
	Limits []LimitSpec `json:"limits,omitempty"`

	// TODO Features
}

// Methodpec defines the desired state of Product's Method
type Methodpec struct {
	// +optional
	FriendlyName *string `json:"friendlyName,omitempty"`
	// +optional
	Description *string `json:"description,omitempty"`
}

// MetricSpec defines the desired state of Product's Metric
type MetricSpec struct {
	// +optional
	FriendlyName *string `json:"friendlyName,omitempty"`
	// +optional
	Description *string `json:"description,omitempty"`
	// +optional
	Unit *string `json:"unit,omitempty"`
}

// MappingRuleSpec defines the desired state of Product's MappingRule
type MappingRuleSpec struct {
	Verb            string `json:"verb,omitempty"`
	Pattern         string `json:"pattern,omitempty"`
	MetricMethodRef string `json:"metricMethodRef,omitempty"`
	// +optional
	Increment *int `json:"increment,omitempty"`
	// +optional
	Position *int `json:"position,omitempty"`
}

// BackendUsageSpec defines the desired state of Product's Backend Usages
type BackendUsageSpec struct {
	Path string `json:"path,omitempty"`
}

// SecuritySpec defines the desired state of Authentication Security
type SecuritySpec struct {
	// HostHeader Lets you define a custom Host request header. This is needed if your API backend only accepts traffic from a specific host.
	// +optional
	HostHeader *string `json:"hostHeader,omitempty"`

	// SecretToken Enables you to block any direct developer requests to your API backend;
	// each 3scale API gateway call to your API backend contains a request header called X-3scale-proxy-secret-token.
	// The value of this header can be set by you here. It's up to you ensure your backend only allows calls with this secret header.
	// +optional
	SecretToken *string `json:"hostHeader,omitempty"`
}

// AppKeyAppIDAuthenticationSpec defines the desired state of AppKey&AppId Authentication
type AppKeyAppIDAuthenticationSpec struct {
	// AppID is the name of the parameter that acts of behalf of app id
	// +optional
	AppID *string `json:"appID,omitempty"`

	// AppKey is the name of the parameter that acts of behalf of app key
	// +optional
	AppKey *string `json:"appKey,omitempty"`

	// Credentials Location available options:
	// header: As HTTP Headers
	// query: As query parameters (GET) or body parameters (POST/PUT/DELETE)
	// basic: As HTTP Basic Authentication
	// +optional
	// +kubebuilder:validation:Enum=header;query;basic
	CredentialsLocation *string `json:"credentials,omitempty"`

	// +optional
	Security *SecuritySpec `json:"security,omitempty"`

	// TODO GATEWAY RESPONSE
}

// UserKeyAuthenticationSpec defines the desired state of User Key Authentication
type UserKeyAuthenticationSpec struct {
	// +optional
	AuthUserKey *string `json:"authUserKey,omitempty"`

	// Credentials Location available options:
	// header: As HTTP Headers
	// query: As query parameters (GET) or body parameters (POST/PUT/DELETE)
	// basic: As HTTP Basic Authentication
	// +optional
	// +kubebuilder:validation:Enum=header;query;basic
	CredentialsLocation *string `json:"credentials,omitempty"`

	// +optional
	Security *SecuritySpec `json:"security,omitempty"`

	// TODO GATEWAY RESPONSE
}

// AuthenticationSpec defines the desired state of Product Authentication
type AuthenticationSpec struct {
	// +optional
	UserKeyAuthentication *UserKeyAuthenticationSpec `json:"userkey,omitempty"`
	// +optional
	AppKeyAppIDAuthentication *AppKeyAppIDAuthenticationSpec `json:"appKeyAppID,omitempty"`

	// TODO OpenID
}

// ApicastHostedSpec defines the desired state of Product Apicast Hosted
type ApicastHostedSpec struct {
	// +optional
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
}

// ApicastSelfManagedSpec defines the desired state of Product Apicast Self Managed
type ApicastSelfManagedSpec struct {
	// +optional
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
	// +optional
	// +kubebuilder:validation:Pattern=^https?:\/\/.*$
	StagingPublicBaseURL *string `json:"stagingPublicBaseURL,omitempty"`
	// +optional
	// +kubebuilder:validation:Pattern=^https?:\/\/.*$
	ProductionPublicBaseURL *string `json:"productionPublicBaseURL,omitempty"`
}

// ProductDeploymentSpec defines the desired state of Product Deployment
type ProductDeploymentSpec struct {
	// +optional
	ApicastHosted *ApicastHostedSpec `json:"apicastHosted,omitempty"`
	// +optional
	ApicastSelfManaged *ApicastSelfManagedSpec `json:"apicastSelfManaged,omitempty"`
}

// ProductSpec defines the desired state of Product
type ProductSpec struct {
	Name string `json:"name,omitempty"`
	// +optional
	SystemName string `json:"systemName,omitempty"`
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	Deployment *ProductDeploymentSpec `json:"deployment,omitempty"`
	// +optional
	MappingRules []MappingRuleSpec `json:"mappingRules,omitempty"`

	// Backend usage will be a map of
	// Map: backend_id -> BackendSpec
	// Having backend_id as the index, the structure ensures single backend is not used multiple times.
	// +optional
	BackendUsages map[string]BackendUsageSpec `json:"backendUsages,omitempty"`

	// Metrics and methods
	// Map: system_name -> MetricSpec or MethodSpec
	// system_name attr is unique for all metrics AND methods
	// In other words, if metric's system_name is A, there is no metric or method with system_name A.
	// +optional
	Metrics map[string]MetricSpec `json:"metrics,omitempty"`
	// +optional
	Methods map[string]Methodpec `json:"methods,omitempty"`

	// +optional
	// Application Plans
	// Map: system_name -> Application Plan Spec
	ApplicationPlans map[string]ApplicationPlanSpec `json:"applicationPlans,omitempty"`
}

// ProductStatus defines the observed state of Product
type ProductStatus struct {
	ID        int64  `json:"productId,omitempty"`
	State     string `json:"state,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`

	// In the event that there is a terminal problem reconciling the
	// product, both ErrorReason and ErrorMessage will be set. ErrorReason
	// will be populated with a succinct value suitable for machine
	// interpretation, while ErrorMessage will contain a more verbose
	// string suitable for logging and human consumption.
	//
	// These fields should not be set for transitive errors that a
	// controller faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Product's spec or the configuration of
	// the product controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the product controller, or the
	// responsible product controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Products
	// can be added as events to the Product object and/or logged in the
	// controller's output.
	// +optional
	ErrorReason *ProductStatusError `json:"errorReason,omitempty"`
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Product is the Schema for the products API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=products,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="3scale Product"
type Product struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProductSpec   `json:"spec,omitempty"`
	Status ProductStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProductList contains a list of Product
type ProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Product `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Product{}, &ProductList{})
}
