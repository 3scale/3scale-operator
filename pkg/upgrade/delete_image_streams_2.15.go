package upgrade

import (
	"context"
	"fmt"

	imagev1 "github.com/openshift/api/image/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// DeleteImageStreams deletes the ImageStream objects managed by APIManager as they are no longer needed with Deployments
// 3scale 2.14 -> 2.15
func DeleteImageStreams(namespace string, client k8sclient.Client) error {
	var imageStreams = []string{
		"amp-apicast",
		"amp-backend",
		"amp-system",
		"amp-zync",
		"backend-redis",
		"system-memcached",
		"system-mysql",
		"system-redis",
		"system-searchd",
		"zync-database-postgresql",
	}

	// First check if any ImageStreams exist, if they've already been deleted then break out
	imageStreamList := &imagev1.ImageStreamList{}
	listOps := []k8sclient.ListOption{
		k8sclient.InNamespace(namespace),
	}
	err := client.List(context.TODO(), imageStreamList, listOps...)
	if err != nil {
		return fmt.Errorf("failed to list ImageStreams: %v", err)
	}
	if len(imageStreamList.Items) == 0 {
		return nil
	}

	// Delete the specified ImageStreams
	for _, imageStreamName := range imageStreams {
		imageStream := &imagev1.ImageStream{
			ObjectMeta: metav1.ObjectMeta{
				Name:      imageStreamName,
				Namespace: namespace,
			},
		}

		err := client.Delete(context.TODO(), imageStream)
		if err != nil {
			if !k8serr.IsNotFound(err) {
				return fmt.Errorf("error deleting ImageStream %s: %v", imageStream.Name, err)
			}
		}

	}

	return nil
}
