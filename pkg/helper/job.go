package helper

import (
	"context"
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// UIDBasedJobName returns a Job name that is compromised of the provided prefix,
// a hyphen and the provided uid: "<prefix>-<uid>". The returned name is a
// DNS1123 Label compliant name. Due to UIDs are 36 characters of length this
// means that the maximum prefix lenght that can be provided is of 26
// characters. If the generated name is not DNS1123 compliant an error is
// returned
func UIDBasedJobName(prefix string, uid types.UID) (string, error) {
	uidStr := string(uid)
	jobName := fmt.Sprintf("%s-%s", prefix, uidStr)
	errStrings := validation.IsDNS1123Label(jobName)
	var err error

	if len(errStrings) > 0 {
		err = fmt.Errorf("Error generating UID-based K8s Job Name: '%s'", strings.Join(errStrings, "\n"))
	}

	return jobName, err
}

// HasJobCompleted checks if the provided Job has completed
func HasJobCompleted(jName string, jNamespace string, client k8sclient.Client) bool {
	job := &batchv1.Job{}
	err := client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: jNamespace,
		Name:      jName,
	}, job)

	// Return false on error
	if err != nil {
		return false
	}

	// Check if Job has completed
	jobConditions := job.Status.Conditions
	for _, condition := range jobConditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
