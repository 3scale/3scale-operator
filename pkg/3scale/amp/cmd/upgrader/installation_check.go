package upgrader

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CheckCurrentInstallation(cl client.Client, ns string) error {
	existing := &v1.ConfigMap{}
	err := cl.Get(
		context.TODO(),
		types.NamespacedName{Name: "system-environment", Namespace: ns},
		existing)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("3scale installation not found in %s. Maybe different namespace?", ns)
		}
		return err
	}

	ampRelease, exists := existing.Data["AMP_RELEASE"]
	if !exists {
		return errors.New("AMP_RELEASE not found in system-environment configmap. Corrupted installation?")
	}

	if ampRelease != "2.7" && ampRelease != "2.8" {
		return fmt.Errorf("3scale release %s installation found. This tool is meant to upgrade from release 2.7", ampRelease)
	}
	return nil
}
