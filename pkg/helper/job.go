package helper

import (
	"context"
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// HasJobImageChanged returns true if the Job's existing image is different from the provided image
func HasJobImageChanged(jName string, jNamespace string, desiredImage string, client k8sclient.Client) (bool, error) {
	job := &batchv1.Job{}
	err := client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: jNamespace,
		Name:      jName,
	}, job)

	// Return error if can't get Job
	if err != nil && !k8serr.IsNotFound(err) {
		return false, fmt.Errorf("error getting job %s: %w", job.Name, err)
	}

	// Return false if the Job doesn't exist yet
	if k8serr.IsNotFound(err) {
		return false, nil
	}

	// Return true if the Job is using the old image
	if job.Spec.Template.Spec.Containers[0].Image != desiredImage {
		return true, nil
	}

	return false, nil
}

func DeleteJob(jName string, jNamespace string, client k8sclient.Client) error {
	job := &batchv1.Job{}
	err := client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: jNamespace,
		Name:      jName,
	}, job)

	// Breakout if the Job has already been deleted
	if k8serr.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting job %s: %v", job.Name, err)
	}

	// Return error if Job is currently running
	if !HasJobCompleted(job.Name, job.Namespace, client) {
		return fmt.Errorf("won't delete job %s because it's still running", job.Name)
	}

	deleteOptions := []k8sclient.DeleteOption{
		k8sclient.GracePeriodSeconds(int64(0)),
		k8sclient.PropagationPolicy(metav1.DeletePropagationForeground),
	}

	// Delete the Job
	err = client.Delete(context.TODO(), job, deleteOptions...)
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return fmt.Errorf("error deleting job %s: %v", job.Name, err)
		}
	}

	return nil
}
