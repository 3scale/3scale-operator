/*
Copyright 2020 Red Hat.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/3scale/3scale-operator/pkg/upgrade"

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimachinerymetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	subController "github.com/3scale/3scale-operator/controllers/subscription"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/handlers"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
)

// APIManagerReconciler reconciles a APIManager object
type APIManagerReconciler struct {
	*reconcilers.BaseReconciler
	SecretLabelSelector apimachinerymetav1.LabelSelector
	WatchedNamespace    string
}

// blank assignment to verify that APIManagerReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &APIManagerReconciler{}

// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.3scale.net,namespace=placeholder,resources=apimanagers/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace=placeholder,resources=pods;services;services/finalizers;replicationcontrollers;endpoints;persistentvolumeclaims;events;configmaps;secrets;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace=placeholder,resources=deployments;daemonsets;replicasets;statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace=placeholder,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,namespace=placeholder,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.openshift.io,namespace=placeholder,resources=imagestreams;imagestreams/layers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=image.openshift.io,namespace=placeholder,resources=imagestreamtags,verbs=get;list;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,namespace=placeholder,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,namespace=placeholder,resources=routes/custom-host,verbs=create
// +kubebuilder:rbac:groups=route.openshift.io,namespace=placeholder,resources=routes/status,verbs=get
// +kubebuilder:rbac:groups=apps.openshift.io,namespace=placeholder,resources=deploymentconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,namespace=placeholder,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,namespace=placeholder,resources=podmonitors;servicemonitors;prometheusrules,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=integreatly.org,namespace=placeholder,resources=grafanadashboards,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch
// +kubebuilder:rbac:groups=autoscaling,namespace=placeholder,resources=horizontalpodautoscalers,verbs=create;delete;list;watch

func (r *APIManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.BaseReconciler.Logger().WithValues("apimanager", req.NamespacedName)
	logger.Info("ReconcileAPIManager", "Operator version", version.Version, "3scale release", product.ThreescaleRelease)

	// your logic here

	instance, err := r.apiManagerInstance(req.NamespacedName)
	if err != nil {
		logger.Error(err, "Error fetching apimanager instance")
		return ctrl.Result{}, err
	}
	if instance == nil {
		logger.Info("resource not found. Ignoring since object must have been deleted")
		return ctrl.Result{}, nil
	}

	err = r.validateCR(instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Establish whether or not the preflights checks should be run
	result, preflightsRequired, err := r.instanceRequiresPreflights(instance)
	if err != nil {
		if result.Requeue {
			logger.Info("failed to establish whether the preflights should be run or not")
			return ctrl.Result{}, err
		}

		// If we do not receive signal to requeue, it means we are in multi minor hop and should stop reconciliation after updating APIM Status.
		statusResult, statusErr := r.reconcileAPIManagerStatus(instance, err)
		if statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		if statusResult.Requeue {
			logger.Info("Reconciling not finished. Requeueing.")
			return statusResult, nil
		}

		return ctrl.Result{}, nil
	}

	// Perform preflight checks if the preflights checks are required

	delayedRequeue := false
	var preflightChecksError error
	if preflightsRequired {
		// Perform preflight checks
		result, preflightChecksError = r.PreflightChecks(instance, logger)

		// If an error during preflights is encountered
		if preflightChecksError != nil {
			/*  If there is Requeue (either immediate requeue or requeue after) - we are in failed hard state, do not progress with APIM reconciliation, instead, requeue immediately or after certain time period
			    This can happen if:
				- there are any generic errors, for example, throwaway pod creation or os.exec call
				- an Error with Requeue after, means that we are in error state because of the version mismatch, we do want to give some time for the user to bump the databases etc before requeuing instead of spamming the controller
			*/
			if result.Requeue || result.RequeueAfter > 0 {
				logger.Info("encountered error during preflights checks.")
				statusResult, statusErr := r.reconcileAPIManagerStatus(instance, preflightChecksError)
				if statusErr != nil {
					return ctrl.Result{}, statusErr
				}
				// Only if there is an update error due to resource version mismatch
				if statusResult.Requeue {
					logger.Info("Reconciling not finished. Requeueing.")
					return statusResult, nil
				}
				return result, nil
			}
			/*
				If there is an error, but we are not getting a signal to immediately requeue, it means we have encountered a database version mismatch, but we are in scenario where we want to continue with
				the reconciliation since it's not a hard block that could affect existing 3scale instance. This can for example happen when:
				2.15 is being used, 2.16 requires a newer version of postgres, but 3scale instance is connected to older version of postgres, in this scenario currently installed operator version and the
				3scale instance should remain unaffected, meaning, operator should still reconcile the instance as it would normally but it also should periodically check for the database version to eventually
				pass the preflight giving the upgrade a green light.
			*/
			if !result.Requeue {
				delayedRequeue = true
				statusResult, statusErr := r.reconcileAPIManagerStatus(instance, preflightChecksError)
				if statusErr != nil {
					return ctrl.Result{}, statusErr
				}
				// Only if there is an update error due to resource version mismatch
				if statusResult.Requeue {
					logger.Info("Reconciling not finished. Requeueing.")
					return statusResult, nil
				}
			}
		}
		// If the requeue signal was sent with no error, skip the requeue to avoid controller run since the only situation in when the Requeue is sent back with no error is when there was an update done
		// to the APIM annotation to confirm that the preflights have passed. This call on it's own will trigger a reconcile.
		if result.Requeue {
			return ctrl.Result{}, nil
		}
	}

	res, err := r.setAPIManagerDefaults(instance)
	if err != nil {
		logger.Error(err, "Error")
		return ctrl.Result{}, err
	}
	if res.Requeue {
		logger.Info("Defaults set for APIManager resource")
		return res, nil
	}

	specResult, specErr := r.reconcileAPIManagerLogic(instance)
	if specErr != nil && specResult.Requeue {
		logger.Info("Reconciling not finished. Requeueing.")
		return specResult, nil
	}

	// reconcile status regardless specErr
	statusResult, statusErr := r.reconcileAPIManagerStatus(instance, preflightChecksError)
	if statusErr != nil {
		return ctrl.Result{}, statusErr
	}

	if specErr != nil {
		return ctrl.Result{}, specErr
	}

	if statusResult.Requeue {
		logger.Info("Reconciling not finished. Requeueing.")
		return statusResult, nil
	}

	// This isn't ideal as potentially, if there's a mismatch of version in incoming version of 3scale operator, and user does multiple changes to APIManager, we might get multiple delayedRequests queued up.
	// However, on the other side, this only happens when user switches channel to next minor version, which means they do attempt to upgrade so the chances that they flipped channel without the intention of
	// upgrading are slim so this queue won't grow too big.
	if delayedRequeue {
		return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
	}

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) PreflightChecks(apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (ctrl.Result, error) {
	systemRedisVerified := false
	backendRedisVerified := false
	systemDatabaseVerified := false
	var apimVersion string

	logger.Info("Starting preflight checks...")
	reqConfigMap, err := subController.RetrieveRequirementsConfigMap(r.Client())
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	// If it's fresh install with internal components - no preflights are required, pass requirements directly from here.
	// TODO: Improvement - not having dbs set as external doesn't necessarily mean they are internal.
	if apimInstance.IsInFreshInstallationScenario() {
		logger.Info("Running in fresh install scenario")
		backendRedisVerified, systemRedisVerified, systemDatabaseVerified = helper.InternalDatabases(*apimInstance, logger)

		// If all are already verified, exit earlier
		if backendRedisVerified && systemRedisVerified && systemDatabaseVerified {
			err := r.setRequirementsAnnotation(apimInstance, reqConfigMap.GetResourceVersion())
			if err != nil {
				return ctrl.Result{Requeue: true}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	} else {
		apimVersion = apimInstance.RetrieveRHTVersion()
	}

	// if not in fresh install - rht_comp_version must always be present, if it's not, there is something wrong with the CSV and we should not continue.
	incomingVersion := reqConfigMap.Data[helper.RHTThreescaleVersion]
	if incomingVersion == "" {
		return ctrl.Result{Requeue: true}, fmt.Errorf("unable to find %s on requirements config map", helper.RHTThreescaleVersion)
	}

	// If system redis requirements are not found, pass.
	systemRedisRequirement := reqConfigMap.Data[helper.RHTThreescaleSystemRedisRequirements]
	if systemRedisRequirement == "" {
		logger.Info("No requirements for System Redis found")
		systemRedisVerified = true
	}

	// If backend redis requirements are not found, pass.
	backendRedisRequirement := reqConfigMap.Data[helper.RHTThreescaleBackendRedisRequirements]
	if backendRedisRequirement == "" {
		logger.Info("No requirements for Backend Redis found")
		backendRedisVerified = true
	}

	// If system db requirements are not found, pass.
	mysqlDatabaseRequirement := reqConfigMap.Data[helper.RHTThreescaleMysqlRequirements]
	postgresDatabaseRequirement := reqConfigMap.Data[helper.RHTThreescalePostgresRequirements]
	if mysqlDatabaseRequirement == "" && postgresDatabaseRequirement == "" {
		logger.Info("No requirements for System Database found")
		systemDatabaseVerified = true
	}

	// If system redis is not verified by now, verify.
	if !systemRedisVerified {
		systemRedisVerified, err = helper.VerifySystemRedis(r.Client(), reqConfigMap, apimInstance, logger)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// If backend redis is not verified by now, verify.
	if !backendRedisVerified {
		backendRedisVerified, err = helper.VerifyBackendRedis(r.Client(), reqConfigMap, apimInstance, logger)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// If system database is not verified by now, verify.
	if !systemDatabaseVerified {
		systemDatabaseVerified, err = helper.VerifySystemDatabase(r.Client(), reqConfigMap, apimInstance, logger)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}

	// At this stage, we either want to triggered a Delayed Requeue to give time to handle the components version or, update the APIManager to confirm that requirements are met.

	// Following cases must be covered
	// Fresh Installation case with external components.
	/*
		threescaleProductVersion = 2.15
		incomingVersion = 2.15
		APIMVersion = ""

		Scenario:
		I have no 3scale instance, and now, I've installed 2.15 with External databases

		Expected outcome:
		Requirements must be confirmed. Do not continue if they are not.
	*/
	if (!systemDatabaseVerified || !backendRedisVerified || !systemRedisVerified) && (incomingVersion == product.ThreescaleRelease) && (apimVersion == "") {
		return ctrl.Result{RequeueAfter: time.Minute * 10}, fmt.Errorf("fresh installation detected but the requirements have not been met")
	}

	/*
		threescaleProductVersion = 2.15
		incomingVersion = 2.15
		APIMVersion = 2.14

		Scenario:
		I have 2.14 installed, I've approved 2.15 upgrade

		Expected outcome:
		Do not let 2.15 installation if pre-flights are not confirmed
	*/
	if (!systemDatabaseVerified || !backendRedisVerified || !systemRedisVerified) && (incomingVersion == product.ThreescaleRelease) && (apimVersion != product.ThreescaleRelease) {
		return ctrl.Result{RequeueAfter: time.Minute * 10}, fmt.Errorf("upgrade to %s have been performed but the requirements are not met", product.ThreescaleRelease)
	}

	/*
		threescaleProductVersion = 2.14
		incomingVersion = 2.15
		APIMVersion = 2.14

		Scenario:
		I have 2.14 installed, 2.15 upgrade is discovered

		Expected outcome:
		Do not let 2.15 progress, but do reconcile existing 2.14
	*/
	if (!systemDatabaseVerified || !backendRedisVerified || !systemRedisVerified) && (incomingVersion != product.ThreescaleRelease) && (apimVersion == product.ThreescaleRelease) {
		return ctrl.Result{}, fmt.Errorf("attempted upgrade to %s have been performed but the requirements are not met, operator will keep reconciling but ensure requirements are met in order to proceed with upgrade", incomingVersion)
	}

	// At this point, all requirements are confirmed
	err = r.setRequirementsAnnotation(apimInstance, reqConfigMap.GetResourceVersion())
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *APIManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	secretToApimanagerEventMapper := &SecretToApimanagerEventMapper{
		Context:   r.Context(),
		K8sClient: r.Client(),
		Logger:    r.Logger().WithName("secretToApimanagerEventMapper"),
		Namespace: r.WatchedNamespace,
	}

	configMapToApimanagerEventMapper := &ConfigMapToApimanagerEventMapper{
		Context:   r.Context(),
		K8sClient: r.Client(),
		Logger:    r.Logger().WithName("configMapToApimanagerEventMapper"),
		Namespace: r.WatchedNamespace,
	}

	handlers := &handlers.APIManagerRoutesEventMapper{
		Context:   r.Context(),
		K8sClient: r.Client(),
		Logger:    r.Logger().WithName("APIManagerRoutesHandler"),
	}

	operatorNamespace, err := helper.GetOperatorNamespace()
	if err != nil {
		return err
	}

	labelSelectorPredicate, err := predicate.LabelSelectorPredicate(r.SecretLabelSelector)
	if err != nil {
		return nil
	}

	resourceVersionChangePredicate := predicate.ResourceVersionChangedPredicate{}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.APIManager{}).
		Watches(
			&v1.Secret{},
			handler.EnqueueRequestsFromMapFunc(secretToApimanagerEventMapper.Map),
			builder.WithPredicates(labelSelectorPredicate),
		).
		Owns(&k8sappsv1.Deployment{}).
		Watches(&routev1.Route{}, handler.EnqueueRequestsFromMapFunc(handlers.Map)).
		Watches(
			&v1.ConfigMap{
				ObjectMeta: apimachinerymetav1.ObjectMeta{
					Name:      helper.OperatorRequirementsConfigMapName,
					Namespace: operatorNamespace,
				},
			},
			handler.EnqueueRequestsFromMapFunc(configMapToApimanagerEventMapper.Map),
			builder.WithPredicates(resourceVersionChangePredicate),
		).
		Complete(r)
}

func (r *APIManagerReconciler) validateCR(cr *appsv1alpha1.APIManager) error {
	fieldError := field.ErrorList{}
	// internal validation
	fieldError = append(fieldError, cr.Validate()...)

	fieldError = append(fieldError, r.validateApicastTLSCertificates(cr)...)

	if len(fieldError) > 0 {
		return fieldError.ToAggregate()
	}

	return nil
}

func (r *APIManagerReconciler) apiManagerInstance(namespacedName types.NamespacedName) (*appsv1alpha1.APIManager, error) {
	// Fetch the APIManager instance
	instance := &appsv1alpha1.APIManager{}

	err := r.Client().Get(context.TODO(), namespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil, nil
		}
		return nil, err
	}
	return instance, nil
}

func (r *APIManagerReconciler) setAPIManagerDefaults(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	updated := cr.UpdateExternalComponentsFromHighAvailability()

	defaultsUpdated, err := cr.SetDefaults()
	if err != nil {
		return ctrl.Result{}, err
	}
	updated = updated || defaultsUpdated

	if updated {
		err = r.Client().Update(context.TODO(), cr)
	}

	return ctrl.Result{Requeue: updated}, err
}

func (r *APIManagerReconciler) reconcileAPIManagerLogic(cr *appsv1alpha1.APIManager) (reconcile.Result, error) {
	baseAPIManagerLogicReconciler := operator.NewBaseAPIManagerLogicReconciler(r.BaseReconciler, cr)
	imageReconciler := operator.NewAMPImagesReconciler(baseAPIManagerLogicReconciler)
	result, err := imageReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	dependencyReconciler := r.dependencyReconcilerForComponents(cr, baseAPIManagerLogicReconciler)
	result, err = dependencyReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	backendReconciler := operator.NewBackendReconciler(baseAPIManagerLogicReconciler)
	result, err = backendReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	memcachedReconciler := operator.NewMemcachedReconciler(baseAPIManagerLogicReconciler)
	result, err = memcachedReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	systemSearchdReconciler := operator.NewSystemSearchdReconciler(baseAPIManagerLogicReconciler)
	result, err = systemSearchdReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	systemReconciler := operator.NewSystemReconciler(baseAPIManagerLogicReconciler)
	result, err = systemReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	zyncReconciler := operator.NewZyncReconciler(baseAPIManagerLogicReconciler)
	result, err = zyncReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	apicastReconciler := operator.NewApicastReconciler(baseAPIManagerLogicReconciler)
	result, err = apicastReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	genericMonitoringReconciler := operator.NewGenericMonitoringReconciler(baseAPIManagerLogicReconciler)
	result, err = genericMonitoringReconciler.Reconcile()
	if err != nil || result.Requeue {
		return result, err
	}

	// 3scale 2.14 -> 2.15
	err = upgrade.DeleteImageStreams(cr.Namespace, r.Client())
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) reconcileAPIManagerStatus(cr *appsv1alpha1.APIManager, preflightsError error) (reconcile.Result, error) {
	statusReconciler := NewAPIManagerStatusReconciler(r.BaseReconciler, cr, preflightsError)
	res, err := statusReconciler.Reconcile()
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("Failed to update APIManager status: %w", err)
	}

	return res, nil
}

func (r *APIManagerReconciler) validateApicastTLSCertificates(cr *appsv1alpha1.APIManager) field.ErrorList {
	fieldErrors := field.ErrorList{}

	if cr.Spec.Apicast != nil && cr.Spec.Apicast.ProductionSpec != nil && cr.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef != nil {
		secretPath := field.NewPath("spec").Child("apicast").Child("productionSpec").Child("httpsCertificateSecretRef")
		if cr.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef.Name == "" {
			fieldErrors = append(fieldErrors, field.Required(secretPath.Child("name"), "secret name not provided"))
		} else {
			nn := types.NamespacedName{
				Name:      cr.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef.Name,
				Namespace: cr.Namespace,
			}
			err := helper.ValidateTLSSecret(nn, r.Client())
			if err != nil {
				fieldErrors = append(fieldErrors, field.Invalid(secretPath, cr.Spec.Apicast.ProductionSpec.HTTPSCertificateSecretRef, err.Error()))
			}
		}
	}

	if cr.Spec.Apicast != nil && cr.Spec.Apicast.StagingSpec != nil && cr.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef != nil {
		secretPath := field.NewPath("spec").Child("apicast").Child("stagingSpec").Child("httpsCertificateSecretRef")
		if cr.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef.Name == "" {
			fieldErrors = append(fieldErrors, field.Required(secretPath.Child("name"), "secret name not provided"))
		} else {
			nn := types.NamespacedName{
				Name:      cr.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef.Name,
				Namespace: cr.Namespace,
			}
			err := helper.ValidateTLSSecret(nn, r.Client())
			if err != nil {
				fieldErrors = append(fieldErrors, field.Invalid(secretPath, cr.Spec.Apicast.StagingSpec.HTTPSCertificateSecretRef, err.Error()))
			}
		}
	}

	return fieldErrors
}

func (r *APIManagerReconciler) dependencyReconcilerForComponents(cr *appsv1alpha1.APIManager, baseAPIManagerLogicReconciler *operator.BaseAPIManagerLogicReconciler) operator.DependencyReconciler {
	// Helper type that contains the constructors for a dependency reconciler
	// whether it's external or internal
	type constructors struct {
		External operator.DependencyReconcilerConstructor
		Internal operator.DependencyReconcilerConstructor
	}

	// Helper function that instantiates a dependency reconciler depending
	// on whether it's external or internal
	selectReconciler := func(cs constructors, selectIsExternal func(*appsv1alpha1.ExternalComponentsSpec) bool) operator.DependencyReconciler {
		constructor := cs.Internal
		if selectIsExternal(cr.Spec.ExternalComponents) {
			constructor = cs.External
		}

		return constructor(baseAPIManagerLogicReconciler)
	}

	// Select whether to use PostgreSQL or MySQL for the System database
	var systemDatabaseReconcilerConstructor operator.DependencyReconcilerConstructor
	if cr.Spec.System.DatabaseSpec != nil && cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		systemDatabaseReconcilerConstructor = operator.CompositeDependencyReconcilerConstructor(
			operator.NewSystemPostgreSQLReconciler,
		)
	} else {
		systemDatabaseReconcilerConstructor = operator.CompositeDependencyReconcilerConstructor(
			operator.NewSystemMySQLReconciler,
		)
	}

	systemDatabaseConstructors := constructors{
		External: operator.NewSystemExternalDatabaseReconciler,
		Internal: systemDatabaseReconcilerConstructor,
	}
	systemRedisConstructors := constructors{
		External: operator.NewSystemExternalRedisReconciler,
		Internal: operator.NewSystemRedisDependencyReconciler,
	}
	backendRedisConstructors := constructors{
		External: operator.NewBackendExternalRedisReconciler,
		Internal: operator.NewBackendRedisDependencyReconciler,
	}

	// Build the final reconciler composed by the chosen external/internal
	// combination
	result := []operator.DependencyReconciler{
		selectReconciler(systemDatabaseConstructors, appsv1alpha1.SystemDatabase),
		selectReconciler(systemRedisConstructors, appsv1alpha1.SystemRedis),
		selectReconciler(backendRedisConstructors, appsv1alpha1.BackendRedis),
	}

	return &operator.CompositeDependencyReconciler{
		Reconcilers: result,
	}
}

func (r *APIManagerReconciler) IsIncomingVersionDifferent(incomingVersion string) bool {
	return incomingVersion != product.ThreescaleRelease
}

func (r *APIManagerReconciler) instanceRequiresPreflights(cr *appsv1alpha1.APIManager) (ctrl.Result, bool, error) {
	requirementsConfigMap := &v1.ConfigMap{}

	if helper.IsPreflightBypassed() {
		return ctrl.Result{}, false, nil
	}

	isMultiHopDetected, err := cr.IsMultiMinorHopDetected()
	if err != nil {
		// If establishing if it's multihop failed, requeue
		return ctrl.Result{Requeue: true}, false, err
	}
	if isMultiHopDetected {
		// if it is multihop, do not requeue but process the error update on APIM Status
		return ctrl.Result{}, false, fmt.Errorf("multi minor hop detected")
	}

	requirementsConfigMap, err = subController.RetrieveRequirementsConfigMap(r.Client())
	if err != nil {
		// If configMap isn't ready yet, requeue
		return ctrl.Result{Requeue: true}, false, err
	}

	// Check if current requirements are already confirmed
	requirementsAlreadyConfirmed := cr.RequirementsConfirmed(requirementsConfigMap.GetResourceVersion())
	if requirementsAlreadyConfirmed {
		return ctrl.Result{}, false, nil
	}

	return ctrl.Result{}, true, nil
}

func (r *APIManagerReconciler) setRequirementsAnnotation(apim *appsv1alpha1.APIManager, resourceVersion string) error {
	currentAnnotations := apim.GetAnnotations()
	if currentAnnotations == nil {
		currentAnnotations = make(map[string]string)
	}
	currentAnnotations[appsv1alpha1.ThreescaleRequirementsConfirmed] = resourceVersion
	apim.SetAnnotations(currentAnnotations)

	err := r.Client().Update(context.TODO(), apim)
	if err != nil {
		return err
	}

	return nil
}
