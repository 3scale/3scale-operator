package v1beta1

import (
	"encoding/json"
	"fmt"
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

// MetricMethodRefSpec defines method or metric reference
// Metric or method can optionally belong to used backends
type MetricMethodRefSpec struct {
	// SystemName identifies uniquely the metric or methods
	SystemName string `json:"systemName"`

	// BackendSystemName identifies uniquely the backend
	// Backend reference must be used by the product
	// +optional
	BackendSystemName *string `json:"backend,omitempty"`
}

func (m *MetricMethodRefSpec) String() string {
	backendPrefix := ""
	if m.BackendSystemName != nil {
		backendPrefix = fmt.Sprintf("%s.", *m.BackendSystemName)
	}
	return fmt.Sprintf("%s%s", backendPrefix, m.SystemName)
}

// LimitSpec defines the maximum value a metric can take on a contract before the user is no longer authorized to use resources.
// Once a limit has been passed in a given period, reject messages will be issued if the service is accessed under this contract.
type LimitSpec struct {
	// Limit Period
	// +kubebuilder:validation:Enum=eternity;year;month;week;day;hour;minute
	Period string `json:"period"`

	// Limit Value
	Value int `json:"value"`

	// Metric or Method Reference
	MetricMethodRef MetricMethodRefSpec `json:"metricMethodRef"`
}

// PricingRuleSpec defines the cost of each operation performed on an API.
// Multiple pricing rules on the same metric divide up the ranges of when a pricing rule applies.
type PricingRuleSpec struct {
	// Range From
	From int `json:"from"`

	// Range To
	To int `json:"to"`

	// Metric or Method Reference
	MetricMethodRef MetricMethodRefSpec `json:"metricMethodRef"`

	// Price per unit (USD)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d{2})?$`
	PricePerUnit string `json:"pricePerUnit"`
}

// ApplicationPlanSpec defines the desired state of Product's Application Plan
type ApplicationPlanSpec struct {
	// +optional
	Name *string `json:"name,omitempty"`

	// Set whether or not applications can be created on demand
	// or if approval is required from you before they are activated.
	// +optional
	AppsRequireApproval *bool `json:"appsRequireApproval,omitempty"`

	// Trial Period (days)
	// +kubebuilder:validation:Minimum=0
	// +optional
	TrialPeriod *int `json:"trialPeriod,omitempty"`

	// Setup fee (USD)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d{2})?$`
	// +optional
	SetupFee *string `json:"setupFee,omitempty"`

	// Cost per Month (USD)
	// +kubebuilder:validation:Pattern=`^\d+(\.\d{2})?$`
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

// MethodSpec defines the desired state of Product's Method
type MethodSpec struct {
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
	// +kubebuilder:validation:Enum=GET;HEAD;POST;PUT;DELETE;OPTIONS;TRACE;PATCH;CONNECT
	HTTPMethod      string `json:"httpMethod"`
	Pattern         string `json:"pattern"`
	MetricMethodRef string `json:"metricMethodRef"`
	Increment       int    `json:"increment"`
	// +optional
	Position *int `json:"position,omitempty"`
	// +optional
	Last *bool `json:"last,omitempty"`
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
	SecretToken *string `json:"secretToken,omitempty"`
}

func (s *SecuritySpec) SecuritySecretToken() *string {
	return s.SecretToken
}

func (s *SecuritySpec) HostRewrite() *string {
	return s.HostHeader
}

// AppKeyAppIDAuthenticationSpec defines the desired state of AppKey&AppId Authentication
type AppKeyAppIDAuthenticationSpec struct {
	// AppID is the name of the parameter that acts of behalf of app id
	// +optional
	AppID *string `json:"appID,omitempty"`

	// AppKey is the name of the parameter that acts of behalf of app key
	// +optional
	AppKey *string `json:"appKey,omitempty"`

	// CredentialsLoc available options:
	// headers: As HTTP Headers
	// query: As query parameters (GET) or body parameters (POST/PUT/DELETE)
	// authorization: As HTTP Basic Authentication
	// +optional
	// +kubebuilder:validation:Enum=headers;query;authorization
	CredentialsLoc *string `json:"credentials,omitempty"`

	// +optional
	Security *SecuritySpec `json:"security,omitempty"`

	// TODO GATEWAY RESPONSE
}

func (a *AppKeyAppIDAuthenticationSpec) SecuritySecretToken() *string {
	if a.Security == nil {
		return nil
	}

	return a.Security.SecuritySecretToken()
}

func (a *AppKeyAppIDAuthenticationSpec) HostRewrite() *string {
	if a.Security == nil {
		return nil
	}

	return a.Security.HostRewrite()
}

func (a *AppKeyAppIDAuthenticationSpec) CredentialsLocation() *string {
	return a.CredentialsLoc
}

func (a *AppKeyAppIDAuthenticationSpec) AuthAppID() *string {
	return a.AppID
}

func (a *AppKeyAppIDAuthenticationSpec) AuthAppKey() *string {
	return a.AppKey
}

// UserKeyAuthenticationSpec defines the desired state of User Key Authentication
type UserKeyAuthenticationSpec struct {
	// +optional
	Key *string `json:"authUserKey,omitempty"`

	// Credentials Location available options:
	// headers: As HTTP Headers
	// query: As query parameters (GET) or body parameters (POST/PUT/DELETE)
	// authorization: As HTTP Basic Authentication
	// +optional
	// +kubebuilder:validation:Enum=headers;query;authorization
	CredentialsLoc *string `json:"credentials,omitempty"`

	// +optional
	Security *SecuritySpec `json:"security,omitempty"`

	// TODO GATEWAY RESPONSE
}

func (u *UserKeyAuthenticationSpec) SecuritySecretToken() *string {
	if u.Security == nil {
		return nil
	}

	return u.Security.SecuritySecretToken()
}

func (u *UserKeyAuthenticationSpec) HostRewrite() *string {
	if u.Security == nil {
		return nil
	}

	return u.Security.HostRewrite()
}

func (u *UserKeyAuthenticationSpec) CredentialsLocation() *string {
	return u.CredentialsLoc
}

func (u *UserKeyAuthenticationSpec) AuthUserKey() *string {
	return u.Key
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

func (a *AuthenticationSpec) SecuritySecretToken() *string {
	// authentication is oneOf by CRD openapiV3 validation
	if a.UserKeyAuthentication != nil {
		return a.UserKeyAuthentication.SecuritySecretToken()
	}

	if a.AppKeyAppIDAuthentication == nil {
		panic("product authenticationspec: both userkey and appid_appkeyare nil")
	}

	return a.AppKeyAppIDAuthentication.SecuritySecretToken()
}

func (a *AuthenticationSpec) HostRewrite() *string {
	// authentication is oneOf by CRD openapiV3 validation
	if a.UserKeyAuthentication != nil {
		return a.UserKeyAuthentication.HostRewrite()
	}

	if a.AppKeyAppIDAuthentication == nil {
		panic("product authenticationspec: both userkey and appid_appkeyare nil")
	}

	return a.AppKeyAppIDAuthentication.HostRewrite()
}

func (a *AuthenticationSpec) CredentialsLocation() *string {
	// authentication is oneOf by CRD openapiV3 validation
	if a.UserKeyAuthentication != nil {
		return a.UserKeyAuthentication.CredentialsLocation()
	}

	if a.AppKeyAppIDAuthentication == nil {
		panic("product authenticationspec: both userkey and appid_appkeyare nil")
	}

	return a.AppKeyAppIDAuthentication.CredentialsLocation()
}

func (a *AuthenticationSpec) AuthUserKey() *string {
	// authentication is oneOf by CRD openapiV3 validation
	if a.UserKeyAuthentication != nil {
		return a.UserKeyAuthentication.AuthUserKey()
	}

	return nil
}

func (a *AuthenticationSpec) AuthAppID() *string {
	// authentication is oneOf by CRD openapiV3 validation
	if a.AppKeyAppIDAuthentication != nil {
		return a.AppKeyAppIDAuthentication.AuthAppID()
	}

	return nil
}

func (a *AuthenticationSpec) AuthAppKey() *string {
	// authentication is oneOf by CRD openapiV3 validation
	if a.AppKeyAppIDAuthentication != nil {
		return a.AppKeyAppIDAuthentication.AuthAppKey()
	}

	return nil
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

func (a *ApicastHostedSpec) SecuritySecretToken() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.SecuritySecretToken()
}

func (a *ApicastHostedSpec) HostRewrite() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.HostRewrite()
}

func (a *ApicastHostedSpec) CredentialsLocation() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.CredentialsLocation()
}

func (a *ApicastHostedSpec) AuthUserKey() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.AuthUserKey()
}

func (a *ApicastHostedSpec) AuthAppID() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.AuthAppID()
}

func (a *ApicastHostedSpec) AuthAppKey() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.AuthAppKey()
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

func (a *ApicastSelfManagedSpec) ProdPublicBaseURL() *string {
	return a.ProductionPublicBaseURL
}

func (a *ApicastSelfManagedSpec) StagPublicBaseURL() *string {
	return a.StagingPublicBaseURL
}

func (a *ApicastSelfManagedSpec) SecuritySecretToken() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.SecuritySecretToken()
}

func (a *ApicastSelfManagedSpec) HostRewrite() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.HostRewrite()
}

func (a *ApicastSelfManagedSpec) CredentialsLocation() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.CredentialsLocation()
}

func (a *ApicastSelfManagedSpec) AuthUserKey() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.AuthUserKey()
}

func (a *ApicastSelfManagedSpec) AuthAppID() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.AuthAppID()
}

func (a *ApicastSelfManagedSpec) AuthAppKey() *string {
	if a.Authentication == nil {
		return nil
	}
	return a.Authentication.AuthAppKey()
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

func (d *ProductDeploymentSpec) ProdPublicBaseURL() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		// Hosted deployment mode does not allow updating public base urls
		return nil
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	return d.ApicastSelfManaged.ProdPublicBaseURL()
}

func (d *ProductDeploymentSpec) StagingPublicBaseURL() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		// Hosted deployment mode does not allow updating public base urls
		return nil
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	return d.ApicastSelfManaged.StagPublicBaseURL()
}

func (d *ProductDeploymentSpec) SecuritySecretToken() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		return d.ApicastHosted.SecuritySecretToken()
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	return d.ApicastSelfManaged.SecuritySecretToken()
}

func (d *ProductDeploymentSpec) HostRewrite() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		return d.ApicastHosted.HostRewrite()
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	return d.ApicastSelfManaged.HostRewrite()
}

func (d *ProductDeploymentSpec) CredentialsLocation() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		return d.ApicastHosted.CredentialsLocation()
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	return d.ApicastSelfManaged.CredentialsLocation()
}

func (d *ProductDeploymentSpec) AuthUserKey() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		return d.ApicastHosted.AuthUserKey()
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	return d.ApicastSelfManaged.AuthUserKey()
}

func (d *ProductDeploymentSpec) AuthAppID() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		return d.ApicastHosted.AuthAppID()
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	return d.ApicastSelfManaged.AuthAppID()
}

func (d *ProductDeploymentSpec) AuthAppKey() *string {
	// spec.deployment is oneOf by CRD openapiV3 validation
	if d.ApicastHosted != nil {
		return d.ApicastHosted.AuthAppKey()
	}

	if d.ApicastSelfManaged == nil {
		panic("product spec.deployment apicasthosted and selfmanaged are nil")
	}

	return d.ApicastSelfManaged.AuthAppKey()
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

	// Description is a human readable text of the product
	// +optional
	Description string `json:"description,omitempty"`

	// Deployment defined 3scale product deployment mode
	// +optional
	Deployment *ProductDeploymentSpec `json:"deployment,omitempty"`

	// Mapping Rules
	// Array: MappingRule Spec
	// +optional
	MappingRules []MappingRuleSpec `json:"mappingRules,omitempty"`

	// Backend usage will be a map of
	// Map: system_name -> BackendUsageSpec
	// Having system_name as the index, the structure ensures one backend is not used multiple times.
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
	Methods map[string]MethodSpec `json:"methods,omitempty"`

	// Application Plans
	// Map: system_name -> Application Plan Spec
	// +optional
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

func (s *ProductSpec) ProdPublicBaseURL() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.ProdPublicBaseURL()
}

func (s *ProductSpec) StagingPublicBaseURL() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.StagingPublicBaseURL()
}

func (s *ProductSpec) SecuritySecretToken() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.SecuritySecretToken()
}

func (s *ProductSpec) HostRewrite() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.HostRewrite()
}

func (s *ProductSpec) CredentialsLocation() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.CredentialsLocation()
}

func (s *ProductSpec) AuthUserKey() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.AuthUserKey()
}

func (s *ProductSpec) AuthAppID() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.AuthAppID()
}

func (s *ProductSpec) AuthAppKey() *string {
	if s.Deployment == nil {
		return nil
	}
	return s.Deployment.AuthAppKey()
}

// ProductStatus defines the observed state of Product
// +k8s:openapi-gen=true
type ProductStatus struct {
	// +optional
	ID *int64 `json:"productId,omitempty"`
	// +optional
	State *string `json:"state,omitempty"`

	// 3scale control plane host
	// +optional
	ProviderAccountHost string `json:"providerAccountHost,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Product Spec.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Current state of the 3scale product.
	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions common.Conditions `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

func (p *ProductStatus) Equals(other *ProductStatus, logger logr.Logger) bool {
	if !reflect.DeepEqual(p.ID, other.ID) {
		diff := cmp.Diff(p.ID, other.ID)
		logger.V(1).Info("ID not equal", "difference", diff)
		return false
	}

	if !reflect.DeepEqual(p.State, other.State) {
		diff := cmp.Diff(p.State, other.State)
		logger.V(1).Info("State not equal", "difference", diff)
		return false
	}

	if p.ProviderAccountHost != other.ProviderAccountHost {
		diff := cmp.Diff(p.ProviderAccountHost, other.ProviderAccountHost)
		logger.V(1).Info("ProviderAccountHost not equal", "difference", diff)
		return false
	}

	if p.ObservedGeneration != other.ObservedGeneration {
		diff := cmp.Diff(p.ObservedGeneration, other.ObservedGeneration)
		logger.V(1).Info("ObservedGeneration not equal", "difference", diff)
		return false
	}

	// Marshalling sorts by condition type
	currentMarshaledJSON, _ := p.Conditions.MarshalJSON()
	otherMarshaledJSON, _ := other.Conditions.MarshalJSON()
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		diff := cmp.Diff(string(currentMarshaledJSON), string(otherMarshaledJSON))
		logger.V(1).Info("Conditions not equal", "difference", diff)
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

func (product *Product) SetDefaults(logger logr.Logger) bool {
	updated := false

	// Respect 3scale API defaults
	if product.Spec.SystemName == "" {
		product.Spec.SystemName = productSystemNameRegexp.ReplaceAllString(product.Spec.Name, "")
		updated = true
	}

	// 3scale API ignores case of the system name field
	systemNameLowercase := strings.ToLower(product.Spec.SystemName)
	if product.Spec.SystemName != systemNameLowercase {
		logger.Info("System name updated", "from", product.Spec.SystemName, "to", systemNameLowercase)
		product.Spec.SystemName = systemNameLowercase
		updated = true
	}

	if product.Spec.Metrics == nil {
		product.Spec.Metrics = map[string]MetricSpec{}
		updated = true
	}

	// Hits metric
	hitsFound := false
	for systemName := range product.Spec.Metrics {
		if systemName == "hits" {
			hitsFound = true
		}
	}
	if !hitsFound {
		logger.Info("Hits metric added")
		product.Spec.Metrics["hits"] = MetricSpec{
			Name:        "Hits",
			Unit:        "hit",
			Description: "Number of API hits",
		}
		updated = true
	}

	return updated
}

func (product *Product) Validate() field.ErrorList {
	errors := field.ErrorList{}
	specFldPath := field.NewPath("spec")
	metricsFldPath := specFldPath.Child("metrics")
	mappingRulesFldPath := specFldPath.Child("mappingRules")
	applicationPlansFldPath := specFldPath.Child("applicationPlans")
	methodsFldPath := specFldPath.Child("methods")

	// check hits metric exists
	if len(product.Spec.Metrics) == 0 {
		errors = append(errors, field.Required(metricsFldPath, "Product spec does not allow empty metrics."))
	} else {
		if _, ok := product.Spec.Metrics["hits"]; !ok {
			errors = append(errors, field.Invalid(metricsFldPath, nil, "metrics map not valid for Product. 'hits' metric must exist."))
		}
	}

	metricSystemNameMap := map[string]interface{}{}
	// Check metric systemNames are unique for all metric and methods
	for systemName := range product.Spec.Metrics {
		if _, ok := metricSystemNameMap[systemName]; ok {
			metricIdxFldPath := metricsFldPath.Key(systemName)
			errors = append(errors, field.Invalid(metricIdxFldPath, product.Spec.Metrics[systemName], "metric system_name not unique."))
		} else {
			metricSystemNameMap[systemName] = nil
		}
	}
	// Check method systemNames are unique for all metric and methods
	for systemName := range product.Spec.Methods {
		if _, ok := metricSystemNameMap[systemName]; ok {
			methodIdxFldPath := methodsFldPath.Key(systemName)
			errors = append(errors, field.Invalid(methodIdxFldPath, product.Spec.Methods[systemName], "method system_name not unique."))
		} else {
			metricSystemNameMap[systemName] = nil
		}
	}

	metricFirendlyNameMap := map[string]interface{}{}
	// Check metric names are unique for all metric and methods
	for systemName, metricSpec := range product.Spec.Metrics {
		if _, ok := metricFirendlyNameMap[metricSpec.Name]; ok {
			metricIdxFldPath := metricsFldPath.Key(systemName)
			errors = append(errors, field.Invalid(metricIdxFldPath, metricSpec, "metric name not unique."))
		} else {
			metricFirendlyNameMap[systemName] = nil
		}
	}
	// Check method names are unique for all metric and methods
	for systemName, methodSpec := range product.Spec.Methods {
		if _, ok := metricFirendlyNameMap[methodSpec.Name]; ok {
			methodIdxFldPath := methodsFldPath.Key(systemName)
			errors = append(errors, field.Invalid(methodIdxFldPath, methodSpec, "method name not unique."))
		} else {
			metricFirendlyNameMap[systemName] = nil
		}
	}

	// Check mapping rules metrics and method refs exists
	for idx, spec := range product.Spec.MappingRules {
		if !product.FindMetricOrMethod(spec.MetricMethodRef) {
			mappingRulesIdxFldPath := mappingRulesFldPath.Index(idx)
			errors = append(errors, field.Invalid(mappingRulesIdxFldPath, spec.MetricMethodRef, "mappingrule does not have valid metric or method reference."))
		}
	}

	// Check application plan limits local metricOrMethod ref exists
	for planSystemName, planSpec := range product.Spec.ApplicationPlans {
		planFldPath := applicationPlansFldPath.Key(planSystemName)
		limitsFldPath := planFldPath.Child("limits")
		for idx, limitSpec := range planSpec.Limits {
			// Only local references
			if limitSpec.MetricMethodRef.BackendSystemName == nil && !product.FindMetricOrMethod(limitSpec.MetricMethodRef.SystemName) {
				limitFldPath := limitsFldPath.Index(idx)
				metricRefFldPath := limitFldPath.Child("metricMethodRef")
				errors = append(errors, field.Invalid(metricRefFldPath, limitSpec.MetricMethodRef.SystemName, "limit does not have valid local metric or method reference."))
			}
		}
	}

	// Check application plan limits keys (periods, metric) are unique
	for planSystemName, planSpec := range product.Spec.ApplicationPlans {
		planFldPath := applicationPlansFldPath.Key(planSystemName)
		limitsFldPath := planFldPath.Child("limits")
		periods := map[string]interface{}{}
		for idx, limitSpec := range planSpec.Limits {
			key := fmt.Sprintf("%s:%s", limitSpec.Period, limitSpec.MetricMethodRef.String())
			if _, ok := periods[key]; ok {
				limitFldPath := limitsFldPath.Index(idx)
				errors = append(errors, field.Invalid(limitFldPath, key, "limit period is not unique for the same metric."))
			} else {
				periods[key] = nil
			}
		}
	}

	// Check application plan pricing rule local metricOrMethod ref exists
	for planSystemName, planSpec := range product.Spec.ApplicationPlans {
		planFldPath := applicationPlansFldPath.Key(planSystemName)
		rulesFldPath := planFldPath.Child("pricingRules")
		for idx, pruleSpec := range planSpec.PricingRules {
			// Only local references
			if pruleSpec.MetricMethodRef.BackendSystemName == nil && !product.FindMetricOrMethod(pruleSpec.MetricMethodRef.SystemName) {
				ruleFldPath := rulesFldPath.Index(idx)
				metricRefFldPath := ruleFldPath.Child("metricMethodRef")
				errors = append(errors, field.Invalid(metricRefFldPath, pruleSpec.MetricMethodRef.SystemName, "Pricing rule does not have valid local metric or method reference."))
			}
		}
	}

	// Check application plan pricing rules From < To
	for planSystemName, planSpec := range product.Spec.ApplicationPlans {
		planFldPath := applicationPlansFldPath.Key(planSystemName)
		rulesFldPath := planFldPath.Child("pricingRules")
		for idx, ruleSpec := range planSpec.PricingRules {
			if ruleSpec.From > ruleSpec.To {
				ruleFldPath := rulesFldPath.Index(idx)
				bytes, _ := json.Marshal(ruleSpec)
				errors = append(errors, field.Invalid(ruleFldPath, string(bytes), "'To' value cannot be less than your 'From' value."))
			}
		}
	}

	// Check application plan pricing rules ranges are not overlapping
	for planSystemName, planSpec := range product.Spec.ApplicationPlans {
		planFldPath := applicationPlansFldPath.Key(planSystemName)
		rulesFldPath := planFldPath.Child("pricingRules")
		overlappedIndex := detectOverlappingPricingRuleRanges(planSpec.PricingRules)
		if overlappedIndex >= 0 {
			ruleFldPath := rulesFldPath.Index(overlappedIndex)
			bytes, _ := json.Marshal(planSpec.PricingRules[overlappedIndex])
			errors = append(errors, field.Invalid(ruleFldPath, string(bytes), "'From' value cannot be less than 'To' values of current rules for the same metric."))
		}
	}

	return errors
}

func (product *Product) IsSynced() bool {
	return product.Status.Conditions.IsTrueFor(ProductSyncedConditionType)
}

func (product *Product) FindMetricOrMethod(ref string) bool {
	if len(product.Spec.Metrics) > 0 {
		if _, ok := product.Spec.Metrics[ref]; ok {
			return true
		}
	}

	if len(product.Spec.Methods) > 0 {
		if _, ok := product.Spec.Methods[ref]; ok {
			return true
		}
	}

	return false
}

func detectOverlappingPricingRuleRanges(rules []PricingRuleSpec) int {
	rulesPerMetricMap := make(map[string][]PricingRuleSpec)
	for _, spec := range rules {
		key := spec.MetricMethodRef.String()
		rulesPerMetricMap[key] = append(rulesPerMetricMap[key], spec)
	}

	for _, rulesPerMetric := range rulesPerMetricMap {
		idx := isOverlappingRanges(rulesPerMetric)
		if idx >= 0 {
			ruleSpec := rulesPerMetric[idx]
			// search and return index
			for idx, spec := range rules {
				if spec == ruleSpec {
					return idx
				}
			}
		}
	}
	return -1
}

func isOverlappingRanges(rules []PricingRuleSpec) int {
	// Naive implementation: check rule X with all predecessors if there is overlapping
	if len(rules) < 2 {
		return -1
	}

	for a := 1; a < len(rules); a++ {
		for b := 0; b < a; b++ {
			// Assume for all rules From <= To
			// Let Condition A Mean that ruleRange A Completely After ruleRange B: true if ToB <= FromA
			// Let Condition B Mean that ruleRange A Completely Before ruleRange B: true if ToA <= FromB
			// Then Overlap exists if Neither A Nor B is true
			//  ~(A or B) <=> ~a and ~b
			// Thus: Overlap <=> ToB > FromA && ToA > FromB
			if rules[a].From < rules[b].To && rules[a].To > rules[b].From {
				return a
			}
		}
	}

	return -1
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
