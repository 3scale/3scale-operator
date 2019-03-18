package v1alpha1

import (
	"context"
	"fmt"
	portaClient "github.com/3scale/3scale-porta-go-client/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MetricSpec defines the desired state of Metric
type MetricSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Unit           string `json:"unit"`
	Description    string `json:"description"`
	IncrementsHits bool   `json:"incrementHits"`
}

// MetricStatus defines the observed state of Metric
type MetricStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Metric is the Schema for the metrics API
// +k8s:openapi-gen=true
type Metric struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetricSpec   `json:"spec,omitempty"`
	Status MetricStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetricList contains a list of Metric
type MetricList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Metric `json:"items"`
}

type InternalMetric struct {
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description"`
}

type MetricsDiff struct {
	MissingFromA []InternalMetric
	MissingFromB []InternalMetric
	Equal        []InternalMetric
	NotEqual     []MetricsPair
}
type MetricsPair struct {
	A InternalMetric
	B InternalMetric
}

func diffMetrics(metrics1, metrics2 []InternalMetric) MetricsDiff {

	var metricsDiff MetricsDiff

	if len(metrics2) == 0 {
		metricsDiff.MissingFromB = metrics1
		return metricsDiff
	}

	for i := 0; i < 2; i++ {
		for _, metric1 := range metrics1 {
			found := false
			for _, metric2 := range metrics2 {
				if metric2.Name == metric1.Name {
					if i == 0 {
						if metric1 == metric2 {
							metricsDiff.Equal = append(metricsDiff.Equal, metric1)
						} else {
							metricPair := MetricsPair{
								A: metric1,
								B: metric2,
							}
							metricsDiff.NotEqual = append(metricsDiff.NotEqual, metricPair)
						}
					}
					found = true
					break
				}
			}
			if !found {
				switch i {
				case 0:
					metricsDiff.MissingFromB = append(metricsDiff.MissingFromB, metric1)
				case 1:
					metricsDiff.MissingFromA = append(metricsDiff.MissingFromA, metric1)
				}
			}

		}
		if i == 0 {
			metrics1, metrics2 = metrics2, metrics1
		}
	}
	return metricsDiff
}
func (d *MetricsDiff) ReconcileWith3scale(c *portaClient.ThreeScaleClient, serviceId string, api InternalAPI) error {

	for _, metric := range d.MissingFromB {
		err := createInternalMetricIn3scale(c, api, metric)
		if err != nil {
			return err
		}
	}

	for _, metric := range d.MissingFromA {
		err := deleteInternalMetricFrom3scale(c, api, metric)
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
func getMetrics(namespace string, matchLabels map[string]string, c client.Client) (*MetricList, error) {
	metrics := &MetricList{}
	opts := client.ListOptions{}
	opts.InNamespace(namespace)
	opts.MatchingLabels(matchLabels)
	err := c.List(context.TODO(), &opts, metrics)
	return metrics, err
}

func metricNametoMetric(c *portaClient.ThreeScaleClient, serviceID string, metricName string) (portaClient.Metric, error) {
	m := portaClient.Metric{}
	metrics, err := c.ListMetrics(serviceID)
	if err != nil {
		return m, err
	}

	for _, metric := range metrics.Metrics {
		if metricName == metric.FriendlyName {
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
func createInternalMetricIn3scale(c *portaClient.ThreeScaleClient, api InternalAPI, metric InternalMetric) error {

	service, err := getServiceFromInternalAPI(c, api.Name)
	if err != nil {
		return err
	}
	_, err = c.CreateMetric(service.ID, metric.Name, metric.Description, metric.Unit)
	return err
}
func newInternalMetricFromMetric(metric Metric) *InternalMetric {
	internalMetric := InternalMetric{
		Name:        metric.Name,
		Unit:        metric.Spec.Unit,
		Description: metric.Spec.Description,
	}

	return &internalMetric
}

func deleteInternalMetricFrom3scale(c *portaClient.ThreeScaleClient, api InternalAPI, metric InternalMetric) error {

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

func init() {
	SchemeBuilder.Register(&Metric{}, &MetricList{})
}
