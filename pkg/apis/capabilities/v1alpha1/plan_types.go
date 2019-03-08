package v1alpha1

import (
	"context"
	"fmt"
	portaClient "github.com/3scale/3scale-porta-go-client/client"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strconv"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PlanSpec defines the desired state of Plan
type PlanSpec struct {
	PlanBase      `json:",inline"`
	PlanSelectors `json:",inline"`
}

type PlanBase struct {
	Default         bool  `json:"default"`
	TrialPeriod     int64 `json:"trialPeriod"`
	AprovalRequired bool  `json:"aprovalRequired"`
	// +optional
	Costs PlanCost `json:"costs,omitempty"`
}

type PlanSelectors struct {
	LimitSelector metav1.LabelSelector `json:"limitSelector"`
}

type PlanCost struct {
	SetupFee  float64 `json:"setupFee,omitempty"`
	CostMonth float64 `json:"costMonth,omitempty"`
}

// PlanStatus defines the observed state of Plan
type PlanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Plan is the Schema for the plans API
// +k8s:openapi-gen=true
type Plan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlanSpec   `json:"spec,omitempty"`
	Status PlanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PlanList contains a list of Plan
type PlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Plan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Plan{}, &PlanList{})
}

type InternalPlan struct {
	Name             string          `json:"name"`
	Default          bool            `json:"default"`
	TrialPeriodDays  int64           `json:"trialPeriodDays"`
	ApprovalRequired bool            `json:"approvalRequired"`
	Costs            PlanCost        `json:"costs"`
	Limits           []InternalLimit `json:"limits"`
}

func (plan *InternalPlan) Sort() {
	sort.Slice(plan.Limits, func(i, j int) bool {
		if plan.Limits[i].Name != plan.Limits[j].Name {
			return plan.Limits[i].Name < plan.Limits[j].Name
		} else {
			return plan.Limits[i].MaxValue < plan.Limits[j].MaxValue
		}
	})
}

type plansDiff struct {
	MissingFromA []InternalPlan
	MissingFromB []InternalPlan
	Equal        []InternalPlan
	NotEqual     []planPair
}
type planPair struct {
	A InternalPlan
	B InternalPlan
}

func (d *plansDiff) reconcileWith3scale(c *portaClient.ThreeScaleClient, serviceId string, api InternalAPI) error {

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
		_, err = c.UpdateAppPlan(serviceId, plan3scale.ID, plan3scale.PlanName, "", params)
		if err != nil {
			return err
		}
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
		stateEvent := ""

		if plan3scale.State != "published" {
			stateEvent = "publish"
		}

		_, err = c.UpdateAppPlan(serviceId, plan3scale.ID, plan3scale.PlanName, stateEvent, params)
		if err != nil {
			return err
		}

		if planPair.A.Default {
			_, err = c.SetDefaultPlan(serviceId, plan3scale.ID)
		}

		limitsDiff := diffLimits(planPair.A.Limits, planPair.B.Limits)
		err = limitsDiff.reconcileWith3scale(c, serviceId, plan3scale.ID)
		if err != nil {
			return err
		}
	}
	return nil

}
func diffPlans(Plans1 []InternalPlan, Plans2 []InternalPlan) plansDiff {

	var plansDiff plansDiff
	if len(Plans2) == 0 {
		plansDiff.MissingFromB = Plans1
		return plansDiff
	}
	for i := 0; i < 2; i++ {
		for _, plan1 := range Plans1 {
			found := false
			for _, plan2 := range Plans2 {
				if plan2.Name == plan1.Name {
					if i == 0 {
						if comparePlans(plan1, plan2) {
							plansDiff.Equal = append(plansDiff.Equal, plan1)
						} else {
							planPair := planPair{
								A: plan1,
								B: plan2,
							}
							plansDiff.NotEqual = append(plansDiff.NotEqual, planPair)
						}
					}
					found = true
					break
				}
			}
			if !found {
				switch i {
				case 0:
					plansDiff.MissingFromB = append(plansDiff.MissingFromB, plan1)
				case 1:
					plansDiff.MissingFromA = append(plansDiff.MissingFromA, plan1)
				}
			}
		}
		// Switch
		if i == 0 {
			Plans1, Plans2 = Plans2, Plans1
		}
	}
	return plansDiff
}
func comparePlans(a, b InternalPlan) bool {

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
func getPlans(namespace string, matchLabels map[string]string, c client.Client) (*PlanList, error) {
	plans := &PlanList{}
	opts := client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), &opts, plans)
	return plans, err
}
func newInternalPlanFromPlan(plan Plan, c client.Client) (*InternalPlan, error) {

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
			internalLimit, err := newInternalLimitFromLimit(limit, c)
			if err != nil {
				//TODO: UPDATE STATUS OBJECT
				log.Printf("limit %s couldn't be converted: %s", limit.Name, err)
			} else {
				internalPlan.Limits = append(internalPlan.Limits, *internalLimit)
			}
		}
	}
	return &internalPlan, nil
}
