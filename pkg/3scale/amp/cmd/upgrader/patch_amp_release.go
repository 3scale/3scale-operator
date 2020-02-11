package upgrader

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PatchAMPRelese(cl client.Client, ns string) error {
	changed := false
	existing := &v1.ConfigMap{}
	configMapNamespacedName := types.NamespacedName{Name: "system-environment", Namespace: ns}
	err := cl.Get(context.TODO(), configMapNamespacedName, existing)
	if err != nil {
		return err
	}

	ampRelease := existing.Data["AMP_RELEASE"]

	if ampRelease != "2.8" {
		existing.Data["AMP_RELEASE"] = "2.8"
		changed = true
	}

	if changed {
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
	}

	return nil
}
