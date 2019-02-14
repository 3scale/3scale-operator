package v1alpha1

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	portaClient "github.com/3scale/3scale-porta-go-client/client"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strconv"
	"strings"
)

// TODO: Add options to enable defaults to builders

// ConsolidatedSpec defines the desired state of Consolidated
type ConsolidatedSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Credentials InternalCredentials `json:"credentials"`
	APIs        []InternalAPI       `json:"apis"`
}

//Sort function for consolidatedSpec
func (c ConsolidatedSpec) Sort() ConsolidatedSpec {

	for _, api := range c.APIs {
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
			sort.Slice(plan.Limits, func(i, j int) bool {
				if plan.Limits[i].Name != plan.Limits[j].Name {
					return plan.Limits[i].Name < api.Plans[j].Name
				} else {
					return plan.Limits[i].MaxValue < plan.Limits[j].MaxValue
				}
			})
		}
	}

	sort.Slice(c.APIs, func(i, j int) bool {
		if c.APIs[i].Name != c.APIs[j].Name {
			return c.APIs[i].Name < c.APIs[j].Name
		} else {
			return c.APIs[i].Description < c.APIs[j].Description
		}
	})

	return c
}

// ConsolidatedStatus defines the observed state of Consolidated
type ConsolidatedStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	PreviousVersion string `json:"previousVersion,omitempty"`
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
	Name string `json:"name"`
	APIBaseInternal
	Metrics []InternalMetric `json:"metrics,omitempty"`
	Plans   []InternalPlan   `json:"Plans,omitempty"`
}

func (api *InternalAPI) GetIntegrationName() string {
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
func (api *InternalAPI) GetIntegration() Integration {
	if api.IntegrationMethod.ApicastHosted != nil {
		return api.IntegrationMethod.ApicastHosted
	} else if api.IntegrationMethod.ApicastOnPrem != nil {
		return api.IntegrationMethod.ApicastOnPrem
	} else if api.IntegrationMethod.CodePlugin != nil {
		return api.IntegrationMethod.CodePlugin
	}
	return nil
}

type APIBaseInternal struct {
	APIBase
	// We shadow the APIBase IntegrationMethod to point to our Internal representation
	IntegrationMethod InternalIntegration `json:"integrationMethod"`
}

//func (api *InternalAPI) GetValidIntegrationMethod() (IntegrationMethod) {
//
//}

type InternalIntegration struct {
	ApicastOnPrem *InternalApicastOnPrem `json:"apicastOnPrem"`
	CodePlugin    *InternalCodePlugin    `json:"codePlugin"`
	ApicastHosted *InternalApicastHosted `json:"apicastHosted"`
}

type Integration interface {
	getMappingRules() []InternalMappingRule
	getCredentialTypeName() string
}

type InternalApicastHosted struct {
	APIcastBaseOptions
	MappingRules []InternalMappingRule `json:"mappingRules"`
	Policies     []InternalPolicy      `json:"policies"`
}

func (i *InternalApicastHosted) getCredentialTypeName() string {
	if i.AuthenticationSettings.Credentials.OpenIDConnector != nil {
		return "OpenIDConnector"

	} else if i.AuthenticationSettings.Credentials.APIKey != nil {
		return "APIKey"

	} else if i.AuthenticationSettings.Credentials.AppID != nil {
		return "AppID"

	}
	return ""
}
func (i *InternalApicastHosted) getMappingRules() []InternalMappingRule {
	return i.MappingRules
}

type InternalApicastOnPrem struct {
	APIcastBaseOptions
	StagingPublicBaseURL    string                `json:"stagingPublicBaseURL"`
	ProductionPublicBaseURL string                `json:"productionPublicBaseURL"`
	MappingRules            []InternalMappingRule `json:"mappingRules"`
	Policies                []InternalPolicy      `json:"policies"`
}

func (i *InternalApicastOnPrem) getCredentialTypeName() string {
	if i.AuthenticationSettings.Credentials.OpenIDConnector != nil {
		return "OpenIDConnector"

	} else if i.AuthenticationSettings.Credentials.APIKey != nil {
		return "APIKey"

	} else if i.AuthenticationSettings.Credentials.AppID != nil {
		return "AppID"
	}
	return ""
}
func (i *InternalApicastOnPrem) getMappingRules() []InternalMappingRule {
	return i.MappingRules
}

type InternalCodePlugin struct {
	AuthenticationSettings CodePluginAuthenticationSettings `json:"authenticationSettings"`
}

func (i *InternalCodePlugin) getCredentialTypeName() string {
	if i.AuthenticationSettings.Credentials.OpenIDConnector != nil {
		return "OpenIDConnector"
	} else if i.AuthenticationSettings.Credentials.APIKey != nil {
		return "APIKey"

	} else if i.AuthenticationSettings.Credentials.AppID != nil {
		return "AppID"

	}
	return ""

}
func (i *InternalCodePlugin) getMappingRules() []InternalMappingRule {
	return []InternalMappingRule{}
}

type InternalPolicy struct {
	Enabled bool `json:"enabled"`
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
	Default          bool            `json:"default"`
	TrialPeriodDays  int64           `json:"trialPeriodDays"`
	ApprovalRequired bool            `json:"approvalRequired"`
	Costs            PlanCost        `json:"costs"`
	Limits           []InternalLimit `json:"limits"`
}
type InternalLimit struct {
	Name     string `json:"name"`
	Period   string `json:"period"`
	MaxValue int64  `json:"maxValue"`
	Metric   string `json:"metric"`
}

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

func (d *APIsDiff) ReconcileWith3scale(creds InternalCredentials) error {

	for _, api := range d.MissingFromB {
		err := CreateInternalAPIIn3scale(creds, api)
		if err != nil {
			return err
		}
	}

	for _, api := range d.MissingFromA {
		err := DeleteInternalAPIFrom3scale(creds, api)
		if err != nil {
			return err
		}
	}

	for _, apiPair := range d.NotEqual {

		c, err := NewPortaClient(creds)
		if err != nil {
			return err
		}

		serviceNeedsUpdate := false
		service, err := getServiceFromInternalAPI(c, apiPair.A.Name)
		if err != nil {
			return err
		}
		serviceParams := portaClient.Params{}

		// Check if DeploymentOption is correct
		desiredDeploymentOption := IntegrationMethodToDeploymentType[apiPair.A.GetIntegrationName()]
		existingDeploymentOption := IntegrationMethodToDeploymentType[apiPair.B.GetIntegrationName()]

		if desiredDeploymentOption != existingDeploymentOption {
			serviceNeedsUpdate = true
			serviceParams.AddParam("deployment_option", desiredDeploymentOption)
		}

		// Check if BackendVersion is correct
		desiredBackendVersion := CredentialTypeToBackendVersion[apiPair.A.GetIntegration().getCredentialTypeName()]
		existingBackendVersion := CredentialTypeToBackendVersion[apiPair.B.GetIntegration().getCredentialTypeName()]

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
		metricsDiff := DiffMetrics(apiPair.A.Metrics, apiPair.B.Metrics)
		err = metricsDiff.ReconcileWith3scale(c, service.ID, apiPair.A)
		if err != nil {
			return err
		}

		// ReconcileWith3scale Mapping Rules
		// TODO: Create diffMappingRules
		desiredMappingRules := apiPair.A.GetIntegration().getMappingRules()
		existingMappingRules := apiPair.B.GetIntegration().getMappingRules()

		for _, desiredMappingRule := range desiredMappingRules {
			found := false
			for _, existingMappingRule := range existingMappingRules {
				if desiredMappingRule.Method == existingMappingRule.Method &&
					desiredMappingRule.Increment == existingMappingRule.Increment &&
					desiredMappingRule.Path == existingMappingRule.Path &&
					desiredMappingRule.Metric == existingMappingRule.Metric {
					found = true
					break
				}
			}

			if !found {
				metric, err := metricNametoMetric(c, service.ID, desiredMappingRule.Metric)
				if err != nil {
					return err
				}

				_, err = c.CreateMappingRule(service.ID, strings.ToUpper(desiredMappingRule.Method), desiredMappingRule.Path, int(desiredMappingRule.Increment), metric.ID)
				if err != nil {
					return err
				}
			}
		}

		for _, existingMappingRule := range existingMappingRules {
			found := false

			for _, desiredMappingRule := range desiredMappingRules {
				if desiredMappingRule.Method == existingMappingRule.Method &&
					desiredMappingRule.Increment == existingMappingRule.Increment &&
					desiredMappingRule.Path == existingMappingRule.Path &&
					desiredMappingRule.Metric == existingMappingRule.Metric {
					found = true
					break
				}
			}

			if !found {
				mappingRule, err := get3scaleMappingRulefromInternalMappingRule(c, service.ID, existingMappingRule)
				if err != nil {
					return err
				}
				err = c.DeleteMappingRule(service.ID, mappingRule.ID)
				if err != nil {
					return err
				}
			}
		}

		// Because MappingRules are not Unique, let's remove duplicated mappingRules

		// ReconcileWith3scale Plans
		plansDiff := DiffPlans(apiPair.A.Plans, apiPair.B.Plans)
		err = plansDiff.ReconcileWith3scale(c, service.ID, apiPair.A)
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

type MetricsDiff struct {
	MissingFromA []InternalMetric
	MissingFromB []InternalMetric
	Equal        []InternalMetric
	NotEqual     []MetricPair
}
type MetricPair struct {
	A InternalMetric
	B InternalMetric
}

func (d *MetricsDiff) ReconcileWith3scale(c *portaClient.ThreeScaleClient, serviceId string, api InternalAPI) error {

	for _, metric := range d.MissingFromB {
		err := CreateInternalMetricIn3scale(c, api, metric)
		if err != nil {
			return err
		}
	}

	for _, metric := range d.MissingFromA {
		err := DeleteInternalMetricFrom3scale(c, api, metric)
		if err != nil {
			return err
		}
	}

	// Now, update the existing metric with the desired metric, NotEqual contains the
	// metric pair, A and B, being A the desired, and B the existing.
	for _, metric := range d.NotEqual {

		// We need the metric ID in 3scale.
		metric3scale, err := metricNametoMetric(c, serviceId, metric.B.Name)
		if err != nil {
			return err
		}

		// We Update both fields, we don't want to loose any data in stats or so.
		params := portaClient.NewParams()
		params.AddParam("description", metric.A.Description)
		params.AddParam("unit", metric.A.Unit)

		_, err = c.UpdateMetric(serviceId, metric3scale.ID, params)
		if err != nil {
			return err
		}
	}

	return nil

}

type PlansDiff struct {
	MissingFromA []InternalPlan
	MissingFromB []InternalPlan
	Equal        []InternalPlan
	NotEqual     []PlanPair
}
type PlanPair struct {
	A InternalPlan
	B InternalPlan
}

func (d *PlansDiff) ReconcileWith3scale(c *portaClient.ThreeScaleClient, serviceId string, api InternalAPI) error {

	for _, plan := range d.MissingFromA {
		plan3scale, err := get3scalePlanFromInternalPlan(c, serviceId, plan)
		if err != nil {
			return err
		}

		err = c.DeleteAppPlan(serviceId, plan3scale.ID)
		if err != nil {
			return err
		}
	}

	for _, plan := range d.MissingFromB {
		plan3scale, err := c.CreateAppPlan(serviceId, plan.Name, "publish")
		if err != nil {
			return err
		}
		params := portaClient.Params{
			"approval_required": strconv.FormatBool(plan.ApprovalRequired),
			"setup_fee":         strconv.FormatFloat(plan.Costs.SetupFee, 'f', 1, 64),
			"cost_per_month":    strconv.FormatFloat(plan.Costs.CostMonth, 'f', 1, 64),
			"trial_period_days": strconv.FormatInt(plan.TrialPeriodDays, 10),
		}
		_, err = c.UpdateAppPlan(serviceId, plan3scale.ID, plan3scale.PlanName, "publish", params)

		if plan.Default {
			_, err = c.SetDefaultPlan(serviceId, plan3scale.ID)
		}
	}

	for _, planPair := range d.NotEqual {
		plan3scale, err := get3scalePlanFromInternalPlan(c, serviceId, planPair.B)
		if err != nil {
			return err
		}
		params := portaClient.Params{
			"approval_required": strconv.FormatBool(planPair.A.ApprovalRequired),
			"setup_fee":         strconv.FormatFloat(planPair.A.Costs.SetupFee, 'f', 1, 64),
			"cost_per_month":    strconv.FormatFloat(planPair.A.Costs.CostMonth, 'f', 1, 64),
			"trial_period_days": strconv.FormatInt(planPair.A.TrialPeriodDays, 10),
		}
		_, err = c.UpdateAppPlan(serviceId, plan3scale.ID, plan3scale.PlanName, "publish", params)

		if planPair.A.Default {
			_, err = c.SetDefaultPlan(serviceId, plan3scale.ID)
		}

		limitsDiff := DiffLimits(planPair.A.Limits, planPair.B.Limits)
		err = limitsDiff.ReconcileWith3scale(c, serviceId, plan3scale.ID)
		if err != nil {
			return err
		}

	}

	return nil

}

type LimitsDiff struct {
	MissingFromA []InternalLimit
	MissingFromB []InternalLimit
	Equal        []InternalLimit
	NotEqual     []LimitPair
}
type LimitPair struct {
	A InternalLimit
	B InternalLimit
}

func (d *LimitsDiff) ReconcileWith3scale(c *portaClient.ThreeScaleClient, serviceId string, planID string) error {

	for _, limit := range d.MissingFromA {
		metric, err := metricNametoMetric(c, serviceId, limit.Metric)
		if err != nil {
			return err
		}
		limit3scale, err := get3scaleLimitFromInternalLimit(c, serviceId, planID, limit)
		if err != nil {
			return err
		}
		//TODO: Delete always report an error, fix this, for now, we ignore it.
		_ = c.DeleteLimitPerAppPlan(planID, metric.ID, limit3scale.ID)
	}

	for _, limit := range d.MissingFromB {
		metric, err := metricNametoMetric(c, serviceId, limit.Metric)
		if err != nil {
			return err
		}
		_, err = c.CreateLimitAppPlan(planID, metric.ID, limit.Period, int(limit.MaxValue))
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	SchemeBuilder.Register(&Consolidated{}, &ConsolidatedList{})
}

func DiffLimits(aLimits []InternalLimit, bLimits []InternalLimit) LimitsDiff {
	limitDiff := LimitsDiff{}
	for _, aLimit := range aLimits {
		found := false
		for _, bLimit := range bLimits {
			if aLimit.Metric == bLimit.Metric &&
				aLimit.MaxValue == bLimit.MaxValue &&
				aLimit.Period == bLimit.Period {
				found = true
				limitDiff.Equal = append(limitDiff.Equal, aLimit)
				break
			}
		}

		if !found {
			limitDiff.MissingFromB = append(limitDiff.MissingFromB, aLimit)
		}
	}

	for _, bLimit := range bLimits {
		found := false
		for _, aLimit := range aLimits {
			if aLimit == bLimit {
				found = true
				break
			}
		}
		if !found {
			limitDiff.MissingFromA = append(limitDiff.MissingFromA, bLimit)
		}
	}

	return limitDiff

}
func DiffMetrics(aMetrics, bMetrics []InternalMetric) MetricsDiff {

	metricsDiff := MetricsDiff{}
	for _, aMetric := range aMetrics {
		found := false
		for _, bMetric := range bMetrics {
			if aMetric.Name == bMetric.Name {
				found = true
				if aMetric == bMetric {
					metricsDiff.Equal = append(metricsDiff.Equal, aMetric)
				} else {
					metricPair := MetricPair{
						A: aMetric,
						B: bMetric,
					}
					metricsDiff.NotEqual = append(metricsDiff.NotEqual, metricPair)
				}
				break
			}
		}

		if !found {
			metricsDiff.MissingFromB = append(metricsDiff.MissingFromB, aMetric)
		}
	}

	for _, bMetric := range bMetrics {
		found := false
		for _, aMetric := range aMetrics {
			if aMetric == bMetric {
				found = true
				break
			}
		}
		if !found {
			metricsDiff.MissingFromA = append(metricsDiff.MissingFromA, bMetric)
		}
	}

	return metricsDiff
}
func DiffAPIs(desiredAPIs []InternalAPI, existingAPIs []InternalAPI) APIsDiff {

	var apisDiff APIsDiff
	for _, desiredAPI := range desiredAPIs {
		found := false
		for _, existingAPI := range existingAPIs {
			if existingAPI.Name == desiredAPI.Name {
				if CompareInternalAPI(existingAPI, desiredAPI) {
					apisDiff.Equal = append(apisDiff.Equal, desiredAPI)
				} else {
					apiPair := APIPair{
						A: desiredAPI,
						B: existingAPI,
					}
					apisDiff.NotEqual = append(apisDiff.NotEqual, apiPair)
				}
				found = true
				break
			}
		}
		if !found {
			apisDiff.MissingFromB = append(apisDiff.MissingFromB, desiredAPI)
		}
	}

	for _, existingAPI := range existingAPIs {
		found := false
		for _, desiredAPI := range desiredAPIs {
			if desiredAPI.Name == existingAPI.Name {
				found = true
				break
			}
		}
		if !found {
			apisDiff.MissingFromA = append(apisDiff.MissingFromA, existingAPI)
		}
	}

	return apisDiff
}
func DiffPlans(aPlans []InternalPlan, bPlans []InternalPlan) PlansDiff {

	var plansDiff PlansDiff
	for _, aPlan := range aPlans {
		found := false
		for _, bPlan := range bPlans {
			if bPlan.Name == aPlan.Name {
				found = true
				if ComparePlans(aPlan, bPlan) {
					plansDiff.Equal = append(plansDiff.Equal, aPlan)
				} else {
					planPair := PlanPair{
						A: aPlan,
						B: bPlan,
					}
					plansDiff.NotEqual = append(plansDiff.NotEqual, planPair)
				}
				break
			}
		}
		if !found {
			plansDiff.MissingFromB = append(plansDiff.MissingFromB, aPlan)
		}
	}

	for _, bPlan := range bPlans {
		found := false
		for _, aPlan := range aPlans {
			if aPlan.Name == bPlan.Name {
				found = true
				break
			}
		}
		if !found {
			plansDiff.MissingFromA = append(plansDiff.MissingFromA, bPlan)
		}
	}

	return plansDiff
}

func NewConsolidatedFromBinding(binding Binding, c client.Client) (*Consolidated, error) {

	consolidated := Consolidated{
		TypeMeta: metav1.TypeMeta{
			Kind: "Consolidated",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      binding.Name + "-consolidated",
			Namespace: binding.Namespace,
		},
		Spec: ConsolidatedSpec{
			Credentials: InternalCredentials{},
			APIs:        nil,
		},
		Status: ConsolidatedStatus{},
	}
	// GET SECRET
	secret := &v1.Secret{}
	// TODO: fix namespace default
	err := c.Get(context.TODO(), types.NamespacedName{Name: binding.Spec.CredentialsRef.Name, Namespace: binding.Namespace}, secret)

	if err != nil && errors.IsNotFound(err) {
		log.Printf("secret: %s for binding: %s NOT found", binding.Spec.CredentialsRef.Name, binding.Name)
		//return reconcile.Result{}, err
		return nil, err
	} else if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	} else {
		log.Printf("secret: %s found for %s", binding.Spec.CredentialsRef.Name, binding.Name)

		consolidated.Spec.Credentials = InternalCredentials{
			AuthToken: string(secret.Data["token"]),
			AdminURL:  string(secret.Data["adminURL"]),
		}
	}
	// GET APIS
	apis, err := getAPIs(binding.Namespace, binding.Spec.APISelector.MatchLabels, c)
	if err != nil && errors.IsNotFound(err) {
		// No API objects
		log.Printf("Binding: %s in namespace: %s doesn't match any API object", binding.Name, binding.Namespace)
		return nil, err
	} else if err != nil {
		// Something is broken
		log.Printf("Error: %s", err)
		return nil, err
	}

	// Add each API info to the consolidated object
	for _, api := range apis.Items {
		internalAPI, err := NewInternalAPIFromAPI(api, c)
		if err != nil {
			log.Printf("Error on InternalAPI: %s", err)
		} else {
			consolidated.Spec.APIs = append(consolidated.Spec.APIs, *internalAPI)
		}
	}

	return &consolidated, nil
}
func NewInternalAPIFromAPI(api API, c client.Client) (*InternalAPI, error) {

	internalAPI := InternalAPI{
		Name: api.Name,
		APIBaseInternal: APIBaseInternal{
			APIBase: APIBase{
				Description: api.Spec.Description,
			},
			IntegrationMethod: InternalIntegration{},
		},
		Metrics: nil,
		Plans:   nil,
	}

	//Get Metrics for each API

	metrics, err := getMetrics(api.Namespace, api.Spec.MetricSelector.MatchLabels, c)
	if err != nil && errors.IsNotFound(err) {
		// Nothing has been found
		log.Printf("No metrics found for: %s\n", api.Name)
	} else if err != nil {
		// Something is broken
		log.Printf("Error: %s", err)
		return nil, err
	} else {
		// Let's do our job.
		for _, metric := range metrics.Items {
			internalMetric := NewInternalMetricFromMetric(metric)
			internalAPI.Metrics = append(internalAPI.Metrics, *internalMetric)
		}
	}
	//Get Plans for each API
	plans, err := getPlans(api.Namespace, api.Spec.PlanSelector.MatchLabels, c)

	if err != nil && errors.IsNotFound(err) {
		// Nothing has been found
		log.Printf("No plans found for: %s\n", api.Name)
	} else if err != nil {
		// Something is broken
		log.Printf("Error: %s", err)
		return nil, err
	} else {
		// Let's do our job.
		for _, plan := range plans.Items {
			internalPlan, err := NewInternalPlanFromPlan(plan, c)
			if err != nil {
				log.Printf("Error: %s", err)
				return nil, err
			}
			internalAPI.Plans = append(internalAPI.Plans, *internalPlan)
		}
	}

	switch api.getIntegrationMethodType() {

	case "ApicastHosted":

		log.Println("InternalIntegration method: ApicastHosted")
		internalApicastHosted, err := NewInternalApicastHostedFromApicastHosted(api.Namespace, *api.Spec.IntegrationMethod.ApicastHosted, c)
		if err != nil {
			return nil, err
		} else {
			internalAPI.IntegrationMethod.ApicastHosted = internalApicastHosted
		}

	case "ApicastOnPrem":
		log.Println("InternalIntegration method: ApicastOnPrem")
		internalApicastOnPrem, err := NewInternalApicastOnPremFromApicastOnPrem(api.Namespace, *api.Spec.IntegrationMethod.ApicastOnPrem, c)
		if err != nil {
			return nil, err
		} else {
			internalAPI.IntegrationMethod.ApicastOnPrem = internalApicastOnPrem
		}
		// TODO: Policies

	case "CodePlugin":
		// Assume Code plugin.
		internalCodePlugin := InternalCodePlugin{
			AuthenticationSettings: CodePluginAuthenticationSettings{
				Credentials: api.Spec.IntegrationMethod.CodePlugin.AuthenticationSettings.Credentials,
			},
		}
		internalAPI.IntegrationMethod.CodePlugin = &internalCodePlugin
		log.Println("InternalIntegration method: CodePlugin")

	default:
		log.Printf("invalid integration method of api: %s\n", api.Name)
		return nil, err
	}

	return &internalAPI, nil
}
func NewInternalPlanFromPlan(plan Plan, c client.Client) (*InternalPlan, error) {

	// Fill the internal Plan with Plan and Limits.
	internalPlan := InternalPlan{
		Name:             plan.Name,
		Default:          plan.Spec.Default,
		TrialPeriodDays:  plan.Spec.TrialPeriod,
		ApprovalRequired: plan.Spec.AprovalRequired,
		Costs:            plan.Spec.Costs,
		Limits:           nil,
	}
	// Get the Limits now
	limits, err := getLimits(plan.Namespace, plan.Spec.LimitSelector.MatchLabels, c)

	if err != nil && errors.IsNotFound(err) {
		// Nothing has been found
		log.Printf("No Limits found for plan: %s\n", plan.Name)
	} else if err != nil {
		// Something is broken
		return nil, err
	} else {
		// Let's do our job.
		for _, limit := range limits.Items {
			internalLimit, err := NewInternalLimitFromLimit(limit, c)
			if err != nil {
				//TODO: UPDATE STATUS OBJECT
				log.Printf("limit %s couldn't be converted", limit.Name)
			} else {
				internalPlan.Limits = append(internalPlan.Limits, *internalLimit)
			}
		}
	}
	return &internalPlan, nil
}
func NewInternalLimitFromLimit(limit Limit, c client.Client) (*InternalLimit, error) {
	metric := &Metric{}
	var namespace string
	var internalLimit InternalLimit

	if limit.Spec.Metric.Namespace == "" {
		namespace = limit.Namespace
	} else {
		namespace = limit.Spec.Metric.Namespace
	}
	reference := types.NamespacedName{
		Namespace: namespace,
		Name:      limit.Spec.Metric.Name,
	}
	err := c.Get(context.TODO(), reference, metric)
	if err != nil {
		// Something is broken
		return nil, err
	} else {
		internalLimit = InternalLimit{
			Name:     limit.Name,
			Period:   limit.Spec.Period,
			MaxValue: limit.Spec.MaxValue,
			Metric:   metric.Name,
		}
	}
	return &internalLimit, nil
}
func NewInternalMappingRuleFromMappingRule(mappingRule MappingRule, c client.Client) (*InternalMappingRule, error) {
	// GET metric for mapping rule.
	metric := &Metric{}
	var namespace string

	// Handle metrics Hits.
	if mappingRule.Spec.MetricRef.Name == "Hits" ||
		mappingRule.Spec.MetricRef.Name == "hits" {
		metric.Name = "hits"

		// Handle metrics Hits.

	} else {
		if mappingRule.Spec.MetricRef.Namespace == "" {
			namespace = mappingRule.Namespace
		} else {
			namespace = mappingRule.Spec.MetricRef.Namespace
		}
		reference := types.NamespacedName{
			Namespace: namespace,
			Name:      mappingRule.Spec.MetricRef.Name,
		}
		err := c.Get(context.TODO(), reference, metric)
		if err != nil {
			// Something is broken
			return nil, err
		}
	}
	internalMappingRule := InternalMappingRule{
		Name:      mappingRule.Name,
		Path:      mappingRule.Spec.Path,
		Method:    mappingRule.Spec.Method,
		Increment: mappingRule.Spec.Increment,
		Metric:    metric.Name,
	}

	return &internalMappingRule, nil
}
func NewInternalMetricFromMetric(metric Metric) *InternalMetric {
	internalMetric := InternalMetric{
		Name:        metric.Name,
		Unit:        metric.Spec.Unit,
		Description: metric.Spec.Description,
	}

	return &internalMetric
}
func NewInternalApicastHostedFromApicastHosted(namespace string, hosted ApicastHosted, c client.Client) (*InternalApicastHosted, error) {

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
			internalMappingRule, err := NewInternalMappingRuleFromMappingRule(mappingRule, c)
			if err != nil {
				log.Printf("mappingRule %s couldn't be converted", mappingRule.Name)
			} else {
				internalApicastHosted.MappingRules = append(internalApicastHosted.MappingRules, *internalMappingRule)
			}
		}
	}
	return &internalApicastHosted, nil
}
func NewInternalApicastOnPremFromApicastOnPrem(namespace string, prem ApicastOnPrem, c client.Client) (*InternalApicastOnPrem, error) {
	internalApicastOnPrem := InternalApicastOnPrem{
		APIcastBaseOptions: APIcastBaseOptions{
			PrivateBaseURL:         prem.PrivateBaseURL,
			APITestGetRequest:      prem.APITestGetRequest,
			AuthenticationSettings: prem.AuthenticationSettings,
		},
		StagingPublicBaseURL:    prem.StagingPublicBaseURL,
		ProductionPublicBaseURL: prem.PrivateBaseURL,
		MappingRules:            nil,
		Policies:                nil,
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
			internalMappingRule, err := NewInternalMappingRuleFromMappingRule(mappingRule, c)
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

func CompareConsolidated(consolidatedA, consolidatedB Consolidated) bool {

	//Sort all the object
	A := consolidatedA.Spec.Sort()
	B := consolidatedB.Spec.Sort()

	//Check the credentials
	if !reflect.DeepEqual(A.Credentials, B.Credentials) {
		return false
	}
	// Check if we have the same number of APIs
	if len(A.APIs) == len(B.APIs) {
		// Compare APIs one by one.
		for i := range A.APIs {
			if !CompareInternalAPI(A.APIs[i], B.APIs[i]) {
				return false
			}
		}
	} else {
		return false
	}
	return true
}
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
	if APIA.GetIntegrationName() != APIB.GetIntegrationName() {
		return false
	} else {
		switch APIA.GetIntegrationName() {
		case "ApicastOnPrem":
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
func ComparePlans(a, b InternalPlan) bool {

	if a.Name == b.Name && a.ApprovalRequired == b.ApprovalRequired &&
		a.TrialPeriodDays == b.TrialPeriodDays && a.Costs == b.Costs &&
		a.Default == b.Default {
		if len(a.Limits) == len(b.Limits) {
			for i := range a.Limits {
				if a.Limits[i] != b.Limits[i] {
					return false
				}
			}
		} else {
			return false
		}
	} else {
		return false
	}

	return true
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
func getServiceMappingRulesFrom3scale(c *portaClient.ThreeScaleClient, service portaClient.Service) (*[]InternalMappingRule, error) {

	var mappingRules []InternalMappingRule
	mappingRulesFrom3scale, _ := c.ListMappingRule(service.ID)

	for _, mapping := range mappingRulesFrom3scale.MappingRules {

		desiredMetricName := ""
		for _, metric := range service.Metrics.Metrics {
			if metric.ID == mapping.MetricID {
				desiredMetricName = metric.SystemName
			}
		}
		if desiredMetricName == "" {
			// This should never happen
			return nil, fmt.Errorf("mappingrule with invalid metric")
		} else {
			metricIncrement, err := strconv.ParseInt(mapping.Delta, 10, 64)
			if err != nil {
				return nil, err
			}
			internalMappingRule := InternalMappingRule{
				Name:      mapping.XMLName.Local,
				Path:      mapping.Pattern,
				Method:    mapping.HTTPMethod,
				Increment: metricIncrement,
				Metric:    desiredMetricName,
			}
			mappingRules = append(mappingRules, internalMappingRule)
		}
	}
	return &mappingRules, nil
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
	return portaClient.Service{}, fmt.Errorf("notfound")
}
func get3scaleLimitFromInternalLimit(c *portaClient.ThreeScaleClient, serviceID string, planID string, limit InternalLimit) (portaClient.Limit, error) {

	limits3scale, err := c.ListLimitsPerAppPlan(planID)
	if err != nil {
		return portaClient.Limit{}, err
	}
	metric3scale, err := metricNametoMetric(c, serviceID, limit.Metric)
	if err != nil {
		return portaClient.Limit{}, err
	}

	limitValueString := strconv.FormatInt(limit.MaxValue, 10)

	for _, limit3scale := range limits3scale.Limits {

		if limit3scale.Value == limitValueString &&
			limit3scale.Period == limit.Period &&
			limit3scale.MetricID == metric3scale.ID {

			return limit3scale, nil
		}
	}

	return portaClient.Limit{}, fmt.Errorf("limit not found")
}
func get3scalePlanFromInternalPlan(c *portaClient.ThreeScaleClient, serviceID string, plan InternalPlan) (portaClient.Plan, error) {
	plans3scale, err := c.ListAppPlanByServiceId(serviceID)
	if err != nil {
		return portaClient.Plan{}, err

	}
	for _, plan3scale := range plans3scale.Plans {
		if plan3scale.PlanName == plan.Name {
			return plan3scale, nil
		}
	}

	return portaClient.Plan{}, fmt.Errorf("not found")
}
func get3scaleMappingRulefromInternalMappingRule(c *portaClient.ThreeScaleClient, serviceID string, internalMappingRule InternalMappingRule) (portaClient.MappingRule, error) {
	mappingRules, err := c.ListMappingRule(serviceID)
	metric, err := metricNametoMetric(c, serviceID, internalMappingRule.Metric)
	internalIncrement := strconv.FormatInt(internalMappingRule.Increment, 10)
	if err != nil {
		return portaClient.MappingRule{}, err
	}

	for _, mappingRule := range mappingRules.MappingRules {
		if mappingRule.HTTPMethod == internalMappingRule.Method &&
			mappingRule.Pattern == internalMappingRule.Path &&
			mappingRule.MetricID == metric.ID &&
			mappingRule.Delta == internalIncrement {
			return mappingRule, nil
		}
	}

	return portaClient.MappingRule{}, fmt.Errorf("not found")
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

func getAPIs(namespace string, matchLabels map[string]string, c client.Client) (*APIList, error) {
	apis := &APIList{}
	opts := &client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), opts, apis)
	return apis, err
}
func getMappingRules(namespace string, matchLabels map[string]string, c client.Client) (*MappingRuleList, error) {
	mappingRules := &MappingRuleList{}
	opts := client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), &opts, mappingRules)
	return mappingRules, err
}
func getMetrics(namespace string, matchLabels map[string]string, c client.Client) (*MetricList, error) {
	metrics := &MetricList{}
	opts := client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), &opts, metrics)
	return metrics, err
}
func getPlans(namespace string, matchLabels map[string]string, c client.Client) (*PlanList, error) {
	plans := &PlanList{}
	opts := client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), &opts, plans)
	return plans, err
}
func getLimits(namespace string, matchLabels map[string]string, c client.Client) (*LimitList, error) {
	limits := &LimitList{}
	opts := client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), &opts, limits)
	return limits, err
}

func NewConsolidatedFrom3scale(creds InternalCredentials, apis []InternalAPI) (*Consolidated, error) {

	consolidatedSpec := ConsolidatedSpec{
		Credentials: creds,
		APIs:        nil,
	}
	c, err := NewPortaClient(creds)
	if err != nil {
		return &Consolidated{}, err
	}

	for _, desiredAPI := range apis {
		internalAPI, err := GetInternalAPIfrom3scale(c, desiredAPI)
		if err != nil {
			log.Printf("API %s doesn't exists in 3scale", desiredAPI.Name)
		} else {
			consolidatedSpec.APIs = append(consolidatedSpec.APIs, *internalAPI)
		}
	}

	consolidated := Consolidated{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       consolidatedSpec,
		Status:     ConsolidatedStatus{},
	}

	return &consolidated, nil
}
func GetInternalAPIfrom3scale(c *portaClient.ThreeScaleClient, api InternalAPI) (*InternalAPI, error) {

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
				Policies:                nil,
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
				Policies:     nil,
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
			Name:        metric.MetricName,
			Unit:        metric.Unit,
			Description: metric.Description,
		}
		if internalMetric.Name != "hits" {
			internalAPI.Metrics = append(internalAPI.Metrics, internalMetric)
		}
	}

	for _, applicationPlan := range applicationPlans.Plans {

		trialPeriodDays, _ := strconv.ParseInt(applicationPlan.TrialPeriodDays, 10, 64)
		approvalRequired, _ := strconv.ParseBool(applicationPlan.ApprovalRequired)
		setupFee, _ := strconv.ParseFloat(applicationPlan.SetupFee, 64)
		costMonth, _ := strconv.ParseFloat(applicationPlan.CostPerMonth, 64)

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
				Metric:   metric.SystemName,
			}
			internalPlan.Limits = append(internalPlan.Limits, internalLimit)
		}

		internalAPI.Plans = append(internalAPI.Plans, internalPlan)
	}

	return &internalAPI, nil
}

func metricNametoMetric(c *portaClient.ThreeScaleClient, serviceID string, metricName string) (portaClient.Metric, error) {
	m := portaClient.Metric{}
	metrics, err := c.ListMetrics(serviceID)
	if err != nil {
		return m, err
	}

	for _, metric := range metrics.Metrics {
		if metricName == metric.SystemName {
			m = metric
			return m, nil
		}
	}

	return m, fmt.Errorf("metric not found")

}
func metricIDtoMetric(c *portaClient.ThreeScaleClient, serviceID string, metricID string) (portaClient.Metric, error) {
	m := portaClient.Metric{}

	metrics, err := c.ListMetrics(serviceID)
	if err != nil {
		return m, err
	}

	for _, metric := range metrics.Metrics {
		if metricID == metric.ID {
			m = metric
			break
		}
	}

	return m, nil

}
func NewPortaClient(creds InternalCredentials) (*portaClient.ThreeScaleClient, error) {

	systemAdminPortalURL, err := url.Parse(creds.AdminURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing 3scale url from crd - %s", err.Error())
	}

	port := 0
	if systemAdminPortalURL.Port() == "" {
		switch scheme := systemAdminPortalURL.Scheme; scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		}
	} else {
		port, err = strconv.Atoi(systemAdminPortalURL.Port())
		if err != nil {
			return nil, fmt.Errorf("admin portal URL invalid port - %s" + err.Error())
		}
	}

	adminPortal, err := portaClient.NewAdminPortal(systemAdminPortalURL.Scheme, systemAdminPortalURL.Host, port)
	if err != nil {
		return nil, fmt.Errorf("invalid Admin Portal URL: %s", err.Error())
	}

	// TODO - This should be secure by default and overrideable for testing
	// TODO - Set some sensible default here to handle timeouts
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	insecureHttp := &http.Client{Transport: tr}

	c := portaClient.NewThreeScale(adminPortal, creds.AuthToken, insecureHttp)
	return c, nil
}

func CreateInternalMetricIn3scale(c *portaClient.ThreeScaleClient, api InternalAPI, metric InternalMetric) error {

	service, err := getServiceFromInternalAPI(c, api.Name)
	if err != nil {
		return err
	}
	_, err = c.CreateMetric(service.ID, metric.Name, metric.Description, metric.Unit)
	return err
}
func CreateInternalAPIIn3scale(creds InternalCredentials, api InternalAPI) error {

	c, err := NewPortaClient(creds)
	if err != nil {
		return err
	}

	// Get the proper 3scale deployment Option based on the integrationMethod
	deploymentOption := IntegrationMethodToDeploymentType[api.GetIntegrationName()]
	if deploymentOption == "" {
		return fmt.Errorf("unknown integration method")
	}

	// Get the proper backendVersion based on the CredentialType

	backendVersion := CredentialTypeToBackendVersion[api.GetIntegration().getCredentialTypeName()]
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

	for _, mappingRule := range api.GetIntegration().getMappingRules() {
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
			"setup_fee":         strconv.FormatFloat(plan.Costs.SetupFee, 'f', 1, 64),
			"cost_per_month":    strconv.FormatFloat(plan.Costs.CostMonth, 'f', 1, 64),
			"trial_period_days": strconv.FormatInt(plan.TrialPeriodDays, 10),
		}
		_, err = c.UpdateAppPlan(service.ID, plan3scale.ID, plan3scale.PlanName, "publish", params)
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
func DeleteInternalAPIFrom3scale(creds InternalCredentials, api InternalAPI) error {

	c, err := NewPortaClient(creds)
	if err != nil {
		return err
	}

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
func DeleteInternalMetricFrom3scale(c *portaClient.ThreeScaleClient, api InternalAPI, metric InternalMetric) error {

	service, err := getServiceFromInternalAPI(c, api.Name)
	if err != nil {
		return err
	}

	metric3scale, err := metricNametoMetric(c, service.ID, metric.Name)
	if err != nil {
		return err
	}

	// TODO: fix DeleteMetric Returns always errors
	_ = c.DeleteMetric(service.ID, metric3scale.ID)
	//if err != nil {
	//	return err
	//}

	return nil
}
