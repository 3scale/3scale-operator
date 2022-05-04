package helper

import (
	"context"
	"fmt"

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
	// If the version could not be found, return the default value
	if !ok {
		return sumRate, nil
	}

	// If the version is less than 4.9, use sum_rate
	if comparison < 0 {
		sumRate = "sum_rate"
	}

	return sumRate, nil
}
