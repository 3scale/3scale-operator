package operator

import (
	"context"
	"strconv"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	k8sappsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSystemReconciler_ReconcileCreatesResources(t *testing.T) {
	ctx := context.TODO()
	apimanager := basicApimanagerSpecTestSystemOptions()

	objs := []runtime.Object{
		createCompletedPreHookJob(apimanager.Namespace),
		apimanager,
		createSystemDBSecret(apimanager.Namespace),
		createSystemRedisSecret(apimanager.Namespace),
	}

	reconciler, cl := setupTestReconciler(t, ctx, apimanager, objs)

	_, err := reconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      k8sclient.Object
	}{
		{"systemDatabase", "system-database", &v1.Secret{}},
		{"systemPVC", "system-storage", &v1.PersistentVolumeClaim{}},
		{"systemProviderService", "system-provider", &v1.Service{}},
		{"systemMasterService", "system-master", &v1.Service{}},
		{"systemDeveloperService", "system-developer", &v1.Service{}},
		{"systemMemcacheService", "system-memcache", &v1.Service{}},
		{"systemAppDeployment", "system-app", &k8sappsv1.Deployment{}},
		{"systemSideKiqDeployment", "system-sidekiq", &k8sappsv1.Deployment{}},
		{"systemCM", "system", &v1.ConfigMap{}},
		{"systemEnvironmentCM", "system-environment", &v1.ConfigMap{}},
		{"systemSMTPSecret", "system-smtp", &v1.Secret{}},
		{"systemEventsHookSecret", component.SystemSecretSystemEventsHookSecretName, &v1.Secret{}},
		{"systemMasterApicastSecret", component.SystemSecretSystemMasterApicastSecretName, &v1.Secret{}},
		{"systemSeedSecret", component.SystemSecretSystemSeedSecretName, &v1.Secret{}},
		{"systemRecaptchaSecret", component.SystemSecretSystemRecaptchaSecretName, &v1.Secret{}},
		{"systemAppSecret", component.SystemSecretSystemAppSecretName, &v1.Secret{}},
		{"systemMemcachedSecret", component.SystemSecretSystemMemcachedSecretName, &v1.Secret{}},
		{"systemMemcachedSecret", component.SystemSecretSystemMemcachedSecretName, &v1.Secret{}},
		{"systemAppPDB", "system-app", &policyv1.PodDisruptionBudget{}},
		{"systemSidekiqPDB", "system-sidekiq", &policyv1.PodDisruptionBudget{}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			obj := tc.obj
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, obj)
			// object must exist, that is all required to be tested
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}
		})
	}
}

func TestSystemReconciler_Replicas(t *testing.T) {
	const namespace = "operator-unittest"
	ctx := context.TODO()

	cases := []struct {
		testName                 string
		objName                  string
		apimanager               *appsv1alpha1.APIManager
		expectedAmountOfReplicas int32
	}{
		{
			testName:                 "system app replicas set - reconciler overrides manual changes",
			objName:                  "system-app",
			apimanager:               createSystemAPIManager(ptr.To(int64(1)), nil),
			expectedAmountOfReplicas: 1,
		},
		{
			testName:                 "system app replicas not set - reconciler preserves manual changes",
			objName:                  "system-app",
			apimanager:               createSystemAPIManager(nil, nil),
			expectedAmountOfReplicas: 5, // Expects the manually set value to be preserved
		},
		{
			testName:                 "system sidekiq replicas set - reconciler overrides manual changes",
			objName:                  "system-sidekiq",
			apimanager:               createSystemAPIManager(nil, ptr.To(int64(1))),
			expectedAmountOfReplicas: 1,
		},
		{
			testName:                 "system sidekiq replicas not set - reconciler preserves manual changes",
			objName:                  "system-sidekiq",
			apimanager:               createSystemAPIManager(nil, nil),
			expectedAmountOfReplicas: 5, // Expects the manually set value to be preserved
		},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			objs := []runtime.Object{
				tc.apimanager,
				createCompletedPreHookJob(namespace),
				createSystemDBSecret(namespace),
				createSystemRedisSecret(namespace),
			}

			reconciler, cl := setupTestReconciler(subT, ctx, tc.apimanager, objs)

			_, err := reconciler.Reconcile()
			if err != nil {
				subT.Fatal(err)
			}

			deployment := &k8sappsv1.Deployment{}
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}

			err = cl.Get(ctx, namespacedName, deployment)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			// Simulate manual/external modification of replica count to 5
			deployment.Spec.Replicas = ptr.To(int32(5))
			deployment.SetAnnotations(map[string]string{"deployment.kubernetes.io/revision": "1"})
			err = cl.Update(ctx, deployment)
			if err != nil {
				subT.Errorf("error updating deployment of %s: %v", tc.objName, err)
			}

			// Re-run reconciler: if replicas are set in spec, it overrides; if not, it preserves
			_, err = reconciler.Reconcile()
			if err != nil {
				subT.Fatal(err)
			}

			err = cl.Get(ctx, namespacedName, deployment)
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}

			if tc.expectedAmountOfReplicas != *deployment.Spec.Replicas {
				subT.Errorf("%s: expected replicas=%d, got=%d",
					tc.testName,
					tc.expectedAmountOfReplicas,
					*deployment.Spec.Replicas)
			}
		})
	}
}

// TestSystemReconciler_PreHookJobOrchestration tests the PreHook job creation and lifecycle
func TestSystemReconciler_PreHookJobOrchestration(t *testing.T) {
	const testNamespace = "operator-unittest"
	ctx := context.TODO()

	tests := []struct {
		name               string
		additionalObjects  []runtime.Object
		expectedJobExists  bool
		expectedDeployment bool
		expectedRequeue    bool
		validateJob        func(t *testing.T, client k8sclient.Client)
		validateDeployment func(t *testing.T, client k8sclient.Client)
	}{
		{
			name: "FRESH INSTALL: PreHook job created with revision=1 when no deployment exists",
			// Preconditions:
			// - No deployment exists (triggers fresh install path)
			// - No PreHook job exists
			// Expected behavior:
			// - hasSystemImageChanged() returns false (no deployment = fresh install)
			// - getSystemAppDeploymentRevision() returns 1 (default for new install)
			// - Job created with revision=1
			// - Deployment NOT created (blocked until job completes)
			additionalObjects: []runtime.Object{
				// Explicitly NO deployment - this is fresh install
			},
			expectedJobExists:  true,
			expectedDeployment: false,
			expectedRequeue:    true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("PreHook job should exist after first reconcile")
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "1" {
					t.Errorf("fresh install should use revision=1, got %q", job.Annotations[helper.SystemAppRevisionAnnotation])
				}
			},
			validateDeployment: func(t *testing.T, client k8sclient.Client) {
				if deploymentExists(t, client, component.SystemAppDeploymentName, testNamespace) {
					t.Error("system-app deployment should NOT be created while PreHook is incomplete")
				}
			},
		},
		{
			name: "FRESH INSTALL: Deployment created after PreHook completes",
			// Preconditions:
			// - No deployment exists yet (fresh install continues)
			// - PreHook job exists and is completed
			// Expected behavior:
			// - Deployment gets created
			// - Still requeues (deployment needs to become available for PostHook)
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 1),
				// Still no deployment - it will be created by reconciler
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("PreHook job should still exist")
				}
			},
			validateDeployment: func(t *testing.T, client k8sclient.Client) {
				if !deploymentExists(t, client, component.SystemAppDeploymentName, testNamespace) {
					t.Error("system-app deployment should be created after PreHook completes")
				}
			},
		},
		{
			name: "FRESH INSTALL: Deployment blocked while PreHook is running",
			// Preconditions:
			// - No deployment exists (fresh install)
			// - PreHook job exists but is NOT completed
			// Expected behavior:
			// - Deployment NOT created (blocked by incomplete job)
			// - Requeues to wait for job completion
			additionalObjects: []runtime.Object{
				createIncompleteJob(component.SystemAppPreHookJobName, testNamespace, SystemImageURL(), 1),
			},
			expectedJobExists:  true,
			expectedDeployment: false,
			expectedRequeue:    true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("PreHook job should exist")
				}
			},
			validateDeployment: func(t *testing.T, client k8sclient.Client) {
				if deploymentExists(t, client, component.SystemAppDeploymentName, testNamespace) {
					t.Error("system-app deployment should NOT be created while PreHook is incomplete")
				}
			},
		},
		{
			name: "NORMAL RECONCILE: PreHook job recreated when deployment revision changes",
			// Preconditions:
			// - Deployment exists with SAME image (normal reconcile, not upgrade)
			// - Deployment revision is 2 (changed externally, e.g., by kubectl rollout restart)
			// - PreHook job exists for OLD revision 1
			// Expected behavior:
			// - hasSystemImageChanged() returns false (same image)
			// - Old job deleted (revision mismatch)
			// - New job created for revision 2
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 1),
				createSystemAppDeployment(testNamespace, SystemImageURL(), 2, 2, true, false),
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("new PreHook job should be created for new revision")
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)
				// Job should be recreated for revision 2
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "2" {
					t.Errorf("expected job revision annotation to be '2', got %q", job.Annotations[helper.SystemAppRevisionAnnotation])
				}
				// New job should not be completed yet
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
						t.Error("newly created job should not be marked as complete")
					}
				}
			},
		},
		{
			name: "NORMAL RECONCILE: PreHook job not recreated when current",
			// Preconditions:
			// - Deployment exists with SAME image as APIManager
			// - Deployment revision is 1
			// - PreHook job exists for revision 1 and is completed
			// Expected behavior:
			// - hasSystemImageChanged() returns false (same image)
			// - getSystemAppDeploymentRevision() returns 1
			// - Job already exists for current revision - NOT recreated
			// - No requeue needed (everything ready for PostHook)
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 1),
				createSystemAppDeployment(testNamespace, SystemImageURL(), 1, 2, true, false),
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    false,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("PreHook job should still exist")
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)
				// Job should remain with revision 1
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "1" {
					t.Errorf("expected job revision annotation to remain '1', got %q", job.Annotations[helper.SystemAppRevisionAnnotation])
				}
				// Job should remain completed
				completed := false
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
						completed = true
						break
					}
				}
				if !completed {
					t.Error("existing completed job should remain completed")
				}
			},
		},
		{
			name: "UPGRADE: PreHook job recreated with incremented revision on image change",
			// Preconditions:
			// - Deployment exists with OLD image (old-image:v1)
			// - PreHook job exists for OLD image
			// - APIManager now specifies NEW image (SystemImageURL())
			// Expected behavior:
			// - hasSystemImageChanged() returns true (old-image:v1 != SystemImageURL())
			// - Old job deleted
			// - New job created with revision+1 (becomes 2)
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, "old-image:v1", 1),
				createSystemAppDeployment(testNamespace, "old-image:v1", 1, 2, true, false),
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("PreHook job should be recreated for image upgrade")
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)
				// On image change, revision should be incremented to 2
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "2" {
					t.Errorf("expected job revision annotation to be '2' for upgrade, got %q", job.Annotations[helper.SystemAppRevisionAnnotation])
				}
			},
		},
		{
			name: "UPGRADE: Running PreHook job is NOT deleted, reconciler requeues",
			// Preconditions:
			// - Deployment exists with OLD image at revision 1
			// - PreHook job exists for OLD image at revision 1 but is STILL RUNNING
			// - APIManager specifies NEW image (triggering upgrade)
			// Expected behavior:
			// - hasSystemImageChanged() returns true (old-image:v1 != SystemImageURL())
			// - revisionChanged returns false (job revision 1 == deployment revision 1)
			// - Reconciler attempts to delete job to recreate with new image
			// - DeleteJob refuses because job is still running
			// - Reconciler requeues instead of failing
			// - Job is preserved (not deleted)
			additionalObjects: []runtime.Object{
				createIncompleteJob(component.SystemAppPreHookJobName, testNamespace, "old-image:v1", 1),
				createSystemAppDeployment(testNamespace, "old-image:v1", 1, 2, true, false),
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("running PreHook job should be preserved (not deleted)")
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)
				// Job should still be at revision 1 (not recreated)
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "1" {
					t.Errorf("running job should be preserved with original revision=1, got %q",
						job.Annotations[helper.SystemAppRevisionAnnotation])
				}
				// Job should still be incomplete (running)
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
						t.Error("job should still be running (not completed)")
					}
				}
				// Job should still have old image (not updated)
				if len(job.Spec.Template.Spec.Containers) > 0 {
					actualImage := job.Spec.Template.Spec.Containers[0].Image
					if actualImage != "old-image:v1" {
						t.Errorf("running job should preserve old image, got %q", actualImage)
					}
				}
			},
		},
		{
			name: "UPGRADE: Job already created with correct revision - preserved during upgrade",
			// Preconditions:
			// - Deployment exists at revision 1 with OLD image (not updated yet)
			// - PreHook job ALREADY created with revision 2 and NEW image (from previous reconcile)
			// - Job is running or completed
			// - APIManager specifies NEW image (upgrade in progress)
			//
			// This represents the mid-upgrade state after the job was created:
			// - First reconcile: image changed -> job created with rev=2
			// - THIS reconcile (second+): image still changed (deployment hasn't updated) -> preserve job
			//
			// Expected behavior:
			// - hasSystemImageChanged() returns true (deployment still has old image)
			// - targetRevision = 1 + 1 = 2
			// - currentJobRevision = 2 (job already has correct revision)
			// - Deletion condition: 2 != 0 && 2 != 2 -> TRUE && FALSE -> NO deletion ✓
			// - Job is preserved (not deleted and recreated)
			// - Reconciler requeues if job running, or proceeds to deployment if completed
			additionalObjects: []runtime.Object{
				// Job already created with NEW image and target revision during previous reconcile
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 2),
				// Deployment still has OLD image (hasn't been updated yet)
				createSystemAppDeployment(testNamespace, "old-image:v1", 1, 2, true, false),
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    false, // Job completed, ready to update deployment
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("PreHook job should exist")
					return
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)

				// Verify job has correct revision (should be preserved)
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "2" {
					t.Errorf("expected job to have revision '2' (preserved), got %q",
						job.Annotations[helper.SystemAppRevisionAnnotation])
				}

				// Verify job still has NEW image (not changed)
				if len(job.Spec.Template.Spec.Containers) > 0 {
					actualImage := job.Spec.Template.Spec.Containers[0].Image
					if actualImage != SystemImageURL() {
						t.Errorf("expected job to retain new image %q, got %q", SystemImageURL(), actualImage)
					}
				}

				// Verify job is completed (ready to proceed with deployment update)
				completed := false
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
						completed = true
						break
					}
				}
				if !completed {
					t.Error("job should be completed (from previous reconcile)")
				}
			},
		},
		{
			name: "UPGRADE: Job exists without annotation during upgrade - deleted and recreated",
			// Preconditions:
			// - Deployment exists at revision 2 with OLD image
			// - Job exists with OLD image but NO REVISION ANNOTATION
			// - This can happen from: manual job creation, migration from old operator, external tools
			// - APIManager specifies NEW image (triggering upgrade)
			//
			// Expected behavior AFTER FIX:
			// - hasSystemImageChanged() returns true
			// - targetRevision = 2 + 1 = 3
			// - currentJobRevision = getPreHookJobRevision() returns -1 (no annotation)
			// - Condition: -1 != 0 && -1 != 3 evaluates to TRUE
			// - Old job is DELETED
			// - New job created for revision 3 with NEW image
			// - Reconciler requeues waiting for new job to complete
			additionalObjects: []runtime.Object{
				createJobWithoutAnnotation(component.SystemAppPreHookJobName, testNamespace, "old-image:v1", true),
				createSystemAppDeployment(testNamespace, "old-image:v1", 2, 2, true, false),
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    true, // Job recreated, not completed yet
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("job should exist")
					return
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)

				// Verify job has correct annotation
				if job.Annotations == nil || job.Annotations[helper.SystemAppRevisionAnnotation] != "3" {
					t.Errorf("expected job to have revision annotation '3', got %v", job.Annotations)
				}

				// Verify job was recreated with NEW image
				if len(job.Spec.Template.Spec.Containers) > 0 {
					actualImage := job.Spec.Template.Spec.Containers[0].Image
					if actualImage != SystemImageURL() {
						t.Errorf("expected job to have new image %q, got %q", SystemImageURL(), actualImage)
					}
				}

				// Verify job is NOT completed (fresh job)
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
						t.Error("new job should not be marked as complete")
					}
				}
			},
		},
		{
			name: "UPGRADE: Job doesn't exist - creates job for first time (tests GetAppRevision default)",
			// Preconditions:
			// - Deployment exists at revision 1 with OLD image
			// - Job does NOT exist (manually deleted, or never created)
			// - APIManager specifies NEW image (triggering upgrade)
			//
			// This tests that GetAppRevision() default (0) works correctly:
			// - GetAppRevision() returns 0 (job doesn't exist)
			// - targetRevision = 1 + 1 = 2
			// - Deletion condition: 0 != 0 && 0 != 2 evaluates to FALSE
			// - No deletion attempted (correct - nothing to delete)
			// - Job created for the first time with revision 2
			//
			// Expected behavior:
			// - hasSystemImageChanged() returns true
			// - targetRevision = 2
			// - currentJobRevision = 0 (from GetAppRevision - job doesn't exist)
			// - No deletion (0 != 0 is false)
			// - New job created with revision 2
			// - Reconciler requeues waiting for job to complete
			additionalObjects: []runtime.Object{
				// NO JOB - this is the key difference
				createSystemAppDeployment(testNamespace, "old-image:v1", 1, 2, true, false),
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    true, // Job created but not completed
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("PreHook job should be created")
					return
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)

				// Verify job has correct revision annotation
				if job.Annotations == nil || job.Annotations[helper.SystemAppRevisionAnnotation] != "2" {
					t.Errorf("expected job to have revision annotation '2', got %v", job.Annotations)
				}

				// Verify job has NEW image
				if len(job.Spec.Template.Spec.Containers) > 0 {
					actualImage := job.Spec.Template.Spec.Containers[0].Image
					if actualImage != SystemImageURL() {
						t.Errorf("expected job to have new image %q, got %q", SystemImageURL(), actualImage)
					}
				}

				// Verify job is NOT completed (fresh job)
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
						t.Error("newly created job should not be marked as complete")
					}
				}
			},
		},
		{
			name: "EDGE CASE: Both image AND revision changed - old job deleted and new one created",
			// Preconditions:
			// - Deployment exists at revision 2 with OLD image (old-image:v1)
			// - PreHook job exists for revision 1 with OLD image (completed)
			// - APIManager now specifies NEW image (SystemImageURL())
			// - Both revision changed (1 -> 2) AND image changed (old -> new)
			//
			// This represents an edge case that violates the expected state machine:
			// - Normally: image changes -> job created -> job completes -> deployment updates
			// - Here: deployment already updated (rev 2) while job still at old revision (rev 1)
			// - This could happen from manual kubectl rollout or other external changes
			//
			// Expected behavior:
			// - hasSystemImageChanged() returns true (old-image:v1 != SystemImageURL())
			// - targetRevision = currentAppDeploymentRevision + 1 = 3
			// - currentJobRevision = 1 (from job annotation)
			// - Since currentJobRevision (1) != targetRevision (3), old job is deleted
			// - New job created for revision 3 with new image
			// - Reconciler requeues waiting for new job to complete
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, "old-image:v1", 1),
				createSystemAppDeployment(testNamespace, "old-image:v1", 2, 2, true, false),
			},
			expectedJobExists:  true,
			expectedDeployment: true,
			expectedRequeue:    true, // New job is not completed yet, so requeue
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPreHookJobName, testNamespace) {
					t.Error("PreHook job should exist")
				}
				job := getJob(t, client, component.SystemAppPreHookJobName, testNamespace)
				// New job should be created for revision 3 (deployment rev 2 + 1)
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "3" {
					t.Errorf("expected new job with revision '3', got %q",
						job.Annotations[helper.SystemAppRevisionAnnotation])
				}
				// New job should have the NEW image (not old image)
				if len(job.Spec.Template.Spec.Containers) > 0 {
					actualImage := job.Spec.Template.Spec.Containers[0].Image
					if actualImage != SystemImageURL() {
						t.Errorf("new job should have new image %q, got %q", SystemImageURL(), actualImage)
					}
				}
				// New job should NOT be marked as completed (it's a fresh job)
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
						t.Error("new job should not be marked as complete")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apimanager := createSystemAPIManager(nil, nil)
			reconciler, client := setupTestReconciler(t, ctx, apimanager, append(
				[]runtime.Object{
					apimanager,
					createSystemDBSecret(testNamespace),
					createSystemRedisSecret(testNamespace),
				},
				tt.additionalObjects...))

			result, err := reconciler.Reconcile()
			if err != nil {
				t.Fatalf("reconcile failed: %v", err)
			}

			// Validate expected job existence
			jobActuallyExists := jobExists(t, client, component.SystemAppPreHookJobName, testNamespace)
			if tt.expectedJobExists && !jobActuallyExists {
				t.Errorf("expected PreHook job to exist but it doesn't")
			}
			if !tt.expectedJobExists && jobActuallyExists {
				t.Errorf("expected PreHook job NOT to exist but it does")
			}

			// Validate expected deployment existence
			deploymentActuallyExists := deploymentExists(t, client, component.SystemAppDeploymentName, testNamespace)
			if tt.expectedDeployment && !deploymentActuallyExists {
				t.Errorf("expected system-app deployment to exist but it doesn't")
			}
			if !tt.expectedDeployment && deploymentActuallyExists {
				t.Errorf("expected system-app deployment NOT to exist but it does")
			}

			// Validate requeue behavior
			if tt.expectedRequeue {
				if result.RequeueAfter == 0 {
					t.Error("expected requeue but RequeueAfter is 0")
				}
			} else {
				if result.Requeue || result.RequeueAfter != 0 {
					t.Errorf("expected no requeue but got Requeue=%v, RequeueAfter=%v", result.Requeue, result.RequeueAfter)
				}
			}

			// Run custom validations
			if tt.validateJob != nil {
				tt.validateJob(t, client)
			}
			if tt.validateDeployment != nil {
				tt.validateDeployment(t, client)
			}
		})
	}
}

// TestSystemReconciler_PostHookJobOrchestration tests the PostHook job creation and lifecycle
// PostHook is simpler than PreHook because it only runs after deployment is ready
func TestSystemReconciler_PostHookJobOrchestration(t *testing.T) {
	const testNamespace = "operator-unittest"
	ctx := context.TODO()

	tests := []struct {
		name              string
		additionalObjects []runtime.Object
		expectedJobExists bool
		expectedRequeue   bool
		validateJob       func(t *testing.T, client k8sclient.Client)
	}{
		{
			name: "PostHook blocked - Deployment not ready (progressing)",
			// Preconditions:
			// - PreHook completed for revision 1
			// - Deployment exists at revision 1, Available=True, but Progressing=True
			// - Images match (imageChanged = false)
			//
			// Expected behavior:
			// - IsDeploymentProgressing() returns true
			// - systemComponentsReady = false
			// - PostHook should NOT be created
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 1),
				createSystemAppDeployment(testNamespace, SystemImageURL(), 1, 2, true, true), // available=true, progressing=true
			},
			expectedJobExists: false,
			expectedRequeue:   true, // Requeues waiting for deployment to finish rolling out
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if jobExists(t, client, component.SystemAppPostHookJobName, testNamespace) {
					t.Error("PostHook job should NOT be created while deployment is progressing")
				}
			},
		},
		{
			name: "PostHook blocked - Deployment not available",
			// Preconditions:
			// - PreHook completed
			// - Deployment exists but Available=False
			//
			// Expected behavior:
			// - !IsDeploymentAvailable() returns true
			// - systemComponentsReady = false
			// - PostHook should NOT be created
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 1),
				createSystemAppDeployment(testNamespace, SystemImageURL(), 1, 2, false, false), // available=false
			},
			expectedJobExists: false,
			expectedRequeue:   true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if jobExists(t, client, component.SystemAppPostHookJobName, testNamespace) {
					t.Error("PostHook job should NOT be created while deployment is unavailable")
				}
			},
		},
		{
			name: "PostHook blocked - PreHook not completed",
			// Preconditions:
			// - PreHook job exists but is RUNNING (not completed)
			// - Deployment ready and available
			//
			// Expected behavior:
			// - HasJobCompleted() returns false
			// - finished = false
			// - systemComponentsReady = false (via !finished check at line 277)
			// - PostHook should NOT be created
			additionalObjects: []runtime.Object{
				createIncompleteJob(component.SystemAppPreHookJobName, testNamespace, SystemImageURL(), 1), // PreHook running
				createSystemAppDeployment(testNamespace, SystemImageURL(), 1, 2, true, false),
			},
			expectedJobExists: false,
			expectedRequeue:   true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if jobExists(t, client, component.SystemAppPostHookJobName, testNamespace) {
					t.Error("PostHook job should NOT be created while PreHook is running")
				}
			},
		},
		{
			name: "PostHook blocked - Deployment missing",
			// Preconditions:
			// - PreHook completed
			// - Deployment does NOT exist
			//
			// Expected behavior:
			// - k8serr.IsNotFound(err) returns true
			// - systemComponentsReady = false
			// - PostHook should NOT be created
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 1),
				// No deployment!
			},
			expectedJobExists: false,
			expectedRequeue:   true,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if jobExists(t, client, component.SystemAppPostHookJobName, testNamespace) {
					t.Error("PostHook job should NOT be created when deployment doesn't exist")
				}
			},
		},
		{
			name: "PostHook created - All conditions met",
			// Preconditions:
			// - PreHook completed for revision 1
			// - Deployment at revision 1, Available=True, Progressing=False
			// - Images match (imageChanged = false)
			//
			// Expected behavior:
			// - All blocking conditions false
			// - systemComponentsReady = true
			// - PostHook created with current deployment revision (1)
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 1),
				createSystemAppDeployment(testNamespace, SystemImageURL(), 1, 2, true, false), // Ready
			},
			expectedJobExists: true,
			expectedRequeue:   false,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPostHookJobName, testNamespace) {
					t.Error("PostHook job should be created when all conditions met")
					return
				}
				job := getJob(t, client, component.SystemAppPostHookJobName, testNamespace)

				// Verify job has correct revision (matches deployment)
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "1" {
					t.Errorf("expected PostHook job with revision '1', got %q",
						job.Annotations[helper.SystemAppRevisionAnnotation])
				}

				// Verify job has correct image
				if len(job.Spec.Template.Spec.Containers) > 0 {
					actualImage := job.Spec.Template.Spec.Containers[0].Image
					if actualImage != SystemImageURL() {
						t.Errorf("expected PostHook job with image %q, got %q", SystemImageURL(), actualImage)
					}
				}
			},
		},
		{
			name: "PostHook recreated on revision change",
			// Preconditions:
			// - PostHook exists for revision 1 (completed)
			// - Deployment updated to revision 2 (e.g., kubectl rollout restart)
			// - Deployment available and not progressing
			// - PreHook completed for revision 2
			//
			// Expected behavior:
			// - HasAppRevisionChanged() returns true
			// - Old PostHook deleted
			// - New PostHook created for revision 2
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 2),
				createCompletedJob("system-app-post", testNamespace, SystemImageURL(), 1), // Old PostHook
				createSystemAppDeployment(testNamespace, SystemImageURL(), 2, 2, true, false),
			},
			expectedJobExists: true,
			expectedRequeue:   false,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPostHookJobName, testNamespace) {
					t.Error("PostHook job should be recreated for new revision")
					return
				}
				job := getJob(t, client, component.SystemAppPostHookJobName, testNamespace)

				// Verify job was recreated with new revision
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "2" {
					t.Errorf("expected PostHook job with revision '2', got %q",
						job.Annotations[helper.SystemAppRevisionAnnotation])
				}

				// New job should not be completed yet
				for _, condition := range job.Status.Conditions {
					if condition.Type == batchv1.JobComplete && condition.Status == v1.ConditionTrue {
						t.Error("newly created PostHook job should not be marked as complete")
					}
				}
			},
		},
		{
			name: "PostHook without annotation - deleted and recreated",
			// Preconditions:
			// - PostHook exists WITHOUT revision annotation (manual creation, old operator version, etc.)
			// - Deployment at revision 1, ready and available
			// - PreHook completed with revision 1
			// - Images match (no upgrade)
			//
			// Expected behavior (self-healing):
			// - imageChanged = false (normal reconcile path)
			// - Old job is DELETED
			// - New job created for revision 1 with annotation
			additionalObjects: []runtime.Object{
				createCompletedJob("system-app-pre", testNamespace, SystemImageURL(), 1),
				createJobWithoutAnnotation("system-app-post", testNamespace, SystemImageURL(), true),
				createSystemAppDeployment(testNamespace, SystemImageURL(), 1, 1, true, false),
			},
			expectedJobExists: true,
			expectedRequeue:   false,
			validateJob: func(t *testing.T, client k8sclient.Client) {
				if !jobExists(t, client, component.SystemAppPostHookJobName, testNamespace) {
					t.Error("PostHook job should exist")
					return
				}
				job := getJob(t, client, component.SystemAppPostHookJobName, testNamespace)

				// Verify job was recreated with correct annotation
				if job.Annotations[helper.SystemAppRevisionAnnotation] != "1" {
					t.Errorf("expected PostHook job with revision '1', got %v", job.Annotations)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apimanager := createSystemAPIManager(nil, nil)
			reconciler, client := setupTestReconciler(t, ctx, apimanager, append(
				[]runtime.Object{
					apimanager,
					createSystemDBSecret(testNamespace),
					createSystemRedisSecret(testNamespace),
				},
				tt.additionalObjects...))

			result, err := reconciler.Reconcile()
			if err != nil {
				t.Fatalf("reconcile failed: %v", err)
			}

			// Validate expected job existence
			jobActuallyExists := jobExists(t, client, component.SystemAppPostHookJobName, testNamespace)
			if tt.expectedJobExists && !jobActuallyExists {
				t.Errorf("expected PostHook job to exist but it doesn't")
			}
			if !tt.expectedJobExists && jobActuallyExists {
				t.Errorf("expected PostHook job NOT to exist but it does")
			}

			// Validate requeue behavior
			if tt.expectedRequeue {
				if result.RequeueAfter == 0 && !result.Requeue {
					t.Error("expected requeue but got none")
				}
			} else {
				if result.Requeue || result.RequeueAfter != 0 {
					t.Errorf("expected no requeue but got Requeue=%v, RequeueAfter=%v", result.Requeue, result.RequeueAfter)
				}
			}

			// Run custom validations
			if tt.validateJob != nil {
				tt.validateJob(t, client)
			}
		})
	}
}

// TestGetSystemAppDeploymentRevision tests the getSystemAppDeploymentRevision helper function
func TestGetSystemAppDeploymentRevision(t *testing.T) {
	const testNamespace = "operator-unittest"

	tests := []struct {
		name               string
		existingDeployment *k8sappsv1.Deployment
		expectedRevision   int64
		expectError        bool
	}{
		{
			name:               "deployment doesn't exist - returns default revision 1",
			existingDeployment: nil,
			expectedRevision:   1,
			expectError:        false,
		},
		{
			name: "deployment exists without revision annotation - returns 0",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
					// No annotations
				},
			},
			expectedRevision: 0,
			expectError:      false,
		},
		{
			name: "deployment exists with revision annotation - returns revision value",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
					Annotations: map[string]string{
						"deployment.kubernetes.io/revision": "5",
					},
				},
			},
			expectedRevision: 5,
			expectError:      false,
		},
		{
			name: "deployment exists with revision=1",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
					Annotations: map[string]string{
						"deployment.kubernetes.io/revision": "1",
					},
				},
			},
			expectedRevision: 1,
			expectError:      false,
		},
		{
			name: "deployment exists with invalid revision annotation - returns error",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
					Annotations: map[string]string{
						"deployment.kubernetes.io/revision": "not-a-number",
					},
				},
			},
			expectedRevision: 0,
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := []runtime.Object{}
			if tt.existingDeployment != nil {
				objs = append(objs, tt.existingDeployment)
			}

			s := setupScheme(t)
			client := fake.NewClientBuilder().
				WithScheme(s).
				WithRuntimeObjects(objs...).
				Build()

			revision, err := getSystemAppDeploymentRevision(testNamespace, client)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if revision != tt.expectedRevision {
				t.Errorf("expected revision=%d, got=%d", tt.expectedRevision, revision)
			}
		})
	}
}

// TestHasSystemImageChanged tests the hasSystemImageChanged helper function
func TestHasSystemImageChanged(t *testing.T) {
	const testNamespace = "operator-unittest"

	tests := []struct {
		name               string
		existingDeployment *k8sappsv1.Deployment
		desiredImage       string
		expectedChanged    bool
		expectError        bool
	}{
		{
			name:               "no deployment exists - fresh install, returns false",
			existingDeployment: nil,
			desiredImage:       "new-image:v2",
			expectedChanged:    false,
			expectError:        false,
		},
		{
			name: "deployment exists with same image in all containers - returns false",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
				},
				Spec: k8sappsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{Name: "system-master", Image: "image:v1"},
								{Name: "system-provider", Image: "image:v1"},
								{Name: "system-developer", Image: "image:v1"},
							},
						},
					},
				},
			},
			desiredImage:    "image:v1",
			expectedChanged: false,
			expectError:     false,
		},
		{
			name: "deployment exists with different image - returns true",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
				},
				Spec: k8sappsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{Name: "system-master", Image: "old-image:v1"},
								{Name: "system-provider", Image: "old-image:v1"},
								{Name: "system-developer", Image: "old-image:v1"},
							},
						},
					},
				},
			},
			desiredImage:    "new-image:v2",
			expectedChanged: true,
			expectError:     false,
		},
		{
			name: "deployment exists with mixed images (system-master differs) - returns true",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
				},
				Spec: k8sappsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{Name: "system-master", Image: "old-image:v1"},
								{Name: "system-provider", Image: "new-image:v2"},
								{Name: "system-developer", Image: "new-image:v2"},
							},
						},
					},
				},
			},
			desiredImage:    "new-image:v2",
			expectedChanged: true,
			expectError:     false,
		},
		{
			name: "deployment exists with mixed images (system-provider differs) - returns true",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
				},
				Spec: k8sappsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{Name: "system-master", Image: "new-image:v2"},
								{Name: "system-provider", Image: "old-image:v1"},
								{Name: "system-developer", Image: "new-image:v2"},
							},
						},
					},
				},
			},
			desiredImage:    "new-image:v2",
			expectedChanged: true,
			expectError:     false,
		},
		{
			name: "deployment exists with mixed images (system-developer differs) - returns true",
			existingDeployment: &k8sappsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      component.SystemAppDeploymentName,
					Namespace: testNamespace,
				},
				Spec: k8sappsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{Name: "system-master", Image: "new-image:v2"},
								{Name: "system-provider", Image: "new-image:v2"},
								{Name: "system-developer", Image: "old-image:v1"},
							},
						},
					},
				},
			},
			desiredImage:    "new-image:v2",
			expectedChanged: true,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := []runtime.Object{}
			if tt.existingDeployment != nil {
				objs = append(objs, tt.existingDeployment)
			}

			s := setupScheme(t)
			client := fake.NewClientBuilder().
				WithScheme(s).
				WithRuntimeObjects(objs...).
				Build()

			changed, err := hasSystemImageChanged(testNamespace, tt.desiredImage, client)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if changed != tt.expectedChanged {
				t.Errorf("expected changed=%v, got=%v", tt.expectedChanged, changed)
			}
		})
	}
}

// System-specific test helpers

func createSystemAPIManager(appReplicas, sidekiqReplicas *int64) *appsv1alpha1.APIManager {
	var (
		name                  = "example-apimanager"
		namespace             = "operator-unittest"
		wildcardDomain        = "test.3scale.net"
		appLabel              = "someLabel"
		tenantName            = "someTenant"
		trueValue             = true
		tmpApicastRegistryURL = apicastRegistryURL
	)

	return &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			Apicast: &appsv1alpha1.ApicastSpec{RegistryURL: &tmpApicastRegistryURL},
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				AppLabel:                    &appLabel,
				WildcardDomain:              wildcardDomain,
				TenantName:                  &tenantName,
				ResourceRequirementsEnabled: &trueValue,
			},
			System: &appsv1alpha1.SystemSpec{
				AppSpec:         &appsv1alpha1.SystemAppSpec{Replicas: appReplicas},
				SidekiqSpec:     &appsv1alpha1.SystemSidekiqSpec{Replicas: sidekiqReplicas},
				FileStorageSpec: &appsv1alpha1.SystemFileStorageSpec{},
				SearchdSpec:     &appsv1alpha1.SystemSearchdSpec{},
			},
			PodDisruptionBudget: &appsv1alpha1.PodDisruptionBudgetSpec{Enabled: true},
		},
	}
}

// createSystemRedisSecret creates a fake system-redis secret
func createSystemRedisSecret(namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-redis",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"redis-password": []byte("fake-password"),
		},
	}
}

// createSystemDBSecret creates a fake system-database secret
func createSystemDBSecret(namespace string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "system-database",
			Namespace: namespace,
		},
		Data: map[string][]byte{},
	}
}

// createCompletedPreHookJob creates a completed PreHook job fixture for testing
func createCompletedPreHookJob(namespace string) *batchv1.Job {
	return createCompletedJob("system-app-pre", namespace, SystemImageURL(), 1)
}

// setupTestReconciler creates a SystemReconciler with the provided objects for testing
func setupTestReconciler(t *testing.T, ctx context.Context, apimanager *appsv1alpha1.APIManager, objs []runtime.Object) (*SystemReconciler, k8sclient.Client) {
	s := setupScheme(t)

	cl := fake.NewClientBuilder().
		WithScheme(s).
		WithRuntimeObjects(objs...).
		Build()
	clientAPIReader := fake.NewClientBuilder().
		WithScheme(s).
		WithRuntimeObjects(objs...).
		Build()
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)
	log := logf.Log.WithName("operator_test")

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)
	reconciler := NewSystemReconciler(baseAPIManagerLogicReconciler)

	return reconciler, cl
}

// createSystemAppDeployment creates a deployment fixture with configurable state
func createSystemAppDeployment(namespace, image string, revision int64, replicas int32, available, progressing bool) *k8sappsv1.Deployment {
	deployment := &k8sappsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemAppDeploymentName,
			Namespace: namespace,
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": strconv.FormatInt(revision, 10),
			},
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "system"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "system"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "system-master", Image: image},
						{Name: "system-provider", Image: image},
						{Name: "system-developer", Image: image},
					},
				},
			},
		},
		Status: k8sappsv1.DeploymentStatus{
			Replicas: replicas,
		},
	}

	if available {
		deployment.Status.AvailableReplicas = replicas
		deployment.Status.Conditions = append(deployment.Status.Conditions, k8sappsv1.DeploymentCondition{
			Type:   k8sappsv1.DeploymentAvailable,
			Status: v1.ConditionTrue,
		})
	}

	if progressing {
		deployment.Status.UnavailableReplicas = 1
		deployment.Status.Conditions = append(deployment.Status.Conditions, k8sappsv1.DeploymentCondition{
			Type:   k8sappsv1.DeploymentProgressing,
			Status: v1.ConditionTrue,
			Reason: "ReplicaSetUpdated",
		})
	}

	return deployment
}
