package v1beta1

import (
	"reflect"
	"regexp"

	"github.com/3scale/3scale-operator/pkg/common"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ProductKind = "Product"

	// ProductInvalidConditionType represents that the combination of configuration in the ProductSpec
	// is not supported. This is not a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the ProductSpec references non existing internal Metric reference
	ProductInvalidConditionType common.ConditionType = "Invalid"

	// ProductOrphanConditionType represents that the configuration in the ProductSpec
	// contains reference to non existing resource.
	// This is (should be) a transient error, but
	// indicates a state that must be fixed before progress can be made.
	// Example: the ProductSpec references non existing backend resource
	ProductOrphanConditionType common.ConditionType = "Orphan"

	// ProductSyncedConditionType indicates the product has been successfully synchronized.
	// Steady state
	ProductSyncedConditionType common.ConditionType = "Synced"

	// ProductFailedConditionType indicates that an error occurred during synchronization.
	// The operator will retry.
	ProductFailedConditionType common.ConditionType = "Failed"
)

var (
	//
	productSystemNameRegexp = regexp.MustCompile("[^a-zA-Z0-9]+")
)

// LimitSpec define the maximum value a metric can take on a contract before the user is no longer authorized to use resources.
// Once a limit has been passed in a given period, reject messages will be issued if the service is accessed under this contract.
type LimitSpec struct {
	// +kubebuilder:validation:Enum=eternity;year;month;week;day;hour;minute
	Period          string `json:"period"`
	Value           int64  `json:"value"`
	MetricMethodRef string `json:"metricMethodRef"`
}

// PricingRuleSpec defines the desired state of Application Plan's Pricing Rule
type PricingRuleSpec struct {
	From            int    `json:"from"`
	To              int    `json:"to"`
	MetricMethodRef string `json:"metricMethodRef"`
	// Price per unit (USD)
	// +kubebuilder:validation:Pattern=`^\d+.?\d{2}$`
	PricePerUnit string `json:"pricePerUnit"`
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
	// +kubebuilder:validation:Pattern=`^\d+.?\d{2}$`
	// +optional
	SetupFee *string `json:"setupFee,omitempty"`

	// Cost per Month (USD)
	// +kubebuilder:validation:Pattern=`^\d+.?\d{2}$`
	// +optional
	CostMonth *string `json:"costMonth,omitempty"`

	// Pricing Rules
	// +optional
	PricingRules []PricingRuleSpec `json:"pricingRules,omitempty"`

	// Limits
	// +optional
	Limits []LimitSpec `json:"limits,omitempty"`

	// TODO Features
}

// Methodpec defines the desired state of Product's Method
type Methodpec struct {
	Name string `json:"friendlyName"`
	// +optional
	Description string `json:"description,omitempty"`
}

// MetricSpec defines the desired state of Product's Metric
type MetricSpec struct {
	Name string `json:"friendlyName"`
	Unit string `json:"unit"`
	// +optional
	Description string `json:"description,omitempty"`
}

// MappingRuleSpec defines the desired state of Product's MappingRule
type MappingRuleSpec struct {
	Verb            string `json:"verb"`
	Pattern         string `json:"pattern"`
	MetricMethodRef string `json:"metricMethodRef"`
	// +optional
	Increment *int `json:"increment,omitempty"`
	// +optional
	Position *int `json:"position,omitempty"`
}

// BackendUsageSpec defines the desired state of Product's Backend Usages
type BackendUsageSpec struct {
	Path string `json:"path"`
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

func (a *AuthenticationSpec) AuthenticationMode() string {
	// authentication is oneOf by CRD openapiV3 validation
	if a.UserKeyAuthentication != nil {
		return "1"
	}

	// must be appKey&appID
	return "2"
}

// ApicastHostedSpec defines the desired state of Product Apicast Hosted
type ApicastHostedSpec struct {
	// +optional
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
}

func (a *ApicastHostedSpec) AuthenticationMode() *string {
	if a.Authentication == nil {
		return nil
	}
	authenticationMode := a.Authentication.AuthenticationMode()
	return &authenticationMode
}

// ApicastSelfManagedSpec defines the desired state of Product Apicast Self Managed
type ApicastSelfManagedSpec struct {
	// +optional
	Authentication *AuthenticationSpec `json:"authentication,omitempty"`
	// +optional
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	StagingPublicBaseURL *string `json:"stagingPublicBaseURL,omitempty"`
	// +optional
	// +kubebuilder:validation:Pattern=`^https?:\/\/.*$`
	ProductionPublicBaseURL *string `json:"productionPublicBaseURL,omitempty"`
}

func (a *ApicastSelfManagedSpec) AuthenticationMode() *string {
	if a.Authentication == nil {
		return nil
	}
	authenticationMode := a.Authentication.AuthenticationMode()
	return &authenticationMode
}

// ProductDeploymentSpec defines the desired state of Product Deployment
type ProductDeploymentSpec struct {
	// +optional
	ApicastHosted *ApicastHostedSpec `json:"apicastHosted,omitempty"`
	// +optional
	ApicastSelfManaged *ApicastSelfManagedSpec `json:"apicastSelfManaged,omitempty"`
}

func (d *ProductDeploymentSpec) DeploymentOption() string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		return "hosted"
	}

	// must be self managed
	return "self_managed"
}

func (d *ProductDeploymentSpec) AuthenticationMode() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		return d.ApicastHosted.AuthenticationMode()
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	// must be self managed, a
	return d.ApicastSelfManaged.AuthenticationMode()
}

// ProductSpec defines the desired state of Product
// +k8s:openapi-gen=true
type ProductSpec struct {
	// Name is human readable name for the product
	Name string `json:"name"`

	// SystemName identifies uniquely the product within the account provider
	// Default value will be sanitized Name
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

	// +optional
	// Application Plans
	// Map: system_name -> Application Plan Spec
	ApplicationPlans map[string]ApplicationPlanSpec `json:"applicationPlans,omitempty"`

	// ProviderAccountRef references account provider credentials
	// +optional
	ProviderAccountRef *corev1.LocalObjectReference `json:"providerAccountRef,omitempty"`
}

func (s *ProductSpec) DeploymentOption() *string {
	if s.Deployment == nil {
		return nil
	}
	deploymentOption := s.Deployment.DeploymentOption()
	return &deploymentOption
}

func (s *ProductSpec) AuthenticationMode() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.AuthenticationMode()
}

// ProductStatus defines the observed state of Product
// +k8s:openapi-gen=true
type ProductStatus struct {
	// +optional
	ID *int64 `json:"productId,omitempty"`
	// +optional
	State *string `json:"state,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed MachineSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

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
	// TODO enum
	ErrorReason *string `json:"errorReason,omitempty"`
	// A human readable message indicating details about why the resource is in this condition.
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// Current state of the 3scale product.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (p *ProductStatus) Equals(other *ProductStatus) bool {
	if !reflect.DeepEqual(p.ID, other.ID) {
		return false
	}

	if !reflect.DeepEqual(p.State, other.State) {
		return false
	}

	if p.ObservedGeneration != other.ObservedGeneration {
		return false
	}

	if !reflect.DeepEqual(p.ErrorReason, other.ErrorMessage) {
		return false
	}

	// Marshalling sorts by condition type
	currentMarshaledJSON, _ := p.Conditions.MarshalJSON()
	otherMarshaledJSON, _ := other.Conditions.MarshalJSON()
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		return false
	}

	return true
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Product is the Schema for the products API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=products,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="3scale Product"
type Product struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProductSpec   `json:"spec,omitempty"`
	Status ProductStatus `json:"status,omitempty"`
}

func (product *Product) SetDefaults() bool {
	updated := false

	// Respect 3scale API defaults
	if product.Spec.SystemName == "" {
		product.Spec.SystemName = productSystemNameRegexp.ReplaceAllString(product.Spec.Name, "")
		updated = true
	}

	return updated
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
