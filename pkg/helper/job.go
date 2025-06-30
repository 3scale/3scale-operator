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
	SystemAppRevisionAnnotation = "system-app-deployment-generation"
)

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
	job := &batchv1.Job{}
	err := client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: namespace,
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

	// Parse the Job's observed Deployment revision from its annotations
	var trackedRevision int64 = 1
	if job.Annotations != nil {
		for key, val := range job.Annotations {
			if key == SystemAppRevisionAnnotation {
				trackedRevision, err = strconv.ParseInt(val, 10, 64)
				if err != nil {
					return false, fmt.Errorf("failed to parse system-app Deployment's revision from job %s annotations: %w", job.Name, err)
				}
			}
		}
	}

	// Return true if the Deployment's version doesn't match the version tracked in the Job's annotation
	if trackedRevision != revision {
		return true, nil
	}

	return false, nil
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

	// Return error if Job is currently running
	if !HasJobCompleted(ctx, job, client) {
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
