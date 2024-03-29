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
	"os"
	"strings"
	"time"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	operatorConditions "github.com/operator-framework/api/pkg/operators/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/3scale/3scale-operator/controllers/subscription/csvlocator"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/3scale/3scale-operator/pkg/reconcilers"
	"github.com/3scale/3scale-operator/version"
)

// SubscriptionReconciler reconciles 3scale operator subscription object
type SubscriptionReconciler struct {
	*reconcilers.BaseReconciler
	OperatorNamespace string
}

// blank assignment to verify that APIManagerReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &SubscriptionReconciler{}

var (
	systemredisRequirement  string
	backendredisRequirement string
	mysqlRequirement        string
	postgresRequirement     string
	rhtComponentVersion     string
)

const (
	threescaleOperatorSubscription = "3scale"
)

// +kubebuilder:rbac:groups=operators.coreos.com,resources=installplans,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=operators.coreos.com,resources=clusterserviceversions,verbs=get;list
// +kubebuilder:rbac:groups=operators.coreos.com,resources=subscriptions;operatorconditions,verbs=get;list;watch;create;update;patch

// Permission to get the ConfigMap that embeds the CSV for an InstallPlan
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get

func (r *SubscriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.BaseReconciler.Logger().WithValues("subscription", req.NamespacedName)
	logger.Info("ReconcileSubscription", "Operator version", version.Version, "3scale release", product.ThreescaleRelease)

	// Double check if the reconciled subscription is a valid 3scale operator subscription
	if !r.shouldReconcileSubscription(req) {
		return ctrl.Result{}, nil
	}

	// Retrieve the subscription from the operator namespace
	subscription := &operatorsv1alpha1.Subscription{}
	err := r.Client().Get(context.TODO(), client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, subscription)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Retrieve the latestInstallPlan to read the CSV configMap
	latestInstallPlan := &operatorsv1alpha1.InstallPlan{}
	err = wait.Poll(time.Second*5, time.Minute*3, func() (done bool, err error) {
		if subscription.Status.InstallPlanRef == nil {
			logger.Info("ReconcileSubscription", "InstallPlanRef from Subscription is nil, trying again...", fmt.Errorf("subscription doesn't contain install plan reference"))
			return false, nil
		}
		err = r.Client().Get(context.TODO(), client.ObjectKey{Name: subscription.Status.InstallPlanRef.Name, Namespace: subscription.Status.InstallPlanRef.Namespace}, latestInstallPlan)
		if err != nil {
			logger.Info("ReconcileSubscription", "Failed to get InstallPlan, trying again...", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		logger.Info("ReconcileSubscription", "Failed to get the latest InstallPlan", err)
		return ctrl.Result{}, err
	}

	csvLocator := csvlocator.NewCachedCSVLocator(csvlocator.NewConditionalCSVLocator(
		csvlocator.SwitchLocators(
			csvlocator.ForReference,
			csvlocator.ForEmbedded,
		),
	))

	// Need to prep new client to be able to fetch the CSV, this is due to if having own namespace or taret namespace installation mode,
	// the namespace where the config map with CSV exists might not be the namespace of WATCH_NAMESPACE and therefore, any resources on non-cached namespaces
	// are not accessible.
	restConfig := ctrl.GetConfigOrDie()
	restConfig.Timeout = time.Second * 10
	k8sclient, err := client.New(restConfig, client.Options{
		Scheme: r.Scheme(),
	})
	if err != nil {
		logger.Info("ReconcileSubscription", "Failed to create new rest config client", err)
		return ctrl.Result{}, err
	}

	// Fetch CSV
	csv, err := csvLocator.GetCSV(context.TODO(), k8sclient, latestInstallPlan)
	if err != nil {
		logger.Info("ReconcileSubscription", "Failed to get CSV from the latest install plan", err)
		return ctrl.Result{}, err
	}

	if val, ok := csv.ObjectMeta.Annotations[helper.RHTThreescaleBackendRedisRequirements]; ok {
		backendredisRequirement = val
	}

	if val, ok := csv.ObjectMeta.Annotations[helper.RHTThreescaleSystemRedisRequirements]; ok {
		systemredisRequirement = val
	}

	if val, ok := csv.ObjectMeta.Annotations[helper.RHTThreescaleMysqlRequirements]; ok {
		mysqlRequirement = val
	}

	if val, ok := csv.ObjectMeta.Annotations[helper.RHTThreescalePostgresRequirements]; ok {
		postgresRequirement = val
	}

	if val, ok := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec.Template.Labels["rht.comp_ver"]; ok {
		rhtComponentVersion = val
	}

	requirementsConfigMap := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      helper.OperatorRequirementsConfigMapName,
			Namespace: r.OperatorNamespace,
		},
	}

	result, err := ctrl.CreateOrUpdate(context.TODO(), r.Client(), requirementsConfigMap, func() error {
		// If not running locally, set the ownerReferences to the subscription.
		// This is because we want the requirements config map removed in case of forced upgrades
		if helper.IsRunInCluster() {
			if err := controllerutil.SetOwnerReference(subscription, requirementsConfigMap, r.Client().Scheme()); err != nil {
				return err
			}
		}
		requirementsConfigMap.Data = map[string]string{
			helper.RHTThreescaleVersion:                  rhtComponentVersion,
			helper.RHTThreescaleMysqlRequirements:        mysqlRequirement,
			helper.RHTThreescalePostgresRequirements:     postgresRequirement,
			helper.RHTThreescaleSystemRedisRequirements:  systemredisRequirement,
			helper.RHTThreescaleBackendRedisRequirements: backendredisRequirement,
		}
		return nil
	})

	if err != nil {
		logger.Info("ReconcileSubscription", "Failed to create or update requirements config map", err)
		return ctrl.Result{}, err
	}

	if result == controllerutil.OperationResultUpdated {
		return ctrl.Result{Requeue: true}, nil
	}

	conditionReconcileResult, err := r.reconcileOperatorCondtions(fmt.Sprintf("operators.coreos.com/%s.%s", subscription.Name, subscription.Namespace))
	if err != nil {
		return ctrl.Result{}, err
	}

	if conditionReconcileResult.Requeue || conditionReconcileResult.RequeueAfter > 0 {
		return conditionReconcileResult, nil
	}

	return ctrl.Result{}, nil
}

func (r *SubscriptionReconciler) shouldReconcileSubscription(request ctrl.Request) bool {
	// double check if ns is correct - note, Operator namespace is the namespace where operator runs
	if request.Namespace != r.OperatorNamespace {
		return false
	}

	// only subscriptions we want to reconcile are: 3scale-operator or 3scale-community-operator
	if strings.Contains(request.Name, threescaleOperatorSubscription) {
		return true
	}

	return false
}

func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1alpha1.Subscription{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetNamespace() == r.OperatorNamespace
		})).
		Owns(&corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      helper.OperatorRequirementsConfigMapName,
				Namespace: r.OperatorNamespace,
			},
		}).
		Complete(r)
}

func (r *SubscriptionReconciler) reconcileOperatorCondtions(coreosLabel string) (ctrl.Result, error) {
	requirementsConfigMap, err := RetrieveRequirementsConfigMap(r.Client())
	if err != nil {
		return ctrl.Result{}, err
	}

	requirementsConfigMapResourceVersion := requirementsConfigMap.ResourceVersion

	// Fetch all the relevant APIManagers
	apimList, err := r.returnAPIManagersList()
	if err != nil {
		return ctrl.Result{}, err
	}
	// No APIManagers found, requeuing.
	if len(apimList.Items) == 0 {
		return ctrl.Result{Requeue: true}, nil
	}

	// Check if all APIManagers have the requirements confirmed
	apimRequirementsConfirmed := true
	for _, apim := range apimList.Items {
		requirementsAlreadyConfirmed := apim.RequirementsConfirmed(requirementsConfigMapResourceVersion)
		if !requirementsAlreadyConfirmed {
			apimRequirementsConfirmed = false
		}
	}

	// If there's more than one operator condition, it means that the incoming operator version has been applied and approved by user or autoapproved by the operator
	operatorConditionsList, err := r.retrieveOperatorConditions(coreosLabel)
	if err != nil {
		return ctrl.Result{}, err
	}
	isOlmApprovedUpgrade := r.isOlmApprovedUpgradeScenarioDetected(operatorConditionsList)

	// If all APIManagers have confirmed the requirements, and the currently installed operator version is not the latest one coming from the CSV, it means
	// the requirements were checked for incoming upgrade, therefore an override should be created to allow for upgrade to proceed whenever it's auto-approved or approved by user
	if apimRequirementsConfirmed && isOlmApprovedUpgrade {
		// Retrieve operator condition that is blocking the upgrade
		operatorCondition, err := r.retrieveBlockingOperatorUpgradeBlockingCondition(operatorConditionsList)
		if err != nil {
			return ctrl.Result{}, err
		}

		// Apply override
		emptyListOfConditions := &operatorConditions.OperatorConditionSpec{}
		updatedConditions := append(emptyListOfConditions.Overrides, getUpgradableCondition("True", "ApprovedUpgradeScenario", "Upgrade approved, requirements confirmed, operator upgrade is ready to progress"))
		operatorCondition.Spec.Overrides = updatedConditions
		err = r.Client().Update(context.TODO(), operatorCondition)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// If all requirements are confirmed and this is not a scenario of OLM Upgrade, it means there is no known upgrade incoming, therefore, the operator un-upgradable must be set
	if apimRequirementsConfirmed && !isOlmApprovedUpgrade {
		// only 1 condition can be available at this stage
		operatorCondition := &operatorConditionsList.Items[0]
		emptyListOfConditions := &operatorConditions.OperatorConditionSpec{}
		updatedConditions := append(emptyListOfConditions.Conditions, getUpgradableCondition("False", "NoUpgradeAvailable", "No new upgrade available, blocking any automatic upgrades"))
		operatorCondition.Spec.Conditions = updatedConditions
		operatorCondition.Spec.Overrides = emptyListOfConditions.Overrides
		err = r.Client().Update(context.TODO(), operatorCondition)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// If requirements are not confirmed yet by all 3scale instances and we are checking for requirements for an incoming upgrade update the operator condition message to being rejected
	// due to unconfirmed requirements
	if !apimRequirementsConfirmed && isOlmApprovedUpgrade {
		// Retrieve operator condition that is blocking the upgrade
		operatorCondition, err := r.retrieveBlockingOperatorUpgradeBlockingCondition(operatorConditionsList)
		if err != nil {
			return ctrl.Result{}, err
		}
		emptyListOfConditions := &operatorConditions.OperatorConditionSpec{}
		updatedConditions := append(emptyListOfConditions.Conditions, getUpgradableCondition("False", "UpgradeRejected", "Requirements are not confirmed yet by all 3scale instances that are managed by the operator"))
		operatorCondition.Spec.Conditions = updatedConditions
		operatorCondition.Spec.Overrides = emptyListOfConditions.Overrides
		err = r.Client().Update(context.TODO(), operatorCondition)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}

	// If requirements are not confirmed yet by all 3scale instances and no olm upgrade has been detected, we are in scenario where it's either a fresh install (with external dbs) OR forced apply of operator on top of existing installation
	// regardless, if requirements are not confirmed, we do not want to allow any further upgrades until they are confirmed.
	if !apimRequirementsConfirmed && !isOlmApprovedUpgrade {
		// only 1 condition can be available at this stage
		operatorCondition := &operatorConditionsList.Items[0]
		emptyListOfConditions := &operatorConditions.OperatorConditionSpec{}
		updatedConditions := append(emptyListOfConditions.Conditions, getUpgradableCondition("False", "NoUpgradeAvailable", "No new upgrade available or upgrade is manual and has not been approved yet, requirements for current or incoming installation are not met"))
		operatorCondition.Spec.Conditions = updatedConditions
		operatorCondition.Spec.Overrides = emptyListOfConditions.Overrides
		err = r.Client().Update(context.TODO(), operatorCondition)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute * 1}, nil
	}

	return ctrl.Result{}, nil
}

func getUpgradableCondition(trueOrFalse, reason, message string) v1.Condition {
	condition := v1.Condition{
		Type:               operatorConditions.Upgradeable,
		Status:             v1.ConditionStatus(trueOrFalse),
		Reason:             reason,
		Message:            message,
		LastTransitionTime: v1.Now(),
	}

	return condition
}

func (r *SubscriptionReconciler) retrieveBlockingOperatorUpgradeBlockingCondition(operatorConditionsList *operatorConditions.OperatorConditionList) (*operatorConditions.OperatorCondition, error) {
	operatorConditionCR := operatorConditions.OperatorCondition{}

	for _, opCondition := range operatorConditionsList.Items {
		for _, currentCondition := range opCondition.Spec.Conditions {
			if currentCondition.Type == operatorConditions.Upgradeable {
				if currentCondition.Status == "False" {
					err := r.Client().Get(context.TODO(), client.ObjectKey{Name: opCondition.Name, Namespace: opCondition.Namespace}, &operatorConditionCR)
					if err != nil {
						return &opCondition, err
					}
				}

			}
		}
	}

	return &operatorConditionCR, nil
}

func (r *SubscriptionReconciler) isOlmApprovedUpgradeScenarioDetected(operatorConditionsList *operatorConditions.OperatorConditionList) bool {
	return len(operatorConditionsList.Items) > 1
}

func (r *SubscriptionReconciler) retrieveOperatorConditions(operatorConditionsLabel string) (*operatorConditions.OperatorConditionList, error) {
	operatorConditionsList := &operatorConditions.OperatorConditionList{}

	opts := []client.ListOption{client.HasLabels{operatorConditionsLabel}}

	err := r.Client().List(context.Background(), operatorConditionsList, opts...)
	if err != nil {
		if errors.IsNotFound(err) {
			return operatorConditionsList, nil
		}
		return operatorConditionsList, err
	}

	return operatorConditionsList, nil
}

func (r *SubscriptionReconciler) returnAPIManagersList() (*appsv1alpha1.APIManagerList, error) {
	apimanagerList := &appsv1alpha1.APIManagerList{}
	opts := []client.ListOption{}

	// Support namespace scope or cluster scoped
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return apimanagerList, fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}

	if ns != "" {
		opts = append(opts, client.InNamespace(ns))
	}

	err := r.Client().List(context.Background(), apimanagerList, opts...)
	if err != nil {

		return apimanagerList, fmt.Errorf("error listing APIManagers: %s", err)
	}

	return apimanagerList, nil
}

func RetrieveRequirementsConfigMap(k8sclient client.Client) (*corev1.ConfigMap, error) {
	requirementsConfigMap := &corev1.ConfigMap{}

	operatorNs, err := helper.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}

	err = k8sclient.Get(context.TODO(), client.ObjectKey{Name: helper.OperatorRequirementsConfigMapName, Namespace: operatorNs}, requirementsConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("error operator requirements config map does not exist yet, retrying")
		}
		return nil, err
	}

	return requirementsConfigMap, nil
}
