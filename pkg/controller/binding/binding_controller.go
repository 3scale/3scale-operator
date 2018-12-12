package binding

import (
	"context"
	apiv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Binding Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBinding{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("binding-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Binding
	err = c.Watch(&source.Kind{Type: &apiv1alpha1.Binding{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Binding
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &apiv1alpha1.Binding{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBinding{}

// ReconcileBinding reconciles a Binding object
type ReconcileBinding struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Binding object and makes changes based on the state read
// and what is in the Binding.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBinding) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling Binding %s/%s\n", request.Namespace, request.Name)

	// Fetch the Binding instance
	instance := &apiv1alpha1.Binding{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Printf("Error: %s", err)
		return reconcile.Result{}, err
	}

	consolidated := newConsolidated(instance)
	found := &apiv1alpha1.Consolidated{}

	// GET THE Consolidated OBJECT.
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: consolidated.Name, Namespace: consolidated.Namespace}, found)

	// IF Consolidated doesn't exists, let's create it.
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating new Consolidated object %s/%s\n", consolidated.Namespace, consolidated.Name)

		// GET SECRET
		secret := &corev1.Secret{}

		// TODO: fix namespace default
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.CredentialsRef.Name, Namespace: instance.Namespace}, secret)
		InternalCredential := apiv1alpha1.InternalCredential{}

		if err != nil && errors.IsNotFound(err) {

			log.Printf("secret: %s for binding: %s NOT found", instance.Spec.CredentialsRef.Name, instance.Name)
			//return reconcile.Result{}, err
		} else if err != nil {
			log.Printf("Error: %s", err)
			return reconcile.Result{}, err
		} else {

			log.Printf("secret: %s found for %s", instance.Spec.CredentialsRef.Name, instance.Name)

			InternalCredential = apiv1alpha1.InternalCredential{
				AccessToken: string(secret.Data["access_token"]),
				AdminURL:    string(secret.Data["admin_portal_url"]),
			}
		}

		// GET APIS
		apis := &apiv1alpha1.APIList{}
		opts := &client.ListOptions{}
		opts.InNamespace(instance.Namespace)
		opts.MatchingLabels(instance.Spec.APISelector.MatchLabels)

		err = r.client.List(context.TODO(), opts, apis)

		if err != nil && errors.IsNotFound(err) {

			// No API objects
			log.Printf("Binding: %s in namespace: %s doesn't match any API object", instance.Name, instance.Namespace)
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil

		} else if err != nil {
			// Something is broken
			log.Printf("Error: %s", err)
			return reconcile.Result{}, err
		}

		// Add each API info to the consolidated object
		for _, api := range apis.Items {

			internalAPI := apiv1alpha1.InternalAPI{
				Name:        api.Name,
				Description: api.Spec.Description,
				Credentials: InternalCredential,
				Integration: apiv1alpha1.Integration{},
				Metrics:     nil,
				Plans:       nil,
			}

			var opts client.ListOptions

			//Get Metrics for each API
			metrics := &apiv1alpha1.MetricList{}
			opts = client.ListOptions{}
			opts.InNamespace(api.Namespace)
			opts.MatchingLabels(api.Spec.MetricSelector.MatchLabels)
			err = r.client.List(context.TODO(), &opts, metrics)

			if err != nil && errors.IsNotFound(err) {
				// Nothing has been found
				log.Printf("No metrics found for: %s\n", api.Name)
			} else if err != nil {
				// Something is broken
				log.Printf("Error: %s", err)
				return reconcile.Result{}, err
			} else {
				// Let's do our job.
				for _, metric := range metrics.Items {

					internalMetric := apiv1alpha1.InternalMetric{
						Name:        metric.Name,
						Unit:        metric.Spec.Unit,
						Description: metric.Spec.Description,
					}

					internalAPI.Metrics = append(internalAPI.Metrics, internalMetric)
				}
			}

			//Get Plans for each API
			plans := &apiv1alpha1.PlanList{}
			opts = client.ListOptions{}
			opts.InNamespace(api.Namespace)
			opts.MatchingLabels(api.Spec.PlanSelector.MatchLabels)
			err = r.client.List(context.TODO(), &opts, plans)

			if err != nil && errors.IsNotFound(err) {
				// Nothing has been found
				log.Printf("No plans found for: %s\n", api.Name)
			} else if err != nil {
				// Something is broken
				log.Printf("Error: %s", err)
				return reconcile.Result{}, err
			} else {
				// Let's do our job.
				internalPlan := apiv1alpha1.InternalPlan{}
				for _, plan := range plans.Items {

					// Fill the internal Plan with Plan and Limits.
					internalPlan = apiv1alpha1.InternalPlan{
						Name:             plan.Name,
						TrialPeriodDays:  plan.Spec.TrialPeriod,
						ApprovalRequired: plan.Spec.AprovalRequired,
						Costs:            plan.Spec.Costs,
					}

					// Get the Limits now
					limits := &apiv1alpha1.LimitList{}
					opts = client.ListOptions{}
					opts.InNamespace(api.Namespace)
					opts.MatchingLabels(plan.Spec.LimitSelector.MatchLabels)
					err = r.client.List(context.TODO(), &opts, limits)

					if err != nil && errors.IsNotFound(err) {
						// Nothing has been found
						log.Printf("No limits found for: %s\n", plan.Name)
					} else if err != nil {
						// Something is broken
						log.Printf("Error: %s", err)
						return reconcile.Result{}, err

					} else {

						// Let's do our job.
						for _, limit := range limits.Items {

							metric := &apiv1alpha1.Metric{}
							var namespace string

							if limit.Spec.Metric.Namespace == "" {
								namespace = limit.Namespace
							} else {
								namespace = limit.Spec.Metric.Namespace
							}

							reference := types.NamespacedName{
								Namespace: namespace,
								Name:      limit.Spec.Metric.Name,
							}
							err = r.client.Get(context.TODO(), reference, metric)

							if err != nil && errors.IsNotFound(err) {
								// Nothing has been found
								log.Printf("Invalid Metric for limit: %s\n", limit.Name)

								return reconcile.Result{
									Requeue:      true,
									RequeueAfter: 5 * time.Second,
								}, nil

							} else if err != nil {
								// Something is broken
								log.Printf("Error: %s", err)
								return reconcile.Result{}, err
							} else {
								internalLimit := apiv1alpha1.InternalLimit{
									Name:        limit.Name,
									Description: limit.Spec.Description,
									Period:      limit.Spec.Period,
									MaxValue:    limit.Spec.MaxValue,
									Metric:      metric.Name,
								}

								internalPlan.Limits = append(internalPlan.Limits, internalLimit)
							}
						}
					}

					// Add the Plan to the internal API object
					internalAPI.Plans = append(internalAPI.Plans, internalPlan)
				}
			}

			// Get integration Method info.
			if api.Spec.IntegrationMethod.ApicastHosted != nil {
				// ApicastHosted
				log.Println("Integration method: ApicastHosted")

				internalApicastHosted := apiv1alpha1.InternalApicastHosted{
					PrivateBaseURL:         api.Spec.IntegrationMethod.ApicastOnPrem.PrivateBaseURL,
					APITestGetRequest:      api.Spec.IntegrationMethod.ApicastOnPrem.APITestGetRequest,
					AuthenticationSettings: api.Spec.IntegrationMethod.ApicastOnPrem.AuthenticationSettings,
					MappingRules:           nil,
					Policies:               nil,
				}

				// Get Mapping Rules
				mappingRules := &apiv1alpha1.MappingRuleList{}
				opts = client.ListOptions{}
				opts.InNamespace(api.Namespace)
				opts.MatchingLabels(api.Spec.IntegrationMethod.ApicastHosted.MappingRulesSelector.MatchLabels)
				err = r.client.List(context.TODO(), &opts, mappingRules)

				if err != nil && errors.IsNotFound(err) {
					// Nothing has been found
					log.Printf("No mappingRules found for: %s\n", api.Name)
				} else if err != nil {
					// Something is broken
					log.Printf("Error: %s", err)
					return reconcile.Result{}, err
				} else {

					for _, mappingRule := range mappingRules.Items {
						// GET metric for mapping rule.
						metric := &apiv1alpha1.Metric{}
						var namespace string

						if mappingRule.Spec.MetricRef.Namespace == "" {
							namespace = api.Namespace
						} else {
							namespace = mappingRule.Spec.MetricRef.Namespace
						}

						reference := types.NamespacedName{
							Namespace: namespace,
							Name:      mappingRule.Spec.MetricRef.Name,
						}
						err = r.client.Get(context.TODO(), reference, metric)

						// TODO: How to handle metric HITS.

						if err != nil && errors.IsNotFound(err) {
							// Nothing has been found
							log.Printf("Invalid Metric for MappingRule: %s\n", mappingRule.Name)

							return reconcile.Result{
								Requeue:      true,
								RequeueAfter: 5 * time.Second,
							}, nil

						} else if err != nil {

							// Something is broken
							log.Printf("Error: %s", err)
							return reconcile.Result{}, err

						} else {

							internalMappingRule := apiv1alpha1.InternalMappingRule{
								Name:      mappingRule.Name,
								Path:      mappingRule.Spec.Path,
								Method:    mappingRule.Spec.Method,
								Increment: mappingRule.Spec.Increment,
								Metric:    metric.Name,
							}

							internalApicastHosted.MappingRules = append(
								internalApicastHosted.MappingRules,
								internalMappingRule,
							)

						}

					}

				}
				internalAPI.Integration.ApicastHosted = &internalApicastHosted

				// Get Policies
				// api.Spec.IntegrationMethod.ApicastOnPrem.PoliciesSelector

				log.Println("Integration method: ApicastHosted")

			} else if api.Spec.IntegrationMethod.ApicastOnPrem != nil {

				internalApicastOnPrem := apiv1alpha1.InternalApicastOnPrem{
					PrivateBaseURL:          api.Spec.IntegrationMethod.ApicastOnPrem.PrivateBaseURL,
					StagingPublicBaseURL:    api.Spec.IntegrationMethod.ApicastOnPrem.StagingPublicBaseURL,
					ProductionPublicBaseURL: api.Spec.IntegrationMethod.ApicastOnPrem.PrivateBaseURL,
					APITestGetRequest:       api.Spec.IntegrationMethod.ApicastOnPrem.APITestGetRequest,
					AuthenticationSettings:  api.Spec.IntegrationMethod.ApicastOnPrem.AuthenticationSettings,
					MappingRules:            nil,
					Policies:                nil,
				}
				// Get Mapping Rules
				// api.Spec.IntegrationMethod.ApicastOnPrem.MappingRulesSelector

				mappingRules := &apiv1alpha1.MappingRuleList{}
				opts = client.ListOptions{}
				opts.InNamespace(api.Namespace)
				opts.MatchingLabels(api.Spec.IntegrationMethod.ApicastOnPrem.MappingRulesSelector.MatchLabels)
				err = r.client.List(context.TODO(), &opts, mappingRules)

				if err != nil && errors.IsNotFound(err) {
					// Nothing has been found
					log.Printf("No mappingRules found for: %s\n", api.Name)
				} else if err != nil {
					// Something is broken
					log.Printf("Error: %s", err)
					return reconcile.Result{}, err
				} else {

					for _, mappingRule := range mappingRules.Items {
						// GET metric for mapping rule.
						metric := &apiv1alpha1.Metric{}
						var namespace string

						if mappingRule.Spec.MetricRef.Namespace == "" {
							namespace = api.Namespace
						} else {
							namespace = mappingRule.Spec.MetricRef.Namespace
						}

						reference := types.NamespacedName{
							Namespace: namespace,
							Name:      mappingRule.Spec.MetricRef.Name,
						}
						err = r.client.Get(context.TODO(), reference, metric)

						// TODO: How to handle metric HITS.

						if err != nil && errors.IsNotFound(err) {
							// Nothing has been found
							log.Printf("Invalid Metric for MappingRule: %s\n", mappingRule.Name)

							return reconcile.Result{
								Requeue:      true,
								RequeueAfter: 5 * time.Second,
							}, nil

						} else if err != nil {

							// Something is broken
							log.Printf("Error: %s", err)
							return reconcile.Result{}, err

						} else {

							internalMappingRule := apiv1alpha1.InternalMappingRule{
								Name:      mappingRule.Name,
								Path:      mappingRule.Spec.Path,
								Method:    mappingRule.Spec.Method,
								Increment: mappingRule.Spec.Increment,
								Metric:    metric.Name,
							}

							internalApicastOnPrem.MappingRules = append(
								internalApicastOnPrem.MappingRules,
								internalMappingRule,
							)

						}

					}

				}
				internalAPI.Integration.ApicastOnPrem = &internalApicastOnPrem

				// Get Policies
				// api.Spec.IntegrationMethod.ApicastOnPrem.PoliciesSelector

				log.Println("Integration method: ApicastOnPrem")

			} else if api.Spec.IntegrationMethod.CodePlugin != nil {
				// Assume Code plugin.

				internalCodePlugin := apiv1alpha1.InternalCodePlugin{
					AuthenticationSettings: apiv1alpha1.CodePluginAuthenticationSettings{
						Credentials: api.Spec.IntegrationMethod.CodePlugin.AuthenticationSettings.Credentials,
					},
				}

				internalAPI.Integration.CodePlugin = &internalCodePlugin

				log.Println("Integration method: CodePlugin")
			} else {
				log.Println("invalid integration method of api: %s", api.Name)

				return reconcile.Result{
					Requeue:      true,
					RequeueAfter: 5 * time.Second,
				}, nil
			}

			consolidated.Spec.APIs = append(consolidated.Spec.APIs, internalAPI)
		}

		// Create the consolidated object.
		err = r.client.Create(context.TODO(), consolidated)

		// Check if something went wrong
		if err != nil {
			log.Printf("Error: %s", err)
			return reconcile.Result{}, err
		}

		//IF getting the consolidated object failed somehow.
	} else if err != nil {

		// Something is broken
		return reconcile.Result{}, err

		//Consolidated object exists.
	} else {

		// Object Already exists
		log.Printf("Skip reconcile: Consolidated config %s/%s already exists", found.Namespace, found.Name)
	}

	return reconcile.Result{}, nil
}

func newConsolidated(binding *apiv1alpha1.Binding) *apiv1alpha1.Consolidated {

	return &apiv1alpha1.Consolidated{
		TypeMeta: v1.TypeMeta{
			Kind: "Consolidated",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      binding.Name + "-consolidated",
			Namespace: binding.Namespace,
		},
		Spec: apiv1alpha1.ConsolidatedSpec{
			Tenants: nil,
			APIs:    nil,
		},
		Status: apiv1alpha1.ConsolidatedStatus{},
	}
}
