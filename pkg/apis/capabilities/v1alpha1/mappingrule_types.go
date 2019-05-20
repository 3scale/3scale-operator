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
	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MappingRuleSpec defines the desired state of MappingRule
// +k8s:openapi-gen=true
type MappingRuleSpec struct {
	MappingRuleBase      `json:",inline"`
	MappingRuleMetricRef `json:",inline"`
}

type MappingRuleBase struct {
	Path      string `json:"path"`
	Method    string `json:"method"`
	Increment int64  `json:"increment"`
}

type MappingRuleMetricRef struct {
	MetricRef v1.ObjectReference `json:"metricRef"`
}

// MappingRuleStatus defines the observed state of MappingRule
// +k8s:openapi-gen=true
type MappingRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MappingRule is the Schema for the mappingrules API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type MappingRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MappingRuleSpec   `json:"spec,omitempty"`
	Status MappingRuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MappingRuleList contains a list of MappingRule
type MappingRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MappingRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MappingRule{}, &MappingRuleList{})
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
func getMappingRules(namespace string, matchLabels map[string]string, c client.Client) (*MappingRuleList, error) {
	mappingRules := &MappingRuleList{}
	opts := client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), &opts, mappingRules)
	return mappingRules, err
}
func getServiceMappingRulesFrom3scale(c *portaClient.ThreeScaleClient, service portaClient.Service) (*[]InternalMappingRule, error) {

	var mappingRules []InternalMappingRule
	mappingRulesFrom3scale, _ := c.ListMappingRule(service.ID)

	for _, mapping := range mappingRulesFrom3scale.MappingRules {

		desiredMetricName := ""
		for _, metric := range service.Metrics.Metrics {
			if metric.ID == mapping.MetricID {
				desiredMetricName = metric.FriendlyName
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
func newInternalMappingRuleFromMappingRule(mappingRule MappingRule, c client.Client) (*InternalMappingRule, error) {
	// GET metric for mapping rule.
	metric := &Metric{}
	var namespace string

	// Handle metrics Hits.
	if mappingRule.Spec.MetricRef.Name == "Hits" ||
		mappingRule.Spec.MetricRef.Name == "hits" {
		metric.Name = "Hits"

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

type InternalMappingRule struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	Increment int64  `json:"increment"`
	Metric    string `json:"metric"`
}

//TODO: Refactor Diffs.
type MappingRuleDiff struct {
	MissingFromA []InternalMappingRule
	MissingFromB []InternalMappingRule
	Equal        []InternalMappingRule
	NotEqual     []MappingRulePair
}
type MappingRulePair struct {
	A InternalAPI
	B InternalAPI
}

func diffMappingRules(mappingRules1, mappingRules2 []InternalMappingRule) MappingRuleDiff {
	var mappingRuleDiff MappingRuleDiff

	if len(mappingRules2) == 0 {
		mappingRuleDiff.MissingFromB = mappingRules1
		return mappingRuleDiff
	}

	for i := 0; i < 2; i++ {
		for _, mappingRule1 := range mappingRules1 {
			found := false
			for _, mappingRule2 := range mappingRules2 {
				if mappingRule1.Method == mappingRule2.Method &&
					mappingRule1.Increment == mappingRule2.Increment &&
					mappingRule1.Path == mappingRule2.Path &&
					mappingRule1.Metric == mappingRule2.Metric {

					found = true
					break
				}
			}
			if !found {
				switch i {
				case 0:
					mappingRuleDiff.MissingFromB = append(mappingRuleDiff.MissingFromB, mappingRule1)
				case 1:
					mappingRuleDiff.MissingFromA = append(mappingRuleDiff.MissingFromA, mappingRule1)
				}
			}
		}
		if i == 0 {
			mappingRules1, mappingRules2 = mappingRules2, mappingRules1
		}
	}

	return mappingRuleDiff
}

func (m MappingRuleDiff) reconcileWith3scale(c *portaClient.ThreeScaleClient, serviceId string, api InternalAPI) error {
	for _, mappingRule := range m.MissingFromB {
		metric, err := metricNametoMetric(c, serviceId, mappingRule.Metric)
		if err != nil {
			return err
		}
		_, err = c.CreateMappingRule(serviceId, strings.ToUpper(mappingRule.Method), mappingRule.Path, int(mappingRule.Increment), metric.ID)
		if err != nil {
			return err
		}
	}

	for _, mappingRule := range m.MissingFromA {
		mappingRule, err := get3scaleMappingRulefromInternalMappingRule(c, serviceId, mappingRule)
		if err != nil {
			return err
		}
		err = c.DeleteMappingRule(serviceId, mappingRule.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
