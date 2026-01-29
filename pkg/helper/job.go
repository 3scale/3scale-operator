package helper

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SystemAppRevisionAnnotation = "apimanager.apps.3scale.net/system-app-deployment-revision"
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

func lookupJob(ctx context.Context, job k8sclient.Object, client k8sclient.Client) (*batchv1.Job, error) {
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
	lookup, err := lookupJob(ctx, job, client)
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

// HasAppRevisionChanged returns true if the system-app Deployment's revision doesn't match the Job's annotation tracking it
func HasAppRevisionChanged(jName string, revision int64, namespace string, client k8sclient.Client) (bool, error) {
	trackedRevision, err := GetAppRevision(jName, namespace, client)
	if err != nil {
		return false, err
	}

	// Job doesn't exist - no change
	if trackedRevision == 0 {
		return false, nil
	}

	// Job exists but has no annotation - default to revision 1
	if trackedRevision == -1 {
		trackedRevision = 1
	}

	// Return true if the Deployment's version doesn't match the version tracked in the Job's annotation
	return trackedRevision != revision, nil
}

func DeleteJob(ctx context.Context, jName string, jNamespace string, client k8sclient.Client) error {
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

	// Return specific error if Job is currently running
	if !HasJobCompleted(ctx, job, client) {
		return &JobStillRunningError{JobName: jName}
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

func JobExists(ctx context.Context, job k8sclient.Object, client k8sclient.Client) (bool, error) {
	_, err := lookupJob(ctx, job, client)

	if err == nil {
		return true, nil
	}

	if k8serr.IsNotFound(err) {
		return false, nil
	}

	return false, err
}

// GetAppRevision returns the revision annotation value from a job
// Returns 0 if the job doesn't exist
// Returns -1 if the job exists but has no revision annotation (corrupted/unknown state)
// Returns the revision number if annotation exists
func GetAppRevision(jobName, namespace string, client k8sclient.Client) (int64, error) {
	job := &batchv1.Job{}
	err := client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: namespace,
		Name:      jobName,
	}, job)

	// Return 0 if the Job doesn't exist
	if k8serr.IsNotFound(err) {
		return 0, nil
	}

	// Return error if can't get Job
	if err != nil {
		return 0, fmt.Errorf("error getting job %s: %w", jobName, err)
	}

	// Parse the revision from job annotations
	if job.Annotations != nil {
		if revisionStr, ok := job.Annotations[SystemAppRevisionAnnotation]; ok {
			revision, err := strconv.ParseInt(revisionStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse revision from job %s annotations: %w", jobName, err)
			}
			return revision, nil
		}
	}

	// Return -1 if job exists but has no revision annotation
	// This signals corrupted/unknown state that needs deletion and recreation
	return -1, nil
}
