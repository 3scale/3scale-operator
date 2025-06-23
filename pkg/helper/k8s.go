package helper

import (
	"fmt"
	"os"
	"strings"
)

const (
	serviceAccountDir = "/var/run/secrets/kubernetes.io/serviceaccount"
)

// GetWatchNamespace returns the Namespace the operator should be watching for changes
func GetWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	watchNamespaceEnvVar := "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}

// IsRunLocally checks if the operator is run locally
func IsRunLocally() bool {
	return !IsRunInCluster()
}

// IsRunInCluster checks if the operator is run in cluster
func IsRunInCluster() bool {
	_, err := os.Stat(serviceAccountDir)
	if err == nil {
		return true
	}

	return !os.IsNotExist(err)
}

// GetOperatorNamespace returns the namespace the operator should be running in.
func GetOperatorNamespace() (string, error) {
	if IsRunLocally() {
		ns, err := GetWatchNamespace()
		if err != nil {
			return "", fmt.Errorf("running locally but WATCH_NAMESPACE not found")
		}
		return ns, nil
	}
	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("namespace not found for current environment")
		}
		return "", err
	}
	ns := strings.TrimSpace(string(nsBytes))
	return ns, nil
}

func IsPreflightBypassed() bool {
	skipPreflightsValue := GetEnvVar("PREFLIGHT_CHECKS_BYPASS", "false")
	return skipPreflightsValue == "true"
}
