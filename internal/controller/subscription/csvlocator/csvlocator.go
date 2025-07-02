package csvlocator

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-registry/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type CSVLocator interface {
	GetCSV(ctx context.Context, client k8sclient.Client, installPlan *operatorsv1alpha1.InstallPlan) (*operatorsv1alpha1.ClusterServiceVersion, error)
}

type EmbeddedCSVLocator struct{}

var _ CSVLocator = &EmbeddedCSVLocator{}

func (l *EmbeddedCSVLocator) GetCSV(ctx context.Context, client k8sclient.Client, installPlan *operatorsv1alpha1.InstallPlan) (*operatorsv1alpha1.ClusterServiceVersion, error) {
	csv := &operatorsv1alpha1.ClusterServiceVersion{}

	// The latest CSV is only represented in the new install plan while the upgrade is pending approval
	for _, installPlanResources := range installPlan.Status.Plan {
		if installPlanResources.Resource.Kind == operatorsv1alpha1.ClusterServiceVersionKind {
			err := json.Unmarshal([]byte(installPlanResources.Resource.Manifest), &csv)
			if err != nil {
				return csv, fmt.Errorf("failed to unmarshal json: %w", err)
			}
		}
	}

	return csv, nil
}

type ConfigMapCSVLocator struct{}

var _ CSVLocator = &ConfigMapCSVLocator{}

type unpackedBundleReference struct {
	Kind                   string `json:"kind"`
	Name                   string `json:"name"`
	Namespace              string `json:"namespace"`
	CatalogSourceName      string `json:"catalogSourceName"`
	CatalogSourceNamespace string `json:"catalogSourceNamespace"`
	Replaces               string `json:"replaces"`
}

func (l *ConfigMapCSVLocator) GetCSV(ctx context.Context, client k8sclient.Client, installPlan *operatorsv1alpha1.InstallPlan) (*operatorsv1alpha1.ClusterServiceVersion, error) {
	csv := &operatorsv1alpha1.ClusterServiceVersion{}

	// The latest CSV is only represented in the new install plan while the upgrade is pending approval
	for _, installPlanResources := range installPlan.Status.Plan {
		if installPlanResources.Resource.Kind != operatorsv1alpha1.ClusterServiceVersionKind {
			continue
		}

		// Get the reference to the ConfigMap that contains the CSV
		ref := &unpackedBundleReference{}
		err := json.Unmarshal([]byte(installPlanResources.Resource.Manifest), &ref)
		if err != nil {
			return csv, fmt.Errorf("failed to unmarshal json: %w", err)
		}

		// Get the ConfigMap
		csvConfigMap := &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      ref.Name,
				Namespace: ref.Namespace,
			},
		}
		if err = client.Get(ctx, k8sclient.ObjectKey{Name: csvConfigMap.Name, Namespace: csvConfigMap.Namespace}, csvConfigMap); err != nil {
			return csv, fmt.Errorf("error retrieving ConfigMap %s/%s: %v", ref.Namespace, ref.Name, err)
		}

		for _, resourceByte := range csvConfigMap.BinaryData {
			// Decode base64 to string and compress before decompressing from gzip
			compressedData, err := base64.StdEncoding.DecodeString(string(resourceByte))
			if err != nil {
				return nil, fmt.Errorf("failed to decode base64: %s", err)
			}

			// Decompress from gzip
			reader, err := gzip.NewReader(bytes.NewBuffer(compressedData))
			if err != nil {
				return nil, err
			}

			result, err := io.ReadAll(reader)
			if err != nil {
				return nil, err
			}

			csvCandidate, err := getCSVfromCM(string(result))
			if csvCandidate == nil {
				continue
			}
			if err != nil {
				return nil, err
			}

			err = reader.Close()
			if err != nil {
				return nil, err
			}

			csv = csvCandidate
		}
	}

	return csv, nil
}

func getCSVfromCM(resourceStr string) (*operatorsv1alpha1.ClusterServiceVersion, error) {
	csv := &operatorsv1alpha1.ClusterServiceVersion{}

	// Decode the manifest
	reader := strings.NewReader(resourceStr)
	resource, decodeErr := registry.DecodeUnstructured(reader)
	if decodeErr != nil {
		return csv, decodeErr
	}

	// If the kind is not CSV, skip it
	if resource.GetKind() != operatorsv1alpha1.ClusterServiceVersionKind {
		return nil, nil
	}

	// Encode the unstructured CSV as Json to decode it back to the
	// structured object
	resourceJSON, err := resource.MarshalJSON()
	if err != nil {
		return csv, err
	}

	if err := json.Unmarshal(resourceJSON, csv); err != nil {
		return csv, fmt.Errorf("failed to unmarshall yaml: %v", err)
	}
	return csv, nil
}

type CachedCSVLocator struct {
	cache map[string]*operatorsv1alpha1.ClusterServiceVersion

	locator CSVLocator
}

var _ CSVLocator = &CachedCSVLocator{}

func NewCachedCSVLocator(innerLocator CSVLocator) *CachedCSVLocator {
	return &CachedCSVLocator{
		cache:   map[string]*operatorsv1alpha1.ClusterServiceVersion{},
		locator: innerLocator,
	}
}

func (l *CachedCSVLocator) GetCSV(ctx context.Context, client k8sclient.Client, installPlan *operatorsv1alpha1.InstallPlan) (*operatorsv1alpha1.ClusterServiceVersion, error) {
	key := fmt.Sprintf("%s/%s", installPlan.Namespace, installPlan.Name)

	if found, ok := l.cache[key]; ok {
		return found, nil
	}

	csv, err := l.locator.GetCSV(ctx, client, installPlan)
	if err != nil {
		return nil, err
	}

	if csv != nil {
		l.cache[key] = csv
	}

	return csv, nil
}

type ConditionalCSVLocator struct {
	Condition func(installPlan *operatorsv1alpha1.InstallPlan) CSVLocator
}

func NewConditionalCSVLocator(condition func(installPlan *operatorsv1alpha1.InstallPlan) CSVLocator) *ConditionalCSVLocator {
	return &ConditionalCSVLocator{
		Condition: condition,
	}
}

var _ CSVLocator = &ConditionalCSVLocator{}

func (l *ConditionalCSVLocator) GetCSV(ctx context.Context, client k8sclient.Client, installPlan *operatorsv1alpha1.InstallPlan) (*operatorsv1alpha1.ClusterServiceVersion, error) {
	locator := l.Condition(installPlan)
	if locator == nil {
		return nil, fmt.Errorf("no csvlocator found for installplan %s", installPlan.Name)
	}

	return locator.GetCSV(ctx, client, installPlan)
}

func SwitchLocators(conditions ...func(*operatorsv1alpha1.InstallPlan) CSVLocator) func(*operatorsv1alpha1.InstallPlan) CSVLocator {
	return func(installPlan *operatorsv1alpha1.InstallPlan) CSVLocator {
		for _, condition := range conditions {
			if locator := condition(installPlan); locator != nil {
				return locator
			}
		}

		return nil
	}
}

func ForReference(installPlan *operatorsv1alpha1.InstallPlan) CSVLocator {
	for _, installPlanResources := range installPlan.Status.Plan {
		if installPlanResources.Resource.Kind != operatorsv1alpha1.ClusterServiceVersionKind {
			continue
		}

		// Get the reference to the ConfigMap that contains the CSV
		ref := &unpackedBundleReference{}
		err := json.Unmarshal([]byte(installPlanResources.Resource.Manifest), &ref)
		if err != nil || ref.Name == "" || ref.Namespace == "" {
			return nil
		}

		return &ConfigMapCSVLocator{}
	}

	return nil
}

func ForEmbedded(installPlan *operatorsv1alpha1.InstallPlan) CSVLocator {
	csv := &operatorsv1alpha1.ClusterServiceVersion{}

	// The latest CSV is only represented in the new install plan while the upgrade is pending approval
	for _, installPlanResources := range installPlan.Status.Plan {
		if installPlanResources.Resource.Kind == operatorsv1alpha1.ClusterServiceVersionKind {
			err := json.Unmarshal([]byte(installPlanResources.Resource.Manifest), &csv)
			if err != nil || csv.Name == "" || csv.Namespace == "" {
				return nil
			}

			return &EmbeddedCSVLocator{}
		}
	}

	return nil
}
