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

// JobStillRunningError is returned when attempting to delete a job that hasn't completed yet
type JobStillRunningError struct {
	JobName string
}

func (e *JobStillRunningError) Error() string {
	return fmt.Sprintf("job %s is still running and cannot be deleted", e.JobName)
}

// IsJobStillRunning checks if an error is a JobStillRunningError
func IsJobStillRunning(err error) bool {
	_, ok := err.(*JobStillRunningError)
	return ok
}

// UIDBasedJobName returns a Job name that is compromised of the provided prefix,
// a hyphen and the provided uid: "<prefix>-<uid>". The returned name is a
// DNS1123 Label compliant name. Due to UIDs are 36 characters of length this
// means that the maximum prefix length that can be provided is of 26
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

func LookupJob(ctx context.Context, job k8sclient.Object, client k8sclient.Client) (*batchv1.Job, error) {
	lookupKey := types.NamespacedName{
		Name:      job.GetName(),
		Namespace: job.GetNamespace(),
	}

	lookup := &batchv1.Job{}

	if err := client.Get(ctx, lookupKey, lookup); err != nil {
		return nil, err
	}

	return lookup, nil
}

// HasJobCompleted checks if the provided Job has completed
func HasJobCompleted(ctx context.Context, job k8sclient.Object, client k8sclient.Client) bool {
	lookup, err := LookupJob(ctx, job, client)
	// Return false on error
	if err != nil {
		return false
	}

	// Check if Job has completed
	jobConditions := lookup.Status.Conditions
	for _, condition := range jobConditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

func DeleteJob(ctx context.Context, job k8sclient.Object, client k8sclient.Client) error {
	lookup, err := LookupJob(ctx, job, client)

	// Breakout if the Job has already been deleted
	if k8serr.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting job %s: %v", job.GetName(), err)
	}

	// Return specific error if Job is currently running
	if !HasJobCompleted(ctx, lookup, client) {
		return &JobStillRunningError{JobName: lookup.GetName()}
	}

	deleteOptions := []k8sclient.DeleteOption{
		k8sclient.GracePeriodSeconds(int64(0)),
		k8sclient.PropagationPolicy(metav1.DeletePropagationForeground),
	}

	// Delete the Job
	err = client.Delete(context.TODO(), lookup, deleteOptions...)
	if err != nil {
		if !k8serr.IsNotFound(err) {
			return fmt.Errorf("error deleting job %s: %v", lookup.GetName(), err)
		}
	}

	return nil
}
