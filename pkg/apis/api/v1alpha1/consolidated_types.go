package v1alpha1

import (
	"context"
	"encoding/json"
	portaClient "github.com/3scale/3scale-porta-go-client/client"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strconv"
)

// TODO: Add options to enable defaults to builders

// ConsolidatedSpec defines the desired state of Consolidated
type ConsolidatedSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Credentials InternalCredential `json:"credentials"`
	APIs        []InternalAPI      `json:"apis"`
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

type InternalAPI struct {
	Name string `json:"name"`
	APIBaseInternal
	Metrics []InternalMetric `json:"metrics,omitempty"`
	Plans   []InternalPlan   `json:"Plans,omitempty"`
}
type APIBaseInternal struct {
	APIBase
	// We shadow the APIBase IntegrationMethod to point to our Internal representation
	// Should this be improved? :)
	IntegrationMethod InternalIntegration `json:"integrationMethod"`
}
type InternalIntegration struct {
	ApicastOnPrem *InternalApicastOnPrem `json:"apicastOnPrem"`
	CodePlugin    *InternalCodePlugin    `json:"codePlugin"`
	ApicastHosted *InternalApicastHosted `json:"apicastHosted"`
}
type InternalApicastHosted struct {
	APIcastBaseOptions
	MappingRules []InternalMappingRule `json:"mappingRules"`
	Policies     []InternalPolicy      `json:"policies"`
}
type InternalApicastOnPrem struct {
	APIcastBaseOptions
	StagingPublicBaseURL    string                `json:"stagingPublicBaseURL"`
	ProductionPublicBaseURL string                `json:"productionPublicBaseURL"`
	MappingRules            []InternalMappingRule `json:"mappingRules"`
	Policies                []InternalPolicy      `json:"policies"`
}
type InternalCodePlugin struct {
	AuthenticationSettings CodePluginAuthenticationSettings `json:"authenticationSettings"`
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
	TrialPeriodDays  int64           `json:"trialPeriodDays"`
	ApprovalRequired bool            `json:"approvalRequired"`
	Costs            PlanCost        `json:"costs"`
	Limits           []InternalLimit `json:"limits"`
}
type InternalLimit struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Period      string `json:"period"`
	MaxValue    int64  `json:"maxValue"`
	Metric      string `json:"metric"`
}
type InternalCredential struct {
	AccessToken string `json:"accessToken"`
	AdminURL    string `json:"adminURL"`
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
			Credentials: InternalCredential{},
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
		consolidated.Spec.Credentials = InternalCredential{
			AccessToken: string(secret.Data["access_token"]),
			AdminURL:    string(secret.Data["admin_portal_url"]),
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
	// TODO: How to handle metric HITS.
	if err != nil {
		// Something is broken
		return nil, err
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
	//Let's compare only the Spec
	A, _ := json.Marshal(consolidatedA.Spec.Sort())
	B, _ := json.Marshal(consolidatedB.Spec.Sort())
	return reflect.DeepEqual(A, B)
}

func NewConsolidatedFrom3scale(creds InternalCredential, apis []InternalAPI) (*Consolidated, error) {

	consolidated := ConsolidatedSpec{
		Credentials: creds,
		APIs:        nil,
	}

	for _,desiredAPI := range apis {
		internalAPI, err := getInternalAPIfrom3scale(creds, desiredAPI)
		if err != nil {
			log.Printf("API %s doesn't exists in 3scale", internalAPI.Name)
		} else {
			consolidated.APIs = append(consolidated.APIs, *internalAPI)
		}
	}

	return &Consolidated{}, nil
}

func getInternalAPIfrom3scale(creds InternalCredential, api InternalAPI) (*InternalAPI, error) {

	return &InternalAPI{}, nil
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




// Returns a list of InternalPlans from the 3scale account based on a serviceID.
func getInternalPlansFrom3scale(c *portaClient.ThreeScaleClient, svcId, accessToken string) (*[]InternalPlan, error) {
	var internalPlans []InternalPlan
	plansList, err := c.ListAppPlanByServiceId(accessToken, svcId)
	if err != nil {
		return nil, err
	}

	for _, appPlan := range plansList.Plans {
		limits, err := c.ListLimitsPerAppPlan(accessToken, appPlan.ID)
		if err != nil {
			return nil, err
		}
		var internalLimits []InternalLimit
		for _, limit := range limits.Limits {

			maxValue, err := strconv.ParseInt(limit.Value, 10, 64)
			if err != nil {
				return nil, err
			}

			internalLimit := InternalLimit{
				Name:        limit.XMLName.Local,
				Description: "",
				Period:      limit.Period,
				MaxValue:    maxValue,
				Metric:      limit.MetricID,
			}

			internalLimits = append(internalLimits, internalLimit)
		}

		trialPeriodDays, err := strconv.ParseInt(appPlan.TrialPeriodDays, 10, 64)
		if err != nil {
			return nil, err
		}

		setupFee, err := strconv.ParseInt(appPlan.SetupFee, 10, 64)
		if err != nil {
			return nil, err
		}

		costMonth, err := strconv.ParseInt(appPlan.CostPerMonth, 10, 64)
		if err != nil {
			return nil, err
		}

		approvalRequired, err := strconv.ParseBool(appPlan.EndUserRequired)
		if err != nil {
			return nil, err
		}

		internalPlan := &InternalPlan{
			Name:             appPlan.PlanName,
			TrialPeriodDays:  trialPeriodDays,
			ApprovalRequired: approvalRequired,
			Costs: PlanCost{
				SetupFee:  setupFee,
				CostMonth: costMonth,
			},
			Limits: internalLimits,
		}

		internalPlans = append(internalPlans, *internalPlan)

	}

	return &internalPlans, nil

}
