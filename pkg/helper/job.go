package helper

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	k8sappsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SystemAppGenerationAnnotation = "system-app-deployment-generation"
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

// HasAppGenerationChanged returns true if the system-app Deployment's generation doesn't match the Job's annotation tracking it
func HasAppGenerationChanged(jName string, dName string, namespace string, client k8sclient.Client) (bool, error) {
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

	deployment := &k8sappsv1.Deployment{}
	err = client.Get(context.TODO(), k8sclient.ObjectKey{
		Namespace: namespace,
		Name:      dName,
	}, deployment)
	// Return error if can't get Deployment
	if err != nil && !k8serr.IsNotFound(err) {
		return false, fmt.Errorf("error getting deployment %s: %w", deployment.Name, err)
	}
	// Return false if the Deployment doesn't exist yet
	if k8serr.IsNotFound(err) {
		return false, nil
	}

	// Parse the Job's observed Deployment generation from its annotations
	var trackedGeneration int64 = 1
	if job.Annotations != nil {
		for key, val := range job.Annotations {
			if key == SystemAppGenerationAnnotation {
				trackedGeneration, err = strconv.ParseInt(val, 10, 64)
				if err != nil {
					return false, fmt.Errorf("failed to parse system-app Deployment's generation from job %s annotations: %w", job.Name, err)
				}
			}
		}
	}

	// Return true if the Deployment's version doesn't match the version tracked in the Job's annotation
	if trackedGeneration != deployment.ObjectMeta.Generation {
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
