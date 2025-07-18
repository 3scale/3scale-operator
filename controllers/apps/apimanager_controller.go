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
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"github.com/3scale/3scale-operator/pkg/handlers"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/pkg/upgrade"
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
// +kubebuilder:rbac:groups=grafana.integreatly.org,namespace=placeholder,resources=grafanadashboards,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=config.openshift.io,resources=clusterversions,verbs=get;list;watch
// +kubebuilder:rbac:groups=autoscaling,namespace=placeholder,resources=horizontalpodautoscalers,verbs=create;delete;list;watch

func (r *APIManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.BaseReconciler.Logger().WithValues("apimanager", req.NamespacedName)
	logger.Info("ReconcileAPIManager", "Operator version", version.Version, "3scale release", version.ThreescaleVersionMajorMinor())

	// your logic here
	var delayedRequeue bool

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
	result, preflightsRequired, err := r.instanceRequiresPreflights(instance, logger)
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

	var preflightChecksError error
	if preflightsRequired {
		result, err, preflightChecksError = r.PreflightChecks(instance, logger)
		if err != nil {
			return ctrl.Result{}, err
		}
		if preflightChecksError != nil {
			if result.RequeueAfter > 0 {
				logger.Info("preflights failed, requeueing")
				_, statusErr := r.reconcileAPIManagerStatus(instance, preflightChecksError)
				if statusErr != nil {
					return ctrl.Result{}, statusErr
				}
				return result, nil
			}
			if result.Requeue {
				logger.Info("preflights for incoming upgrade failed, requeueing")
				delayedRequeue = true
				statusResult, statusErr := r.reconcileAPIManagerStatus(instance, preflightChecksError)
				if statusErr != nil {
					return ctrl.Result{}, statusErr
				}
				if statusResult.Requeue {
					logger.Info("Reconciling not finished. Requeueing.")
					return statusResult, nil
				}
			}
		}

		if result.Requeue && preflightChecksError == nil {
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

	if delayedRequeue {
		return ctrl.Result{RequeueAfter: time.Minute * 10}, nil
	}

	return ctrl.Result{}, nil
}

func (r *APIManagerReconciler) PreflightChecks(apimInstance *appsv1alpha1.APIManager, logger logr.Logger) (ctrl.Result, error, error) {
	logger.Info("Running requirements check for preflights...")

	systemRedisVerified := false
	backendRedisVerified := false
	systemDatabaseVerified := false
	var apimVersion string

	reqConfigMap, err := subController.RetrieveRequirementsConfigMap(r.Client())
	if err != nil {
		return ctrl.Result{}, err, nil
	}

	// Fail preflights early if system Redis, backend Redis or system database are not external. Since 2.16, there is no internal databases support.
	externalDatabasesCheckError := r.externalDatabasesPreflightsChecks(apimInstance, logger)
	if externalDatabasesCheckError != nil {
		return ctrl.Result{RequeueAfter: time.Minute * 10}, nil, externalDatabasesCheckError
	}

	// Just in case the requirements are ever disabled via CSV
	incomingVersion, systemRedisRequirement, backendRedisRequirement, mysqlDatabaseRequirement, postgresDatabaseRequirements := retrieveRequiredVersion(*reqConfigMap)
	if systemRedisRequirement == "" {
		systemRedisVerified = true
	}
	if backendRedisRequirement == "" {
		backendRedisVerified = true
	}
	if mysqlDatabaseRequirement == "" && postgresDatabaseRequirements == "" {
		systemDatabaseVerified = true
	}

	apimVersion = apimInstance.RetrieveRHTVersion()

	logger.Info("Starting preflight checks...")
	if !systemRedisVerified {
		systemRedisVerified, err = helper.VerifySystemRedis(r.Client(), reqConfigMap, systemRedisRequirement, apimInstance, logger)
		if err != nil {
			return ctrl.Result{RequeueAfter: time.Minute * 10}, nil, fmt.Errorf("failed to verify system redis version. Ensure that the system-redis secret is correctly configured. Error: %s", err)
		}
	}
	if !backendRedisVerified {
		backendRedisVerified, err = helper.VerifyBackendRedis(r.Client(), reqConfigMap, backendRedisRequirement, apimInstance, logger)
		if err != nil {
			return ctrl.Result{RequeueAfter: time.Minute * 10}, nil, fmt.Errorf("failed to verify backend redis version. Ensure that the backend-redis secret is correctly configured. Error: %s", err)
		}
	}
	if !systemDatabaseVerified {
		systemDatabaseVerified, err = helper.VerifySystemDatabase(r.Client(), reqConfigMap, apimInstance, logger)
		if err != nil {
			return ctrl.Result{RequeueAfter: time.Minute * 10}, nil, fmt.Errorf("failed to verify system database version. Ensure that the system-database secret is correctly configured. Error: %s", err)
		}
	}

	culprit := ""
	if !systemDatabaseVerified || !backendRedisVerified || !systemRedisVerified {
		culprit = retrieveCulprit(systemDatabaseVerified, backendRedisVerified, systemRedisVerified, postgresDatabaseRequirements, mysqlDatabaseRequirement, backendRedisRequirement, systemRedisRequirement)
	}

	// Fresh install scenario
	if (!systemDatabaseVerified || !backendRedisVerified || !systemRedisVerified) && (incomingVersion == version.ThreescaleVersionMajorMinor()) && (apimVersion == "") {
		return ctrl.Result{RequeueAfter: time.Minute * 10}, nil, fmt.Errorf("fresh installation detected but the requirements have not been met, %s", culprit)
	}

	// upgrade + any manual (remove + re-install operator) operator installs scenarios
	if (!systemDatabaseVerified || !backendRedisVerified || !systemRedisVerified) && (incomingVersion == version.ThreescaleVersionMajorMinor()) && (apimVersion != version.ThreescaleVersionMajorMinorPatch()) {
		return ctrl.Result{RequeueAfter: time.Minute * 10}, nil, fmt.Errorf("upgrade to %s have been performed but the requirements are not met, %s", version.ThreescaleVersionMajorMinor(), culprit)
	}

	// Incoming upgrade scenarios
	if (!systemDatabaseVerified || !backendRedisVerified || !systemRedisVerified) && (incomingVersion != version.ThreescaleVersionMajorMinor()) && (apimVersion == version.ThreescaleVersionMajorMinorPatch()) {
		return ctrl.Result{Requeue: true}, nil, fmt.Errorf("attempted upgrade to %s have been performed but the requirements are not met, operator will keep reconciling but ensure requirements are met in order to proceed with upgrade, %s", incomingVersion, culprit)
	}

	// At this point, all requirements are confirmed
	err = r.setRequirementsAnnotation(apimInstance, reqConfigMap.GetResourceVersion())
	if err != nil {
		return ctrl.Result{}, err, nil
	}
	return ctrl.Result{Requeue: true}, nil, nil
}

func retrieveCulprit(systemDatabaseVerified, backendRedisVerified, systemRedisVerified bool, postgresReq, mysqlReq, backendReq, systemReq string) string {
	message := ""
	if !systemDatabaseVerified {
		message = message + fmt.Sprintf("system database version mismatch - required is Postgres %s, MySQL %s; ", postgresReq, mysqlReq)
	}
	if !backendRedisVerified {
		message = message + fmt.Sprintf("backend redis version mismatch - required is Redis %s; ", backendReq)
	}
	if !systemRedisVerified {
		message = message + fmt.Sprintf("system redis version mismatch - required is Redis %s; ", systemReq)
	}

	return message
}

func retrieveRequiredVersion(reqConfigMap v1.ConfigMap) (string, string, string, string, string) {
	// if not in fresh install - rht_comp_version must always be present, if it's not, there is something wrong with the CSV and we should not continue.
	return reqConfigMap.Data[helper.RHTThreescaleVersion], reqConfigMap.Data[helper.RHTThreescaleSystemRedisRequirements], reqConfigMap.Data[helper.RHTThreescaleBackendRedisRequirements],
		reqConfigMap.Data[helper.RHTThreescaleMysqlRequirements], reqConfigMap.Data[helper.RHTThreescalePostgresRequirements]
}

func (r *APIManagerReconciler) externalDatabasesPreflightsChecks(apimInstance *appsv1alpha1.APIManager, logger logr.Logger) error {
	backendRedisIsInternal, systemRedisIsInternal, systemDatabaseIsInternal := helper.InternalDatabases(*apimInstance, logger)
	// If all are already verified, exit earlier

	if backendRedisIsInternal || systemDatabaseIsInternal || systemRedisIsInternal {
		return fmt.Errorf("cannot continue with the installation. External databases are required to be set in the APIManager for system Redis, backend Redis and system database")
	}

	return nil
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
		return err
	}

	resourceVersionChangePredicate := predicate.ResourceVersionChangedPredicate{}

	redisConfigLabelSelector := &apimachinerymetav1.LabelSelector{
		MatchLabels: map[string]string{
			"threescale_component_element": "redis",
		},
	}
	redisConfigLabelPredicate, err := predicate.LabelSelectorPredicate(*redisConfigLabelSelector)
	if err != nil {
		return err
	}

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
		Owns(&v1.ConfigMap{}, builder.WithPredicates(redisConfigLabelPredicate)).
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

	dependencyReconciler := r.dependencyReconcilerForComponents(baseAPIManagerLogicReconciler)
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

	zyncReconciler := operator.NewZyncReconciler(baseAPIManagerLogicReconciler, cr.IsZyncEnabled())
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

	// Create the hashed secret to track watched secrets' changes
	err = r.reconcileHashedSecret(cr)
	if err != nil {
		return reconcile.Result{}, err
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
		return ctrl.Result{}, fmt.Errorf("failed to update APIManager status: %w", err)
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

func (r *APIManagerReconciler) dependencyReconcilerForComponents(baseAPIManagerLogicReconciler *operator.BaseAPIManagerLogicReconciler) operator.DependencyReconciler {
	// Helper type that contains the constructors for a dependency reconciler
	type constructors struct {
		External operator.DependencyReconcilerConstructor
	}

	// Define constructors for each reconciler
	systemDatabaseConstructors := constructors{
		External: operator.NewSystemExternalDatabaseReconciler,
	}
	systemRedisConstructors := constructors{
		External: operator.NewSystemExternalRedisReconciler,
	}
	backendRedisConstructors := constructors{
		External: operator.NewBackendExternalRedisReconciler,
	}

	// Build the final reconciler composed by the chosen external/internal combination
	result := []operator.DependencyReconciler{
		systemDatabaseConstructors.External(baseAPIManagerLogicReconciler),
		systemRedisConstructors.External(baseAPIManagerLogicReconciler),
		backendRedisConstructors.External(baseAPIManagerLogicReconciler),
	}

	return &operator.CompositeDependencyReconciler{
		Reconcilers: result,
	}
}

func (r *APIManagerReconciler) instanceRequiresPreflights(cr *appsv1alpha1.APIManager, logger logr.Logger) (ctrl.Result, bool, error) {
	if helper.IsPreflightBypassed() {
		return ctrl.Result{}, false, nil
	}

	requirementsConfigMap, err := subController.RetrieveRequirementsConfigMap(r.Client())
	if err != nil {
		// If configMap isn't ready yet, requeue
		return ctrl.Result{Requeue: true}, false, fmt.Errorf("requirements config map not found yet")
	}

	isMultiHopDetected, err := cr.IsMultiMinorHopDetected()
	if err != nil {
		// If establishing if it's multihop failed, requeue
		return ctrl.Result{Requeue: true}, false, err
	}

	if isMultiHopDetected {
		// if it is multihop, do not requeue but process the error update on APIM Status
		return ctrl.Result{}, false, fmt.Errorf("attempted upgrade from %s to %s not allowed", cr.RetrieveRHTVersion(), requirementsConfigMap.Data[helper.RHTThreescaleVersion])
	}

	// Even if requirements are already confirmed, we need to run preflights again for all new preflights added in version N+1 (N=the version new preflights were introduced - 1):
	// List of added preflights to 2.16:
	// - external DB check
	externalDatabasesCheckError := r.externalDatabasesPreflightsChecks(cr, logger)
	if externalDatabasesCheckError != nil {
		return ctrl.Result{}, true, externalDatabasesCheckError
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

func (r *APIManagerReconciler) reconcileHashedSecret(cr *appsv1alpha1.APIManager) error {
	secretLabels := map[string]string{
		"app": *cr.Spec.AppLabel,
	}
	secret, err := component.HashedSecret(r.Context(), r.Client(), cr.Get3scaleSecretRefs(), cr.Namespace, secretLabels)
	if err != nil {
		r.Logger().Error(err, "failed to generate hashed-secret-data secret")
		return err
	}

	secretMutators := []reconcilers.SecretMutateFn{
		reconcilers.SecretStringDataMutator,
	}
	err = r.ReconcileResource(&v1.Secret{}, secret, reconcilers.DeploymentSecretMutator(secretMutators...))
	if err != nil {
		r.Logger().Error(err, "failed to reconcile hashed-secret-data secret")
		return err
	}

	return nil
}
