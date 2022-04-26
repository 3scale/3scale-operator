package helper

import (
	"context"
	"fmt"
	"reflect"

	configv1 "github.com/openshift/api/config/v1"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOpenshiftVersion(ctx context.Context, client client.Client) (string, bool, error) {
	clusterVersion := &configv1.ClusterVersion{}

	if err := client.Get(ctx, types.NamespacedName{
		Name: "version",
	}, clusterVersion); err != nil {
		if errors.IsNotFound(err) {
			return "", false, nil
		}

		return "", false, err
	}

	return clusterVersion.Status.Desired.Version, true, nil
}

func CompareOpenshiftVersion(ctx context.Context, client client.Client, version string) (int, bool, error) {
	currentVersion, ok, err := GetOpenshiftVersion(ctx, client)
	if !ok || err != nil {
		return 0, ok, err
	}

	return semver.Compare(fmt.Sprintf("v%s", currentVersion), fmt.Sprintf("v%s", version)), true, nil
}

func SumRateForOpenshiftVersion(ctx context.Context, client client.Client) (string, error) {
	sumRate := "sum_irate"

	// Compare the current Openshft version to 4.9
	comparison, ok, err := CompareOpenshiftVersion(ctx, client, "4.9")
	if err != nil {
		return "", err
	}
	// If the version could not be found, return the default mutation
	if !ok {
		return sumRate, nil
	}

	// If the version is less than 4.9, use sum_rate
	if comparison < 0 {
		sumRate = "sum_rate"
	}

	return sumRate, nil
}

func SumRateTemplateDataMutation(ctx context.Context, client client.Client) (TemplateDataMutation, error) {
	sumRate, err := SumRateForOpenshiftVersion(ctx, client)
	if err != nil {
		return nil, err
	}

	return AddSumRateField(sumRate), nil
}

// AddSumRateField creates a templateDataMutation that constructs a struct
// identical to the original with the added `SumRate` field, containing the
// value of sumRate
func AddSumRateField(sumRate string) TemplateDataMutation {
	return func(data interface{}) interface{} {
		// Get the type and runtime value of the original data
		t := reflect.TypeOf(data)
		dataValue := reflect.ValueOf(data)
		// If it's a pointer, get the element in the pointer address
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
			dataValue = dataValue.Elem()
		}

		numFields := t.NumField()

		// Create a slice with the new fields, copy the original fields
		// into it
		var offset uintptr
		fields := make([]reflect.StructField, numFields+1)
		for i := 0; i < numFields; i++ {
			field := t.Field(i)
			fields[i] = field
			offset += field.Type.Size()
		}

		// Create the new SumRate field
		sumRateField := reflect.StructField{
			Offset: offset,
			Name:   "SumRate",
			Type:   reflect.TypeOf(""),
		}

		// Add the field as the last element of the new struct fields
		fields[numFields] = sumRateField

		// Create the resulting struct
		v := reflect.New(reflect.StructOf(fields)).Elem()

		// Set the field values from the original struct
		for i := 0; i < numFields; i++ {
			v.Field(i).Set(dataValue.Field(i))
		}
		// Set the SumRate field value
		v.Field(numFields).Set(reflect.ValueOf(sumRate))

		// If the original data is a pointer, return a pointer to the
		// resulting struct
		if reflect.TypeOf(data).Kind() == reflect.Pointer {
			return v.Addr().Interface()
		}

		return v.Interface()
	}
}

type TemplateDataMutation func(interface{}) interface{}
