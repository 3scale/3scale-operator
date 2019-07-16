package operator

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/common"
)

func ObjectInfo(obj common.KubernetesObject) string {
	return fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
}
