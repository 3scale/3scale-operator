package v1alpha1

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/3scale/3scale-operator/pkg/helper"
	portaClient "github.com/3scale/3scale-porta-go-client/client"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// APISpec defines the desired state of API
// +k8s:openapi-gen=true
type APISpec struct {
	APIBase      `json:",inline"`
	APISelectors `json:",inline"`
}

type APIBase struct {
	Description       string            `json:"description"`
	IntegrationMethod IntegrationMethod `json:"integrationMethod"`
}

type APISelectors struct {
	// +optional
	PlanSelector *metav1.LabelSelector `json:"planSelector,omitempty"`
	// +optional
	MetricSelector *metav1.LabelSelector `json:"metricSelector,omitempty"`
}

// APIStatus defines the observed state of API
// +k8s:openapi-gen=true
type APIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// API is the Schema for the apis API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=apis,scope=Namespaced
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
	// +optional
	ApicastOnPrem *ApicastOnPrem `json:"apicastOnPrem,omitempty"`
	// +optional
	CodePlugin *CodePlugin `json:"codePlugin,omitempty"`
	// +optional
	ApicastHosted *ApicastHosted `json:"apicastHosted,omitempty"`
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
func (api API) getInternalAPIfrom3scale(c *portaClient.ThreeScaleClient) (*InternalAPI, error) {

	service, err := getServiceFromInternalAPI(c, api.Name)
	if err != nil {
		return nil, err
	}
	proxyConfig, err := c.ReadProxy(service.ID)
	if err != nil {
		return nil, err
	}
	applicationPlans, err := c.ListAppPlanByServiceId(service.ID)

	// Initialize the InternalAPI with whatever info we have.
	internalAPI := InternalAPI{
		Name: service.Name,
		APIBaseInternal: APIBaseInternal{
			APIBase: APIBase{
				Description: service.Description,
			},
			IntegrationMethod: InternalIntegration{},
		},
		Metrics: nil,
		Plans:   nil,
	}

	var integrationCredentials IntegrationCredentials
	switch service.BackendVersion {

	case "2":
		// backend_version 2 -> APP_ID
		integrationCredentials = IntegrationCredentials{
			AppID: &AppID{
				AppIDParameterName:  proxyConfig.AuthAppID,
				AppKeyParameterName: proxyConfig.AuthAppKey,
				CredentialsLocation: proxyConfig.CredentialsLocation,
			},
		}

	case "oidc":
		// backend_version oidc -> OpenIDConnector
		integrationCredentials = IntegrationCredentials{
			OpenIDConnector: &OpenIDConnector{
				Issuer:              proxyConfig.OidcIssuerEndpoint,
				CredentialsLocation: proxyConfig.CredentialsLocation,
			},
		}

	default:
		// backend_version 1 -> APIKey
		integrationCredentials = IntegrationCredentials{
			APIKey: &APIKey{
				AuthParameterName:   proxyConfig.AuthUserKey,
				CredentialsLocation: proxyConfig.CredentialsLocation,
			},
		}
	}

	authFailedResponseCode, err := strconv.ParseInt(proxyConfig.ErrorStatusAuthFailed, 10, 64)
	if err != nil {
		authFailedResponseCode = 403
	}
	authMissingResponseCode, err := strconv.ParseInt(proxyConfig.ErrorStatusAuthMissing, 10, 64)
	if err != nil {
		authMissingResponseCode = 403
	}

	errorsConfigs := Errors{
		AuthenticationFailed: Authentication{
			ResponseCode: authFailedResponseCode,
			ContentType:  proxyConfig.ErrorHeadersAuthFailed,
			ResponseBody: proxyConfig.ErrorAuthFailed,
		},
		AuthenticationMissing: Authentication{
			ResponseCode: authMissingResponseCode,
			ContentType:  proxyConfig.ErrorHeadersAuthMissing,
			ResponseBody: proxyConfig.ErrorAuthMissing,
		},
	}

	// Let's find out the integrationMethod details of the service.
	switch service.DeploymentOption {

	case "self_managed":
		// This is ApicastOnPrem for us.
		mappingRules, _ := getServiceMappingRulesFrom3scale(c, service)

		internalAPI.APIBaseInternal.IntegrationMethod = InternalIntegration{
			ApicastOnPrem: &InternalApicastOnPrem{
				APIcastBaseOptions: APIcastBaseOptions{
					PrivateBaseURL:    proxyConfig.ApiBackend,
					APITestGetRequest: proxyConfig.ApiTestPath,
					AuthenticationSettings: ApicastAuthenticationSettings{
						HostHeader:  proxyConfig.HostnameRewrite,
						SecretToken: proxyConfig.SecretToken,
						Credentials: integrationCredentials,
						Errors:      errorsConfigs,
					},
				},
				StagingPublicBaseURL:    proxyConfig.SandboxEndpoint,
				ProductionPublicBaseURL: proxyConfig.Endpoint,
				MappingRules:            *mappingRules,
			},
		}

	case "service_mesh_istio":
		//This is ServiceMeshIstio for us. TODO: Implement service_mesh_istio.

	case "hosted":
		// This is ApicastHosted for us.
		mappingRules, _ := getServiceMappingRulesFrom3scale(c, service)

		internalAPI.APIBaseInternal.IntegrationMethod = InternalIntegration{
			ApicastHosted: &InternalApicastHosted{
				APIcastBaseOptions: APIcastBaseOptions{
					PrivateBaseURL:    proxyConfig.ApiBackend,
					APITestGetRequest: proxyConfig.ApiTestPath,
					AuthenticationSettings: ApicastAuthenticationSettings{
						HostHeader:  proxyConfig.HostnameRewrite,
						SecretToken: proxyConfig.SecretToken,
						Credentials: integrationCredentials,
						Errors:      errorsConfigs,
					},
				},
				MappingRules: *mappingRules,
			},
		}

	case "plugin_ruby", "plugin_python", "plugin_rest", "plugin_java", "plugin_php", "plugin_csharp":
		// This is CodePlugin for us.
		internalAPI.APIBaseInternal.IntegrationMethod = InternalIntegration{
			CodePlugin: &InternalCodePlugin{
				AuthenticationSettings: CodePluginAuthenticationSettings{
					Credentials: integrationCredentials,
				},
			},
		}

	default:
		return nil, fmt.Errorf("invalid_deployment")
	}

	// Grab the metrics from 3scale.
	for _, metric := range service.Metrics.Metrics {
		internalMetric := InternalMetric{
			Name:        metric.FriendlyName,
			Unit:        metric.Unit,
			Description: metric.Description,
		}
		if strings.ToLower(internalMetric.Name) != "hits" {
			internalAPI.Metrics = append(internalAPI.Metrics, internalMetric)
		}
	}

	for _, applicationPlan := range applicationPlans.Plans {

		trialPeriodDays, _ := strconv.ParseInt(applicationPlan.TrialPeriodDays, 10, 64)
		approvalRequired, _ := strconv.ParseBool(applicationPlan.ApprovalRequired)
		setupFee := applicationPlan.SetupFee
		costMonth := applicationPlan.CostPerMonth

		internalPlan := InternalPlan{
			Name:             applicationPlan.PlanName,
			TrialPeriodDays:  trialPeriodDays,
			ApprovalRequired: approvalRequired,
			Default:          applicationPlan.Default,
			Costs: PlanCost{
				SetupFee:  setupFee,
				CostMonth: costMonth,
			},
			Limits: nil,
		}

		limits, _ := c.ListLimitsPerAppPlan(applicationPlan.ID)
		for _, limit := range limits.Limits {
			maxValue, _ := strconv.ParseInt(limit.Value, 10, 64)
			metric, _ := metricIDtoMetric(c, service.ID, limit.MetricID)
			internalLimit := InternalLimit{
				Name:     limit.XMLName.Local,
				Period:   limit.Period,
				MaxValue: maxValue,
				Metric:   metric.FriendlyName,
			}
			internalPlan.Limits = append(internalPlan.Limits, internalLimit)
		}

		internalAPI.Plans = append(internalAPI.Plans, internalPlan)
	}

	return &internalAPI, nil
}

func (api API) GetInternalAPI(c client.Client) (*InternalAPI, error) {

	internalAPI := InternalAPI{
		Name: api.Name,
		APIBaseInternal: APIBaseInternal{
			APIBase: APIBase{
				Description: api.Spec.Description,
			},
		},
	}
	//Get Metrics for each API
	metrics, err := getMetrics(api.Namespace, api.Spec.MetricSelector.MatchLabels, c)
	if err != nil && errors.IsNotFound(err) {
		// Nothing has been found
		log.Printf("No metrics found for: %s\n", api.Name)
	} else if err != nil {
		// Something is broken
		return nil, err
	}

	for _, metric := range metrics.Items {
		internalMetric := newInternalMetricFromMetric(metric)
		internalAPI.Metrics = append(internalAPI.Metrics, *internalMetric)
	}

	//Get Plans for each API
	plans, err := getPlans(api.Namespace, api.Spec.PlanSelector.MatchLabels, c)

	if err != nil && errors.IsNotFound(err) {
		// Nothing has been found
		log.Printf("No plans found for: %s\n", api.Name)
	} else if err != nil {
		// Something is broken
		return nil, err
	}
	// Let's do our job.
	for _, plan := range plans.Items {
		internalPlan, err := newInternalPlanFromPlan(plan, c)
		if err != nil {
			return nil, err
		}
		internalAPI.Plans = append(internalAPI.Plans, *internalPlan)
	}
	switch api.getIntegrationMethodType() {
	case "ApicastHosted":
		internalApicastHosted, err := newInternalApicastHostedFromApicastHosted(api.Namespace, *api.Spec.IntegrationMethod.ApicastHosted, c)
		if err != nil {
			return nil, err
		}
		internalAPI.IntegrationMethod.ApicastHosted = internalApicastHosted

	case "ApicastOnPrem":
		internalApicastOnPrem, err := newInternalApicastOnPremFromApicastOnPrem(api.Namespace, *api.Spec.IntegrationMethod.ApicastOnPrem, c)
		if err != nil {
			return nil, err
		}
		internalAPI.IntegrationMethod.ApicastOnPrem = internalApicastOnPrem

		// TODO: Policies
	case "CodePlugin":
		internalCodePlugin := InternalCodePlugin{
			AuthenticationSettings: CodePluginAuthenticationSettings{
				Credentials: api.Spec.IntegrationMethod.CodePlugin.AuthenticationSettings.Credentials,
			},
		}
		internalAPI.IntegrationMethod.CodePlugin = &internalCodePlugin
	default:
		return nil, fmt.Errorf("Not supported integration method")
	}
	return &internalAPI, nil
}

type ApicastHosted struct {
	APIcastBaseOptions   `json:",inline"`
	APIcastBaseSelectors `json:",inline"`
}
type InternalApicastHosted struct {
	APIcastBaseOptions
	MappingRules []InternalMappingRule `json:"mappingRules"`
}

func (i *InternalApicastHosted) GetCredentialTypeName() string {
	if i.AuthenticationSettings.Credentials.OpenIDConnector != nil {
		return "OpenIDConnector"
	} else if i.AuthenticationSettings.Credentials.APIKey != nil {
		return "APIKey"
	} else if i.AuthenticationSettings.Credentials.AppID != nil {
		return "AppID"
	}
	return ""
}
func (i *InternalApicastHosted) GetMappingRules() []InternalMappingRule {
	return i.MappingRules
}

type APIcastBaseOptions struct {
	PrivateBaseURL         string                        `json:"privateBaseURL"`
	APITestGetRequest      string                        `json:"apiTestGetRequest"`
	AuthenticationSettings ApicastAuthenticationSettings `json:"authenticationSettings"`
}

type APIcastBaseSelectors struct {
	// +optional
	MappingRulesSelector *metav1.LabelSelector `json:"mappingRulesSelector,omitempty"`
	// +optional
	PoliciesSelector *metav1.LabelSelector `json:"policiesSelector,omitempty"`
}

type ApicastOnPrem struct {
	APIcastBaseOptions      `json:",inline"`
	StagingPublicBaseURL    string `json:"stagingPublicBaseURL"`
	ProductionPublicBaseURL string `json:"productionPublicBaseURL"`
	APIcastBaseSelectors    `json:",inline"`
}
type InternalApicastOnPrem struct {
	APIcastBaseOptions
	StagingPublicBaseURL    string                `json:"stagingPublicBaseURL"`
	ProductionPublicBaseURL string                `json:"productionPublicBaseURL"`
	MappingRules            []InternalMappingRule `json:"mappingRules"`
}

func (i *InternalApicastOnPrem) GetCredentialTypeName() string {
	if i.AuthenticationSettings.Credentials.OpenIDConnector != nil {
		return "OpenIDConnector"

	} else if i.AuthenticationSettings.Credentials.APIKey != nil {
		return "APIKey"

	} else if i.AuthenticationSettings.Credentials.AppID != nil {
		return "AppID"
	}
	return ""
}
func (i *InternalApicastOnPrem) GetMappingRules() []InternalMappingRule {
	return i.MappingRules
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
	// +optional
	APIKey *APIKey `json:"apiKey,omitempty"`
	// +optional
	AppID *AppID `json:"appID,omitempty"`
	// +optional
	OpenIDConnector *OpenIDConnector `json:"openIDConnector,omitempty"`
}

type OpenIDConnector struct {
	Issuer              string `json:"issuer"`
	CredentialsLocation string `json:"credentialsLocation"`
}

type CodePlugin struct {
	AuthenticationSettings CodePluginAuthenticationSettings `json:"authenticationSettings"`
}
type InternalCodePlugin struct {
	AuthenticationSettings CodePluginAuthenticationSettings `json:"authenticationSettings"`
}

func (i *InternalCodePlugin) GetCredentialTypeName() string {
	if i.AuthenticationSettings.Credentials.OpenIDConnector != nil {
		return "OpenIDConnector"
	} else if i.AuthenticationSettings.Credentials.APIKey != nil {
		return "APIKey"

	} else if i.AuthenticationSettings.Credentials.AppID != nil {
		return "AppID"
	}
	return ""

}
func (i *InternalCodePlugin) GetMappingRules() []InternalMappingRule {
	return []InternalMappingRule{}
}

type CodePluginAuthenticationSettings struct {
	Credentials IntegrationCredentials `json:"credentials"`
}

var CredentialTypeToBackendVersion = map[string]string{
	"OpenIDConnector": "oidc",
	"AppID":           "2",
	"APIKey":          "1",
}

var IntegrationMethodToDeploymentType = map[string]string{
	"ApicastHosted": "hosted",
	"ApicastOnPrem": "self_managed",
	"CodePlugin":    "plugin_rest",
}

type InternalAPI struct {
	Name            string `json:"name"`
	APIBaseInternal `json:",omitempty"`
	Metrics         []InternalMetric `json:"metrics,omitempty"`
	Plans           []InternalPlan   `json:"Plans,omitempty"`
}

// sort sorts an API struct.
func (api *InternalAPI) sort() {

	sort.Slice(api.Metrics, func(i, j int) bool {
		if api.Metrics[i].Name != api.Metrics[j].Name {
			return api.Metrics[i].Name < api.Metrics[j].Name
		} else {
			return api.Metrics[i].Unit < api.Metrics[j].Unit
		}
	})

	sort.Slice(api.Plans, func(i, j int) bool {
		if api.Plans[i].Name != api.Plans[j].Name {
			return api.Plans[i].Name < api.Plans[j].Name
		} else {
			return api.Plans[i].TrialPeriodDays < api.Plans[j].TrialPeriodDays
		}
	})

	if api.IntegrationMethod.ApicastOnPrem != nil {
		sort.Slice(api.IntegrationMethod.ApicastOnPrem.MappingRules, func(i, j int) bool {
			if api.IntegrationMethod.ApicastOnPrem.MappingRules[i].Name != api.IntegrationMethod.ApicastOnPrem.MappingRules[j].Name {
				return api.IntegrationMethod.ApicastOnPrem.MappingRules[i].Name < api.IntegrationMethod.ApicastOnPrem.MappingRules[j].Name
			} else {
				return api.IntegrationMethod.ApicastOnPrem.MappingRules[i].Metric < api.IntegrationMethod.ApicastOnPrem.MappingRules[j].Metric
			}
		})
	}

	if api.IntegrationMethod.ApicastHosted != nil {
		sort.Slice(api.IntegrationMethod.ApicastHosted.MappingRules, func(i, j int) bool {
			if api.IntegrationMethod.ApicastHosted.MappingRules[i].Name != api.IntegrationMethod.ApicastHosted.MappingRules[j].Name {
				return api.IntegrationMethod.ApicastHosted.MappingRules[i].Name < api.IntegrationMethod.ApicastHosted.MappingRules[j].Name
			} else {
				return api.IntegrationMethod.ApicastHosted.MappingRules[i].Metric < api.IntegrationMethod.ApicastHosted.MappingRules[j].Metric
			}
		})
	}

	for _, plan := range api.Plans {
		plan.Sort()
	}
}

// getIntegrationName returns the IntegrationMethod Name from an InternalAPI
func (api InternalAPI) getIntegrationName() string {
	deploymentOption := ""
	if api.IntegrationMethod.ApicastHosted != nil {
		deploymentOption = "ApicastHosted"
	} else if api.IntegrationMethod.ApicastOnPrem != nil {
		deploymentOption = "ApicastOnPrem"
	} else if api.IntegrationMethod.CodePlugin != nil {
		deploymentOption = "CodePlugin"
	}
	return deploymentOption
}

// getIntegration returns the Integration object from an InternalAPI
func (api InternalAPI) getIntegration() Integration {
	if api.IntegrationMethod.ApicastHosted != nil {
		return api.IntegrationMethod.ApicastHosted
	} else if api.IntegrationMethod.ApicastOnPrem != nil {
		return api.IntegrationMethod.ApicastOnPrem
	} else if api.IntegrationMethod.CodePlugin != nil {
		return api.IntegrationMethod.CodePlugin
	}
	return nil
}

// createIn3scale Creates the InternalAPI in 3scale
func (api InternalAPI) createIn3scale(c *portaClient.ThreeScaleClient) error {

	// Get the proper 3scale deployment Option based on the integrationMethod
	deploymentOption := IntegrationMethodToDeploymentType[api.getIntegrationName()]
	if deploymentOption == "" {
		return fmt.Errorf("unknown integration method")
	}

	// Get the proper backendVersion based on the CredentialType

	backendVersion := CredentialTypeToBackendVersion[api.getIntegration().GetCredentialTypeName()]
	if backendVersion == "" {
		return fmt.Errorf("invalid credential type method")
	}

	params := portaClient.Params{
		"description":       api.Description,
		"deployment_option": deploymentOption,
		"backend_version":   backendVersion,
	}

	service, err := c.CreateService(api.Name)
	if err != nil {
		return err
	}
	_, err = c.UpdateService(service.ID, params)
	if err != nil {
		return err
	}
	desiredProxy, err := get3scaleProxyFromInternalAPI(api)
	proxyParams := getProxyParamsFromProxy(desiredProxy, deploymentOption, backendVersion)

	_, err = c.UpdateProxy(service.ID, proxyParams)
	if err != nil {
		return err
	}

	// Promote config if needed
	productionProxy, _ := c.GetLatestProxyConfig(service.ID, "production")
	sandboxProxy, _ := c.GetLatestProxyConfig(service.ID, "sandbox")
	if productionProxy.ProxyConfig.Version != sandboxProxy.ProxyConfig.Version {
		_, err := c.PromoteProxyConfig(service.ID, "sandbox", strconv.Itoa(sandboxProxy.ProxyConfig.Version), "production")
		if err != nil {
			return err
		}
	}

	for _, metric := range api.Metrics {
		_, err := c.CreateMetric(service.ID, metric.Name, metric.Description, metric.Unit)
		if err != nil {
			return err
		}
	}

	defaultMappingRules, err := c.ListMappingRule(service.ID)
	if err != nil {
		return err
	}

	for _, defaultMappingRule := range defaultMappingRules.MappingRules {
		err := c.DeleteMappingRule(service.ID, defaultMappingRule.ID)
		if err != nil {
			return err
		}

	}

	for _, mappingRule := range api.getIntegration().GetMappingRules() {
		metric, err := metricNametoMetric(c, service.ID, mappingRule.Metric)
		if err != nil {
			return err
		}

		_, err = c.CreateMappingRule(service.ID, strings.ToUpper(mappingRule.Method), mappingRule.Path, int(mappingRule.Increment), metric.ID)
		if err != nil {
			return err
		}
	}

	for _, plan := range api.Plans {
		//TODO: expose publishing a plan from the CRD model

		plan3scale, err := c.CreateAppPlan(service.ID, plan.Name, "publish")
		if err != nil {
			return err
		}
		//TODO: add cancellation_period to application Plan
		params := portaClient.Params{
			"approval_required": strconv.FormatBool(plan.ApprovalRequired),
			"setup_fee": plan.Costs.SetupFee,
			"cost_per_month": plan.Costs.CostMonth,
			"trial_period_days": strconv.FormatInt(plan.TrialPeriodDays, 10),
		}
		_, err = c.UpdateAppPlan(service.ID, plan3scale.ID, plan3scale.PlanName, "", params)
		if err != nil {
			return err
		}

		if plan.Default {
			_, err = c.SetDefaultPlan(service.ID, plan3scale.ID)
		}

		for _, limit := range plan.Limits {
			metric, err := metricNametoMetric(c, service.ID, limit.Metric)
			if err != nil {
				return err
			}
			_, err = c.CreateLimitAppPlan(plan3scale.ID, metric.ID, limit.Period, int(limit.MaxValue))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteFrom3scale Removes an InternalAPI from 3scale
func (api InternalAPI) DeleteFrom3scale(c *portaClient.ThreeScaleClient) error {

	services, err := c.ListServices()
	if err != nil {
		return err
	}

	for _, service := range services.Services {
		if service.SystemName == api.Name {
			return c.DeleteService(service.ID)
		}
	}
	return nil
}

type APIBaseInternal struct {
	APIBase `json:",omitempty"`
	// We shadow the APIBase IntegrationMethod to point to our Internal representation
	IntegrationMethod InternalIntegration `json:"integrationMethod"`
}

type InternalIntegration struct {
	ApicastOnPrem *InternalApicastOnPrem `json:"apicastOnPrem"`
	CodePlugin    *InternalCodePlugin    `json:"codePlugin"`
	ApicastHosted *InternalApicastHosted `json:"apicastHosted"`
}

type Integration interface {
	GetMappingRules() []InternalMappingRule
	GetCredentialTypeName() string
}

// +k8s:openapi-gen=false
type InternalCredentials struct {
	AuthToken string `json:"token"`
	AdminURL  string `json:"adminURL"`
}

//TODO: Refactor Diffs.
type APIsDiff struct {
	MissingFromA []InternalAPI
	MissingFromB []InternalAPI
	Equal        []InternalAPI
	NotEqual     []APIPair
}
type APIPair struct {
	A InternalAPI
	B InternalAPI
}

// reconcileWith3scale creates/modifies/deletes APIs based on the information of the APIsDiff object.
func (d *APIsDiff) ReconcileWith3scale(creds InternalCredentials) error {

	c, err := helper.PortaClientFromURLString(creds.AdminURL, creds.AuthToken)

	if err != nil {
		return err
	}

	for _, api := range d.MissingFromB {

		err := api.createIn3scale(c)
		if err != nil {
			return err
		}
	}

	for _, api := range d.MissingFromA {
		err := api.DeleteFrom3scale(c)
		if err != nil {
			return err
		}
	}

	for _, apiPair := range d.NotEqual {

		serviceNeedsUpdate := false
		service, err := getServiceFromInternalAPI(c, apiPair.A.Name)
		if err != nil {
			return err
		}
		serviceParams := portaClient.Params{}

		// Check if DeploymentOption is correct
		desiredDeploymentOption := IntegrationMethodToDeploymentType[apiPair.A.getIntegrationName()]
		existingDeploymentOption := IntegrationMethodToDeploymentType[apiPair.B.getIntegrationName()]

		if desiredDeploymentOption != existingDeploymentOption {
			serviceNeedsUpdate = true
			serviceParams.AddParam("deployment_option", desiredDeploymentOption)
		}

		// Check if BackendVersion is correct
		desiredBackendVersion := CredentialTypeToBackendVersion[apiPair.A.getIntegration().GetCredentialTypeName()]
		existingBackendVersion := CredentialTypeToBackendVersion[apiPair.B.getIntegration().GetCredentialTypeName()]

		if desiredBackendVersion != existingBackendVersion {
			serviceNeedsUpdate = true
			serviceParams.AddParam("backend_version", desiredBackendVersion)
		}

		//Check if api description is different and mark it for update
		if apiPair.A.Description != apiPair.B.Description {
			serviceNeedsUpdate = true
			serviceParams.AddParam("description", apiPair.A.Description)
		}

		// Update the service with the params
		if serviceNeedsUpdate {
			_, err := c.UpdateService(service.ID, serviceParams)
			if err != nil {
				return err
			}
		}

		desiredProxy, err := get3scaleProxyFromInternalAPI(apiPair.A)
		if err != nil {
			return err
		}
		existingProxy, err := get3scaleProxyFromInternalAPI(apiPair.B)
		if err != nil {
			return err
		}

		if desiredProxy != existingProxy {

			proxyParams := getProxyParamsFromProxy(desiredProxy, desiredDeploymentOption, desiredBackendVersion)

			_, err = c.UpdateProxy(service.ID, proxyParams)
			if err != nil {
				return err
			}
		}

		// Get the Difference in Metrics for the API
		metricsDiff := diffMetrics(apiPair.A.Metrics, apiPair.B.Metrics)
		err = metricsDiff.ReconcileWith3scale(c, service.ID, apiPair.A)
		if err != nil {
			return err
		}

		// reconcileWith3scale Mapping Rules
		mappingRulesDiff := diffMappingRules(apiPair.A.getIntegration().GetMappingRules(), apiPair.B.getIntegration().GetMappingRules())
		err = mappingRulesDiff.reconcileWith3scale(c, service.ID, apiPair.A)
		if err != nil {
			return err
		}

		// Because MappingRules are not Unique, let's remove duplicated mappingRules

		// reconcileWith3scale Plans
		plansDiff := diffPlans(apiPair.A.Plans, apiPair.B.Plans)
		err = plansDiff.reconcileWith3scale(c, service.ID, apiPair.A)
		if err != nil {
			return err
		}

		// Promote config if needed
		productionProxy, _ := c.GetLatestProxyConfig(service.ID, "production")
		sandboxProxy, _ := c.GetLatestProxyConfig(service.ID, "sandbox")
		if productionProxy.ProxyConfig.Version != sandboxProxy.ProxyConfig.Version {
			_, err := c.PromoteProxyConfig(service.ID, "sandbox", strconv.Itoa(sandboxProxy.ProxyConfig.Version), "production")
			if err != nil {
				return err
			}
		}

	}
	return nil

}

// DiffAPIs generate an APIsDiff object with equal, different and missing APIs from two InternalAPI slices.
func DiffAPIs(APIs1 []InternalAPI, APIs2 []InternalAPI) APIsDiff {

	var apisDiff APIsDiff

	if len(APIs2) == 0 {
		apisDiff.MissingFromB = APIs1
		return apisDiff
	}

	for i := 0; i < 2; i++ {
		for _, API1 := range APIs1 {
			found := false
			for _, API2 := range APIs2 {
				if API2.Name == API1.Name {
					if i == 0 {
						if CompareInternalAPI(API2, API1) {
							apisDiff.Equal = append(apisDiff.Equal, API1)
						} else {
							apiPair := APIPair{
								A: API1,
								B: API2,
							}
							apisDiff.NotEqual = append(apisDiff.NotEqual, apiPair)
						}
					}
					found = true
					break
				}
			}
			if !found {
				switch i {
				case 0:
					apisDiff.MissingFromB = append(apisDiff.MissingFromB, API1)
				case 1:
					apisDiff.MissingFromA = append(apisDiff.MissingFromA, API1)
				}
			}

		}
		if i == 0 {
			APIs1, APIs2 = APIs2, APIs1
		}
	}
	return apisDiff
}

// newInternalApicastHostedFromApicastHosted Creates an InteranlApicastHosted object from an ApicastHosted object
func newInternalApicastHostedFromApicastHosted(namespace string, hosted ApicastHosted, c client.Client) (*InternalApicastHosted, error) {

	internalApicastHosted := InternalApicastHosted{
		APIcastBaseOptions: APIcastBaseOptions{
			PrivateBaseURL:         hosted.PrivateBaseURL,
			APITestGetRequest:      hosted.APITestGetRequest,
			AuthenticationSettings: hosted.AuthenticationSettings,
		},
		MappingRules: nil,
	}
	// Get Mapping Rules
	mappingRules, err := getMappingRules(namespace, hosted.MappingRulesSelector.MatchLabels, c)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Error: %s", err)
	} else if err != nil {
		// Something is broken
		log.Printf("Error: %s", err)
		return nil, err
	} else {
		for _, mappingRule := range mappingRules.Items {
			internalMappingRule, err := newInternalMappingRuleFromMappingRule(mappingRule, c)
			if err != nil {
				log.Printf("mappingRule %s couldn't be converted", mappingRule.Name)
			} else {
				internalApicastHosted.MappingRules = append(internalApicastHosted.MappingRules, *internalMappingRule)
			}
		}
	}
	return &internalApicastHosted, nil
}

// newInternalApicastOnPremFromApicastOnPrem Creates an InteranlApicastOnPrem object from an ApicastOnPrem object
func newInternalApicastOnPremFromApicastOnPrem(namespace string, prem ApicastOnPrem, c client.Client) (*InternalApicastOnPrem, error) {
	internalApicastOnPrem := InternalApicastOnPrem{
		APIcastBaseOptions: APIcastBaseOptions{
			PrivateBaseURL:         prem.PrivateBaseURL,
			APITestGetRequest:      prem.APITestGetRequest,
			AuthenticationSettings: prem.AuthenticationSettings,
		},
		StagingPublicBaseURL:    prem.StagingPublicBaseURL,
		ProductionPublicBaseURL: prem.ProductionPublicBaseURL,
		MappingRules:            nil,
	}
	// Get Mapping Rules
	// api.Spec.IntegrationMethod.ApicastOnPrem.MappingRulesSelector
	mappingRules, err := getMappingRules(namespace, prem.MappingRulesSelector.MatchLabels, c)

	if err != nil && errors.IsNotFound(err) {
		// Nothing has been found
		log.Printf("Error: %s\n", err)
	} else if err != nil {
		// Something is broken
		return nil, err
	} else {
		for _, mappingRule := range mappingRules.Items {
			internalMappingRule, err := newInternalMappingRuleFromMappingRule(mappingRule, c)
			if err != nil {
				// TODO: UPDATE STATUS OF THE OBJECT
				log.Printf("mappingRule %s couldn't be converted", mappingRule.Name)
			} else {
				internalApicastOnPrem.MappingRules = append(
					internalApicastOnPrem.MappingRules,
					*internalMappingRule,
				)
			}
		}
	}

	return &internalApicastOnPrem, nil
}

// CompareInternalAPI Compares two InternalAPIs and return true or false.
func CompareInternalAPI(APIA, APIB InternalAPI) bool {
	for i := range APIA.Plans {
		for j := range APIA.Plans[i].Limits {
			APIA.Plans[i].Limits[j].Name = "limit"
		}
	}

	for i := range APIB.Plans {
		for j := range APIB.Plans[i].Limits {
			APIB.Plans[i].Limits[j].Name = "limit"
		}
	}

	if APIA.getIntegrationName() != APIB.getIntegrationName() {
		return false
	} else {
		switch APIA.getIntegrationName() {
		case "ApicastOnPrem":

			// Always set he port number, because porta adds it automatically and makes the sync fail.
			APIA.IntegrationMethod.ApicastOnPrem.ProductionPublicBaseURL = helper.SetURLDefaultPort(APIA.IntegrationMethod.ApicastOnPrem.ProductionPublicBaseURL)
			APIA.IntegrationMethod.ApicastOnPrem.StagingPublicBaseURL = helper.SetURLDefaultPort(APIA.IntegrationMethod.ApicastOnPrem.StagingPublicBaseURL)

			APIB.IntegrationMethod.ApicastOnPrem.ProductionPublicBaseURL = helper.SetURLDefaultPort(APIB.IntegrationMethod.ApicastOnPrem.ProductionPublicBaseURL)
			APIB.IntegrationMethod.ApicastOnPrem.StagingPublicBaseURL = helper.SetURLDefaultPort(APIB.IntegrationMethod.ApicastOnPrem.StagingPublicBaseURL)

			for i := range APIA.IntegrationMethod.ApicastOnPrem.MappingRules {
				APIA.IntegrationMethod.ApicastOnPrem.MappingRules[i].Name = "mapping_rule"
			}
			for i := range APIB.IntegrationMethod.ApicastOnPrem.MappingRules {
				APIB.IntegrationMethod.ApicastOnPrem.MappingRules[i].Name = "mapping_rule"
			}
		case "ApicastHosted":
			for i := range APIA.IntegrationMethod.ApicastHosted.MappingRules {
				APIA.IntegrationMethod.ApicastHosted.MappingRules[i].Name = "mapping_rule"
			}
			for i := range APIB.IntegrationMethod.ApicastHosted.MappingRules {
				APIB.IntegrationMethod.ApicastHosted.MappingRules[i].Name = "mapping_rule"
			}
		}
	}

	A, _ := json.Marshal(APIA)
	B, _ := json.Marshal(APIB)

	return reflect.DeepEqual(A, B)
}

func get3scaleProxyFromInternalAPI(api InternalAPI) (portaClient.Proxy, error) {

	//TODO: Improve get3scaleProxyFromInternalAPI.
	var proxy portaClient.Proxy

	integration := api.IntegrationMethod
	if integration.ApicastHosted != nil {
		proxy.HostnameRewrite = integration.ApicastHosted.AuthenticationSettings.HostHeader
		proxy.ApiBackend = integration.ApicastHosted.PrivateBaseURL
		proxy.ApiTestPath = integration.ApicastHosted.APITestGetRequest
		proxy.ErrorStatusAuthFailed = strconv.FormatInt(integration.ApicastHosted.AuthenticationSettings.Errors.AuthenticationFailed.ResponseCode, 10)
		proxy.ErrorAuthFailed = integration.ApicastHosted.AuthenticationSettings.Errors.AuthenticationFailed.ResponseBody
		proxy.ErrorAuthMissing = integration.ApicastHosted.AuthenticationSettings.Errors.AuthenticationMissing.ResponseBody
		proxy.ErrorStatusAuthFailed = strconv.FormatInt(integration.ApicastHosted.AuthenticationSettings.Errors.AuthenticationFailed.ResponseCode, 10)
		proxy.ErrorHeadersAuthFailed = integration.ApicastHosted.AuthenticationSettings.Errors.AuthenticationFailed.ContentType
		proxy.ErrorStatusAuthMissing = strconv.FormatInt(integration.ApicastHosted.AuthenticationSettings.Errors.AuthenticationMissing.ResponseCode, 10)
		proxy.ErrorHeadersAuthMissing = integration.ApicastHosted.AuthenticationSettings.Errors.AuthenticationMissing.ContentType
		proxy.SecretToken = integration.ApicastHosted.AuthenticationSettings.SecretToken
		if integration.ApicastHosted.AuthenticationSettings.Credentials.OpenIDConnector != nil {
			proxy.CredentialsLocation = integration.ApicastHosted.AuthenticationSettings.Credentials.OpenIDConnector.CredentialsLocation
			proxy.OidcIssuerEndpoint = integration.ApicastHosted.AuthenticationSettings.Credentials.OpenIDConnector.Issuer
		} else if integration.ApicastHosted.AuthenticationSettings.Credentials.AppID != nil {
			proxy.CredentialsLocation = integration.ApicastHosted.AuthenticationSettings.Credentials.AppID.CredentialsLocation
			proxy.AuthAppID = integration.ApicastHosted.AuthenticationSettings.Credentials.AppID.AppIDParameterName
			proxy.AuthAppKey = integration.ApicastHosted.AuthenticationSettings.Credentials.AppID.AppKeyParameterName
		} else if integration.ApicastHosted.AuthenticationSettings.Credentials.APIKey != nil {
			proxy.CredentialsLocation = integration.ApicastHosted.AuthenticationSettings.Credentials.APIKey.CredentialsLocation
			proxy.AuthUserKey = integration.ApicastHosted.AuthenticationSettings.Credentials.APIKey.AuthParameterName
		}
	} else if integration.ApicastOnPrem != nil {
		proxy.HostnameRewrite = integration.ApicastOnPrem.AuthenticationSettings.HostHeader
		proxy.ApiBackend = integration.ApicastOnPrem.PrivateBaseURL
		proxy.ApiTestPath = integration.ApicastOnPrem.APITestGetRequest
		// TODO: FIX 3scale adds the port if missing, reconciliation will fail
		proxy.SandboxEndpoint = integration.ApicastOnPrem.StagingPublicBaseURL
		// TODO: FIX 3scale adds the port if missing, reconciliation will fail
		proxy.Endpoint = integration.ApicastOnPrem.ProductionPublicBaseURL
		proxy.ErrorStatusAuthFailed = strconv.FormatInt(integration.ApicastOnPrem.AuthenticationSettings.Errors.AuthenticationFailed.ResponseCode, 10)
		proxy.ErrorAuthFailed = integration.ApicastOnPrem.AuthenticationSettings.Errors.AuthenticationFailed.ResponseBody
		proxy.ErrorAuthMissing = integration.ApicastOnPrem.AuthenticationSettings.Errors.AuthenticationMissing.ResponseBody
		proxy.ErrorStatusAuthFailed = strconv.FormatInt(integration.ApicastOnPrem.AuthenticationSettings.Errors.AuthenticationFailed.ResponseCode, 10)
		proxy.ErrorHeadersAuthFailed = integration.ApicastOnPrem.AuthenticationSettings.Errors.AuthenticationFailed.ContentType
		proxy.ErrorStatusAuthMissing = strconv.FormatInt(integration.ApicastOnPrem.AuthenticationSettings.Errors.AuthenticationMissing.ResponseCode, 10)
		proxy.ErrorHeadersAuthMissing = integration.ApicastOnPrem.AuthenticationSettings.Errors.AuthenticationMissing.ContentType
		proxy.SecretToken = integration.ApicastOnPrem.AuthenticationSettings.SecretToken

		if integration.ApicastOnPrem.AuthenticationSettings.Credentials.OpenIDConnector != nil {
			proxy.CredentialsLocation = integration.ApicastOnPrem.AuthenticationSettings.Credentials.OpenIDConnector.CredentialsLocation
			proxy.OidcIssuerEndpoint = integration.ApicastOnPrem.AuthenticationSettings.Credentials.OpenIDConnector.Issuer

		} else if integration.ApicastOnPrem.AuthenticationSettings.Credentials.AppID != nil {
			proxy.CredentialsLocation = integration.ApicastOnPrem.AuthenticationSettings.Credentials.AppID.CredentialsLocation
			proxy.AuthAppID = integration.ApicastOnPrem.AuthenticationSettings.Credentials.AppID.AppIDParameterName
			proxy.AuthAppKey = integration.ApicastOnPrem.AuthenticationSettings.Credentials.AppID.AppKeyParameterName

		} else if integration.ApicastOnPrem.AuthenticationSettings.Credentials.APIKey != nil {
			proxy.CredentialsLocation = integration.ApicastOnPrem.AuthenticationSettings.Credentials.APIKey.CredentialsLocation
			proxy.AuthUserKey = integration.ApicastOnPrem.AuthenticationSettings.Credentials.APIKey.AuthParameterName
		}
	} else if integration.CodePlugin != nil {
		if integration.CodePlugin.AuthenticationSettings.Credentials.OpenIDConnector != nil {
			proxy.CredentialsLocation = integration.CodePlugin.AuthenticationSettings.Credentials.OpenIDConnector.CredentialsLocation
			proxy.OidcIssuerEndpoint = integration.CodePlugin.AuthenticationSettings.Credentials.OpenIDConnector.Issuer

		} else if integration.CodePlugin.AuthenticationSettings.Credentials.AppID != nil {
			proxy.CredentialsLocation = integration.CodePlugin.AuthenticationSettings.Credentials.AppID.CredentialsLocation
			proxy.AuthAppID = integration.CodePlugin.AuthenticationSettings.Credentials.AppID.AppIDParameterName
			proxy.AuthAppKey = integration.CodePlugin.AuthenticationSettings.Credentials.AppID.AppKeyParameterName

		} else if integration.CodePlugin.AuthenticationSettings.Credentials.APIKey != nil {
			proxy.CredentialsLocation = integration.CodePlugin.AuthenticationSettings.Credentials.APIKey.CredentialsLocation
			proxy.AuthUserKey = integration.CodePlugin.AuthenticationSettings.Credentials.APIKey.AuthParameterName
		}
	} else {
		return proxy, fmt.Errorf("integrationMethod invalid")
	}

	return proxy, nil
}
func getServiceFromInternalAPI(c *portaClient.ThreeScaleClient, serviceName string) (portaClient.Service, error) {
	services, err := c.ListServices()

	if err != nil {
		return portaClient.Service{}, err
	}

	for _, service := range services.Services {
		if service.SystemName == serviceName {
			return service, nil
		}
	}
	return portaClient.Service{}, fmt.Errorf("NotFound")
}
func getProxyParamsFromProxy(proxy portaClient.Proxy, deploymentOption, backendVersion string) portaClient.Params {

	// TODO: This is far from optimal, fix this
	proxyParams := portaClient.NewParams()

	switch deploymentOption {

	case "hosted":
		proxyParams.AddParam("hostname_rewrite", proxy.HostnameRewrite)
		proxyParams.AddParam("api_backend", proxy.ApiBackend)
		proxyParams.AddParam("api_test_path", proxy.ApiTestPath)
		proxyParams.AddParam("error_status_auth_failed", proxy.ErrorStatusAuthFailed)
		proxyParams.AddParam("error_auth_failed", proxy.ErrorAuthFailed)
		proxyParams.AddParam("error_auth_missing", proxy.ErrorAuthMissing)
		proxyParams.AddParam("error_status_auth_failed", proxy.ErrorStatusAuthFailed)
		proxyParams.AddParam("error_headers_auth_failed", proxy.ErrorHeadersAuthFailed)
		proxyParams.AddParam("error_status_auth_missing", proxy.ErrorStatusAuthMissing)
		proxyParams.AddParam("error_headers_auth_missing", proxy.ErrorHeadersAuthMissing)
		proxyParams.AddParam("secret_token", proxy.SecretToken)
	case "self_managed":
		proxyParams.AddParam("hostname_rewrite", proxy.HostnameRewrite)
		proxyParams.AddParam("api_backend", proxy.ApiBackend)
		proxyParams.AddParam("api_test_path", proxy.ApiTestPath)
		// TODO: FIX 3scale adds the port if missing, reconciliation will fail
		proxyParams.AddParam("sandbox_endpoint", proxy.SandboxEndpoint)
		proxyParams.AddParam("endpoint", proxy.Endpoint)
		proxyParams.AddParam("error_auth_failed", proxy.ErrorAuthFailed)
		proxyParams.AddParam("error_auth_missing", proxy.ErrorAuthMissing)
		proxyParams.AddParam("error_status_auth_failed", proxy.ErrorStatusAuthFailed)
		proxyParams.AddParam("error_headers_auth_failed", proxy.ErrorHeadersAuthFailed)
		proxyParams.AddParam("error_status_auth_missing", proxy.ErrorStatusAuthMissing)
		proxyParams.AddParam("error_headers_auth_missing", proxy.ErrorHeadersAuthMissing)
		proxyParams.AddParam("secret_token", proxy.SecretToken)

	case "plugin_rest":
		// Nothing!
	}

	switch backendVersion {
	case "oidc":
		proxyParams.AddParam("credentials_location", proxy.CredentialsLocation)
		proxyParams.AddParam("oidc_issuer_endpoint", proxy.OidcIssuerEndpoint)

	case "2":
		proxyParams.AddParam("credentials_location", proxy.CredentialsLocation)
		proxyParams.AddParam("auth_app_id", proxy.AuthAppID)
		proxyParams.AddParam("auth_app_key", proxy.AuthAppKey)

	case "1":
		proxyParams.AddParam("credentials_location", proxy.CredentialsLocation)
		proxyParams.AddParam("auth_user_key", proxy.AuthUserKey)

	}

	return proxyParams
}
