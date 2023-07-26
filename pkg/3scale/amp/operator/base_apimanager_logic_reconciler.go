package operator

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/3scale/3scale-operator/pkg/helper"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/go-logr/logr"
	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BaseAPIManagerLogicReconciler struct {
	*reconcilers.BaseReconciler
	apiManager           *appsv1alpha1.APIManager
	logger               logr.Logger
	crdAvailabilityCache *baseAPIManagerLogicReconcilerCRDAvailabilityCache
}

type baseAPIManagerLogicReconcilerCRDAvailabilityCache struct {
	grafanaDashboardCRDAvailable *bool
	prometheusRuleCRDAvailable   *bool
	podMonitorCRDAvailable       *bool
	serviceMonitorCRDAvailable   *bool
}

type Accounts struct {
    XMLName xml.Name `xml:"accounts"`
    Account []Account `xml:"account"`
}

type Account struct {
    AdminBaseURL string `xml:"admin_base_url"`
    BaseURL      string `xml:"base_url"`
}

func NewBaseAPIManagerLogicReconciler(b *reconcilers.BaseReconciler, apiManager *appsv1alpha1.APIManager) *BaseAPIManagerLogicReconciler {
	return &BaseAPIManagerLogicReconciler{
		BaseReconciler:       b,
		apiManager:           apiManager,
		logger:               b.Logger().WithValues("APIManager Controller", apiManager.Name),
		crdAvailabilityCache: &baseAPIManagerLogicReconcilerCRDAvailabilityCache{},
	}
}

func (r *BaseAPIManagerLogicReconciler) NamespacedNameWithAPIManagerNamespace(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: r.apiManager.GetNamespace(), Name: obj.GetName()}
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePodDisruptionBudget(desired *policyv1.PodDisruptionBudget, mutatefn reconcilers.MutateFn) error {
	if !r.apiManager.IsPDBEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&policyv1.PodDisruptionBudget{}, desired, mutatefn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileImagestream(desired *imagev1.ImageStream, mutatefn reconcilers.MutateFn) error {
	return r.ReconcileResource(&imagev1.ImageStream{}, desired, mutatefn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileDeploymentConfig(desired *appsv1.DeploymentConfig, mutatefn reconcilers.MutateFn) error {
	return r.ReconcileResource(&appsv1.DeploymentConfig{}, desired, mutatefn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileService(desired *v1.Service, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.Service{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileConfigMap(desired *v1.ConfigMap, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.ConfigMap{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileServiceAccount(desired *v1.ServiceAccount, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.ServiceAccount{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileRoute(desired *routev1.Route, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&routev1.Route{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileSecret(desired *v1.Secret, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.Secret{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePersistentVolumeClaim(desired *v1.PersistentVolumeClaim, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&v1.PersistentVolumeClaim{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileRole(desired *rbacv1.Role, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&rbacv1.Role{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileRoleBinding(desired *rbacv1.RoleBinding, mutateFn reconcilers.MutateFn) error {
	return r.ReconcileResource(&rbacv1.RoleBinding{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileGrafanaDashboard(desired *grafanav1alpha1.GrafanaDashboard, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.HasGrafanaDashboards()
	if err != nil {
		return err
	}

	if !kindExists {
		if r.apiManager.IsMonitoringEnabled() {
			errToLog := fmt.Errorf("Error creating grafana dashboard object '%s'. Install grafana-operator in your cluster to create grafana dashboard objects", desired.Name)
			r.EventRecorder().Eventf(r.apiManager, v1.EventTypeWarning, "ReconcileError", errToLog.Error())
			r.logger.Error(errToLog, "ReconcileError")
		}
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&grafanav1alpha1.GrafanaDashboard{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) findSystemSidekiqPod(apimanager *appsv1alpha1.APIManager) (string, error) {
	namespace := apimanager.GetNamespace()
	podName := ""
	podList := &v1.PodList{}

    // system-sidekiq pod needs to be up & running
    err := r.waitForSystemSidekiq(apimanager)
    if err != nil {
        return "", fmt.Errorf("failed to wait for system-sidekiq: %w", err)
    }

	listOps := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"deploymentConfig": "system-sidekiq"}),
	}

	err = r.Client().List(context.TODO(), podList, listOps...)
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	for _, pod := range podList.Items {
		if pod.Status.Phase == "Running" {
			podName = pod.ObjectMeta.Name
			break
		}
	}
	
	if podName == "" {
		return "",fmt.Errorf("no matching pod found")
	}

	return podName, nil
}

func (r *BaseAPIManagerLogicReconciler) getAccessToken() (string, error) {

    namespace := r.apiManager.GetNamespace()

	secretName := types.NamespacedName{
        Namespace: namespace,
        Name:      "system-seed",
    }

    secret := &v1.Secret{}
    err := r.Client().Get(context.TODO(), secretName, secret)
    if err != nil {
        return "", fmt.Errorf("failed to get secret: %w", err)
    }

    // Retrieve the access token from the secret data
    accessToken, ok := secret.Data["MASTER_ACCESS_TOKEN"]
    if !ok {
        return "", fmt.Errorf("access token not found in secret")
    }

    return string(accessToken), nil
}

func (r *BaseAPIManagerLogicReconciler) getMasterRoute() (string, error) {

	masterPrefix := "master"

	opts := []client.ListOption{
        client.InNamespace(r.apiManager.GetNamespace()),
    }
	foundRoute := ""
	routes := routev1.RouteList{}
    err := r.Client().List(context.TODO(), &routes, opts ...)
    if err != nil {
        return "", err
    }

	for _, route := range routes.Items {
		if strings.HasPrefix(route.Spec.Host, masterPrefix) {
			foundRoute = "https://" + route.Spec.Host
			return foundRoute, nil
		}
	}

	return "", fmt.Errorf("route not found")
}

func (r *BaseAPIManagerLogicReconciler) baseRoutesExist() (bool, error) {
	expectedRoutes := 2
    serviceNames := []string{"system-master", "backend-listener"}
    opts := []client.ListOption{
        client.InNamespace(r.apiManager.GetNamespace()),
    }

    routes := routev1.RouteList{}
    err := r.Client().List(context.TODO(), &routes, opts ...)
    if err != nil {
        return false, err
    }

    serviceCount := 0
    for _, service := range serviceNames {
        found := false
        for _, route := range routes.Items {
            if route.Spec.To.Name == service {
                found = true
                break
            }
        }
        if found {
            serviceCount++
        }
    }

    if serviceCount >= expectedRoutes {
        return true, nil
    }

    return false, fmt.Errorf("base routes not found")
}

func (r *BaseAPIManagerLogicReconciler) getAccountUrls() ([]string, []string, error){

	state := "approved"

	masterRoute, err := r.getMasterRoute()
	if err != nil {
		r.logger.Error(err, "Error getting Master Route")
		return nil, nil, err
	}

	url := masterRoute + "/admin/api/accounts.xml"
	
    accessToken, err := r.getAccessToken()
	if err != nil {
		fmt.Println("Error getting Access Token:", err)
		return nil, nil, err
	}

    // Create a new HTTP GET request
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        fmt.Println("Error Creating HTTP Request:", err)
        return nil, nil, err
    }

    // Add query parameters to the request
    q := req.URL.Query()
    q.Add("access_token", accessToken)
    q.Add("state", state)
    req.URL.RawQuery = q.Encode()

    // Send the HTTP request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error sending HTTP request:", err)
        return nil, nil, err
    }
    defer resp.Body.Close()

    // Read and parse the XML response
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error reading HTTP response:", err)
        return nil, nil, err
    }

    var accounts Accounts
	err = xml.Unmarshal([]byte(body), &accounts)
	if err != nil {
		fmt.Println("Error unmarshalling HTTP response:", err)
		return nil, nil, err
	}

    var adminBaseURLs []string
    var baseURLs []string
    for _, account := range accounts.Account {
        adminBaseURLs = append(adminBaseURLs, account.AdminBaseURL)
        baseURLs = append(baseURLs, account.BaseURL)
    }

    return adminBaseURLs, baseURLs, nil
}

func (r *BaseAPIManagerLogicReconciler) routesExist() (bool, error) {
	_, err:= r.baseRoutesExist()
	if err != nil {
		return false, nil
	}

	adminBaseURLs, baseURLs, err := r.getAccountUrls()
    if err != nil {
        return false, err
    }

	opts := []client.ListOption{
        client.InNamespace(r.apiManager.GetNamespace()),
    }

    routes := routev1.RouteList{}
    err = r.Client().List(context.TODO(), &routes, opts ...)
    if err != nil {
        return false, err
    }

    // Create a map to track the presence of adminBaseURLs and baseURLs
    urlMap := make(map[string]bool)
    for _, url := range append(adminBaseURLs, baseURLs...) {
        urlMap[url] = false
    }

    // Check if all adminBaseURLs and baseURLs are present in the route location URLs
    for _, route := range routes.Items {
		routeHost := "https://" + route.Spec.Host
        for url := range urlMap {
            if routeHost == url {
                urlMap[url] = true
            }
        }
    }

    missingURLs := []string{}
    for url, found := range urlMap {
        if !found {
            missingURLs = append(missingURLs, url)
        }
    }

    if len(missingURLs) == 0 {
		return true, nil
    }

	return false, nil
}


func (r *BaseAPIManagerLogicReconciler) executeCommandOnPod(containerName string, namespace string, podName string, command []string) (string, string, error) {
	podExecutor := helper.NewPodExecutor(r.logger) 

	stdout, stderr, err := podExecutor.ExecuteRemoteContainerCommand(namespace, podName, containerName, command)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute command on pod: %w, stderr: %s", err, stderr)
	}

	fmt.Println("Command output (stdout):", stdout)
	fmt.Println("Command output (stderr):", stderr)
    return stdout, stderr, err
}

func (r *BaseAPIManagerLogicReconciler) waitForSystemSidekiq(apimanager *appsv1alpha1.APIManager) error {

    // Wait until system-sidekiq deployments are ready
    for !helper.ArrayContains(apimanager.Status.Deployments.Ready, "system-sidekiq") {
        r.Logger().Info("system-sidekiq deployments not ready. Waiting", "APIManager", apimanager.Name)
        time.Sleep(5 * time.Second)

        // Refresh APIManager status
        err := r.GetResource(types.NamespacedName{Name: r.apiManager.Name, Namespace: r.apiManager.Namespace}, apimanager)
        if err != nil {
            return fmt.Errorf("failed to get APIManager: %w", err)
        }
    }

    return nil
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePrometheusRules(desired *monitoringv1.PrometheusRule, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.HasPrometheusRules()
	if err != nil {
		return err
	}

	if !kindExists {
		if r.apiManager.IsPrometheusRulesEnabled() {
			errToLog := fmt.Errorf("Error creating prometheusrule object '%s'. Install prometheus-operator in your cluster to create prometheusrule objects", desired.Name)
			r.EventRecorder().Eventf(r.apiManager, v1.EventTypeWarning, "ReconcileError", errToLog.Error())
			r.logger.Error(errToLog, "ReconcileError")
		}
		return nil
	}

	if !r.apiManager.IsPrometheusRulesEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.PrometheusRule{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileServiceMonitor(desired *monitoringv1.ServiceMonitor, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.HasServiceMonitors()
	if err != nil {
		return err
	}

	if !kindExists {
		if r.apiManager.IsMonitoringEnabled() {
			errToLog := fmt.Errorf("Error creating servicemonitor object '%s'. Install prometheus-operator in your cluster to create servicemonitor objects", desired.Name)
			r.EventRecorder().Eventf(r.apiManager, v1.EventTypeWarning, "ReconcileError", errToLog.Error())
			r.logger.Error(errToLog, "ReconcileError")
		}
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.ServiceMonitor{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcilePodMonitor(desired *monitoringv1.PodMonitor, mutateFn reconcilers.MutateFn) error {
	kindExists, err := r.HasPodMonitors()
	if err != nil {
		return err
	}

	if !kindExists {
		if r.apiManager.IsMonitoringEnabled() {
			errToLog := fmt.Errorf("Error creating podmonitor object '%s'. Install prometheus-operator in your cluster to create podmonitor objects", desired.Name)
			r.EventRecorder().Eventf(r.apiManager, v1.EventTypeWarning, "ReconcileError", errToLog.Error())
			r.logger.Error(errToLog, "ReconcileError")
		}
		return nil
	}

	if !r.apiManager.IsMonitoringEnabled() {
		common.TagObjectToDelete(desired)
	}
	return r.ReconcileResource(&monitoringv1.PodMonitor{}, desired, mutateFn)
}

func (r *BaseAPIManagerLogicReconciler) ReconcileResource(obj, desired common.KubernetesObject, mutatefn reconcilers.MutateFn) error {
	desired.SetNamespace(r.apiManager.GetNamespace())

	// Secrets are managed by users so they do not get APIManager-based
	// owned references. In case we want to react to changes to secrets
	// in the future we will need to implement an alternative mechanism to
	// controller-based OwnerReferences due to user-managed secrets might
	// already have controller-based OwnerReferences and K8s objects
	// can only be owned by a single controller-based OwnerReference.
	if desired.GetObjectKind().GroupVersionKind().Kind != "Secret" {
		if err := r.SetOwnerReference(r.apiManager, desired); err != nil {
			return err
		}
	}

	return r.BaseReconciler.ReconcileResource(obj, desired, r.APIManagerMutator(mutatefn))
}

// APIManagerMutator wraps mutator into APIManger mutator
// All resources managed by APIManager are processed by this wrapped mutator
func (r *BaseAPIManagerLogicReconciler) APIManagerMutator(mutateFn reconcilers.MutateFn) reconcilers.MutateFn {
	return func(existing, desired common.KubernetesObject) (bool, error) {
		// Metadata
		updated := helper.EnsureObjectMeta(existing, desired)

		// Secrets are managed by users so they do not get APIManager-based
		// owned references. In case we want to react to changes to secrets
		// in the future we will need to implement an alternative mechanism to
		// controller-based OwnerReferences due to user-managed secrets might
		// already have controller-based OwnerReferences and K8s objects
		// can only be owned by a single controller-based OwnerReference.
		if existing.GetObjectKind().GroupVersionKind().Kind != "Secret" {
			// OwnerRefenrence
			updatedTmp, err := r.EnsureOwnerReference(r.apiManager, existing)
			if err != nil {
				return false, err
			}
			updated = updated || updatedTmp
		}

		updatedTmp, err := mutateFn(existing, desired)
		if err != nil {
			return false, err
		}
		updated = updated || updatedTmp

		return updated, nil
	}
}

func (r *BaseAPIManagerLogicReconciler) Logger() logr.Logger {
	return r.logger
}

func (b *BaseAPIManagerLogicReconciler) HasGrafanaDashboards() (bool, error) {
	if b.crdAvailabilityCache.grafanaDashboardCRDAvailable == nil {
		res, err := b.BaseReconciler.HasGrafanaDashboards()
		if err != nil {
			return res, err
		}
		b.crdAvailabilityCache.grafanaDashboardCRDAvailable = &res
		return res, err
	}

	return *b.crdAvailabilityCache.grafanaDashboardCRDAvailable, nil
}

// HasPrometheusRules checks if the PrometheusRules CRD is supported in current cluster
func (b *BaseAPIManagerLogicReconciler) HasPrometheusRules() (bool, error) {
	if b.crdAvailabilityCache.prometheusRuleCRDAvailable == nil {
		res, err := b.BaseReconciler.HasPrometheusRules()
		if err != nil {
			return res, err
		}
		b.crdAvailabilityCache.prometheusRuleCRDAvailable = &res
		return res, err
	}
	return *b.crdAvailabilityCache.prometheusRuleCRDAvailable, nil
}

func (b *BaseAPIManagerLogicReconciler) HasServiceMonitors() (bool, error) {
	if b.crdAvailabilityCache.serviceMonitorCRDAvailable == nil {
		res, err := b.BaseReconciler.HasServiceMonitors()
		if err != nil {
			return res, err
		}
		b.crdAvailabilityCache.serviceMonitorCRDAvailable = &res
		return res, err
	}
	return *b.crdAvailabilityCache.serviceMonitorCRDAvailable, nil
}

func (b *BaseAPIManagerLogicReconciler) HasPodMonitors() (bool, error) {
	if b.crdAvailabilityCache.podMonitorCRDAvailable == nil {
		res, err := b.BaseReconciler.HasPodMonitors()
		if err != nil {
			return res, err
		}
		b.crdAvailabilityCache.podMonitorCRDAvailable = &res
		return res, err
	}
	return *b.crdAvailabilityCache.podMonitorCRDAvailable, nil
}
