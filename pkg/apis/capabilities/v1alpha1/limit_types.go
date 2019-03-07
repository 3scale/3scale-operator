package v1alpha1

import (
	"context"
	"fmt"
	portaClient "github.com/3scale/3scale-porta-go-client/client"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LimitSpec defines SPEC of the limit object, contains LimitBase and the LimitObjectRef
type LimitSpec struct {
	LimitBase      `json:",inline"`
	LimitObjectRef `json:",inline"`
}

// LimitBase contains the limit period and the max value for said period.
type LimitBase struct {
	Period   string `json:"period"`
	MaxValue int64  `json:"maxValue"`
}

// LimitObjectRef contains he Metric ObjectReference
type LimitObjectRef struct {
	Metric v1.ObjectReference `json:"metricRef"`
}

// LimitStatus defines the observed state of Limit
type LimitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Limit is the Schema for the limits API object
// +k8s:openapi-gen=true
type Limit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LimitSpec   `json:"spec,omitempty"`
	Status LimitStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LimitList contains a list of Limit
type LimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Limit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Limit{}, &LimitList{})
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
	if limit.Spec.Metric.Name == "Hits" || limit.Spec.Metric.Name == "hits" {
		metric.Name = "hits"
	} else {
		err := c.Get(context.TODO(), reference, metric)
		if err != nil {
			// Something is broken
			return nil, err
		}
	}

	internalLimit = InternalLimit{
		Name:     limit.Name,
		Period:   limit.Spec.Period,
		MaxValue: limit.Spec.MaxValue,
		Metric:   metric.Name,
	}

	return &internalLimit, nil
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

type InternalLimit struct {
	Name     string `json:"name"`
	Period   string `json:"period"`
	MaxValue int64  `json:"maxValue"`
	Metric   string `json:"metric"`
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
func getLimits(namespace string, matchLabels map[string]string, c client.Client) (*LimitList, error) {
	limits := &LimitList{}
	opts := client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), &opts, limits)
	return limits, err
}
