package apicast

import (
	"context"
	"fmt"
	"reflect"

	appsv1operator "github.com/3scale/3scale-operator/pkg/apis/apps"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/3scale/3scale-operator/pkg/k8sutils"
)

const (
	ENV_SECRET_RESVER_ANNOTATION        = "apicast.apps.3scale.net/environment-secret-resource-version"
	ADM_PORTAL_SECRET_RESVER_ANNOTATION = "apicast.apps.3scale.net/admin-portal-secret-resource-version"
)

type Reconciler struct {
	Client   client.Client
	Logger   logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	apicastCR, err := r.getAPIcast(request)
	if err != nil {
		return reconcile.Result{}, err
	}
	// if apicast is nil and we did not have an error it means it does
	// not exist and we don't want to return an error because we
	// don't want to requeue the request
	if apicastCR == nil {
		return reconcile.Result{}, nil
	}

	r.Logger.WithValues("Name", apicastCR.Name, "Namespace", apicastCR.Namespace)

	appliedInitialization, err := r.initialize(apicastCR)
	if err != nil {
		return reconcile.Result{}, err
	}
	if appliedInitialization {
		return reconcile.Result{}, nil
	}

	// TODO this function does a little bit of creating the desiredApicast and
	// also validation. Also validation should be done BEFORE
	// the initialization probably
	desiredAPIcast, err := r.internalAPIcast(apicastCR)
	if err != nil {
		return reconcile.Result{}, err
	}

	desiredAPIcastAdminPortalEndpointSecret, err := desiredAPIcast.AdminPortalEndpointSecret()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileAdditionalEnvSecret(apicastCR)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileAdminPortalEndpointSecret(desiredAPIcastAdminPortalEndpointSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	adminPortalEndpointSecret, err := desiredAPIcast.AdminPortalEndpointSecret()
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileDeployment(*desiredAPIcast.Deployment(), desiredAPIcast.additionalEnvironment, &adminPortalEndpointSecret)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileService(*desiredAPIcast.Service())
	if err != nil {
		return reconcile.Result{}, err
	}

	if apicastCR.Spec.ExposedHostname != nil {
		err = r.reconcileIngress(*desiredAPIcast.Ingress())
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *Reconciler) adminPortalCredentials(existingAPIcast *appsv1alpha1.APIcast) (ApicastAdminPortalCredentials, error) {
	adminPortalURLSecretKeyRef := existingAPIcast.Spec.AdminPortal.URLSecretKeyRef
	if adminPortalURLSecretKeyRef.Name == "" {
		return ApicastAdminPortalCredentials{}, fmt.Errorf("Field 'name' not specified for URL Secret Key Reference")
	}
	adminPortalURLSecretNamespacedName := types.NamespacedName{
		Name:      adminPortalURLSecretKeyRef.Name,
		Namespace: existingAPIcast.Namespace,
	}

	adminPortalURLSecret := v1.Secret{}
	err := r.Client.Get(context.TODO(), adminPortalURLSecretNamespacedName, &adminPortalURLSecret)
	if err != nil {
		return ApicastAdminPortalCredentials{}, err
	}
	secretStringData := k8sutils.SecretStringDataFromData(adminPortalURLSecret)
	adminPortalURL, ok := secretStringData[adminPortalURLSecretKeyRef.Key]
	if !ok {
		return ApicastAdminPortalCredentials{}, fmt.Errorf("Key '%s' not found in secret '%s'", adminPortalURLSecretKeyRef.Key, adminPortalURLSecretKeyRef.Name)
	}

	controllerutil.SetControllerReference(existingAPIcast, &adminPortalURLSecret, r.Scheme)
	err = r.Client.Update(context.TODO(), &adminPortalURLSecret)
	if err != nil {
		return ApicastAdminPortalCredentials{}, err
	}

	adminPortalAccessTokenKeyRef := existingAPIcast.Spec.AdminPortal.AccessTokenSecretKeyRef
	if adminPortalAccessTokenKeyRef.Name == "" {
		return ApicastAdminPortalCredentials{}, fmt.Errorf("Field 'name' not specified for URL Access Token Key Reference")
	}
	adminPortalAccessTokenNamespacedName := types.NamespacedName{
		Name:      adminPortalAccessTokenKeyRef.Name,
		Namespace: existingAPIcast.Namespace,
	}
	adminPortalAccessTokenSecret := v1.Secret{}
	err = r.Client.Get(context.TODO(), adminPortalAccessTokenNamespacedName, &adminPortalAccessTokenSecret)
	if err != nil {
		return ApicastAdminPortalCredentials{}, err
	}

	secretStringData = k8sutils.SecretStringDataFromData(adminPortalAccessTokenSecret)
	adminPortalAccessToken, ok := secretStringData[adminPortalAccessTokenKeyRef.Key]
	if !ok {
		return ApicastAdminPortalCredentials{}, fmt.Errorf("Key '%s' not found in secret '%s'", adminPortalAccessTokenKeyRef.Key, adminPortalAccessTokenKeyRef.Name)
	}

	controllerutil.SetControllerReference(existingAPIcast, &adminPortalAccessTokenSecret, r.Scheme)
	err = r.Client.Update(context.TODO(), &adminPortalAccessTokenSecret)
	if err != nil {
		return ApicastAdminPortalCredentials{}, err
	}

	result := ApicastAdminPortalCredentials{
		URL:         adminPortalURL,
		AccessToken: adminPortalAccessToken,
	}
	return result, nil
}

func (r *Reconciler) checkAdditionalEnvironmentConfigurationSecretRef(existingAPIcast *appsv1alpha1.APIcast) error {
	environmentConfigurationSecretRef := existingAPIcast.Spec.EnvironmentConfigurationSecretRef
	if environmentConfigurationSecretRef == nil {
		return nil
	}

	if environmentConfigurationSecretRef.Name == "" {
		return fmt.Errorf("Field 'name' not specified for Additional Env Secret Reference")
	}

	environmentConfigurationSecretNamespacedName := types.NamespacedName{
		Name:      environmentConfigurationSecretRef.Name,
		Namespace: existingAPIcast.Namespace,
	}

	environmentConfigurationSecret := v1.Secret{}
	err := r.Client.Get(context.TODO(), environmentConfigurationSecretNamespacedName, &environmentConfigurationSecret)
	return err
}

func (r *Reconciler) internalAPIcast(existingAPIcast *appsv1alpha1.APIcast) (Apicast, error) {
	apicastFullName := "apicast-" + existingAPIcast.Name
	apicastExposedHostname := ""
	if existingAPIcast.Spec.ExposedHostname != nil {
		apicastExposedHostname = *existingAPIcast.Spec.ExposedHostname
	}
	apicastOwnerRef := asOwner(existingAPIcast)

	adminPortalCredentials, err := r.adminPortalCredentials(existingAPIcast)
	if err != nil {
		return Apicast{}, err
	}

	err = r.checkAdditionalEnvironmentConfigurationSecretRef(existingAPIcast)
	if err != nil {
		return Apicast{}, err
	}

	internalApicastResult := Apicast{
		deploymentName:         apicastFullName,
		serviceName:            apicastFullName,
		replicas:               int32(*existingAPIcast.Spec.Replicas),
		appLabel:               "apicast",
		serviceAccountName:     *existingAPIcast.Spec.ServiceAccount,
		image:                  *existingAPIcast.Spec.Image,
		exposedHostname:        apicastExposedHostname,
		namespace:              existingAPIcast.Namespace,
		ownerReference:         &apicastOwnerRef,
		additionalEnvironment:  existingAPIcast.Spec.EnvironmentConfigurationSecretRef,
		adminPortalCredentials: adminPortalCredentials,
	}

	return internalApicastResult, err
}

func (r *Reconciler) namespacedName(object metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
	}
}

func (r *Reconciler) getAPIcast(request reconcile.Request) (*appsv1alpha1.APIcast, error) {
	instance := appsv1alpha1.APIcast{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, &instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Reconciler) initialize(apicastCR *appsv1alpha1.APIcast) (bool, error) {
	if appliedSomeInitialization := r.applyInitialization(apicastCR); appliedSomeInitialization {
		err := r.Client.Update(context.TODO(), apicastCR)
		if err != nil {
			return false, err
		}
		r.Logger.Info("APIcast resource missed optional fields. Updated CR which triggered a new reconciliation event")
		// the final effect should be stop the reconciliation cycle, without starting a new one
		// and also NOT continue evaluating logic
		return true, nil
	}
	return false, nil
}

func (r *Reconciler) applyInitialization(apicastCR *appsv1alpha1.APIcast) bool {
	var defaultAPIcastReplicas int64 = 1
	defaultServiceAccount := "default"
	defaultAPIcastImage := "registry.access.redhat.com/3scale-amp25/apicast-gateway"
	appliedInitialization := false

	if apicastCR.Spec.Replicas == nil {
		apicastCR.Spec.Replicas = &defaultAPIcastReplicas
		appliedInitialization = true
	}
	if apicastCR.Spec.ServiceAccount == nil {
		apicastCR.Spec.ServiceAccount = &defaultServiceAccount
		appliedInitialization = true
	}
	if apicastCR.Spec.Image == nil {
		apicastCR.Spec.Image = &defaultAPIcastImage
		appliedInitialization = true
	}

	return appliedInitialization
}

// asOwner returns an owner reference set as the tenant CR
func asOwner(a *appsv1alpha1.APIcast) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: appsv1alpha1.SchemeGroupVersion.String(),
		Kind:       appsv1operator.APICastKind,
		Name:       a.Name,
		UID:        a.UID,
		Controller: &trueVar,
	}
}

func (r *Reconciler) reconcileAdditionalEnvSecret(existingAPIcast *appsv1alpha1.APIcast) error {
	environmentConfigurationSecret := v1.Secret{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: existingAPIcast.Spec.EnvironmentConfigurationSecretRef.Name, Namespace: existingAPIcast.Namespace}, &environmentConfigurationSecret)
	if err != nil {
		return err
	}

	err = controllerutil.SetControllerReference(existingAPIcast, &environmentConfigurationSecret, r.Scheme)

	if err != nil {
		r.Logger.Error(err, "Error setting OwnerReference on object. Requeuing request...",
			"Kind", environmentConfigurationSecret.GetObjectKind(),
			"Namespace", environmentConfigurationSecret.GetNamespace(),
			"Name", environmentConfigurationSecret.GetName(),
		)
		return err
	}

	fmt.Println(environmentConfigurationSecret)
	err = r.Client.Update(context.TODO(), &environmentConfigurationSecret)

	return err
}

func (r *Reconciler) reconcileDeployment(desiredDeployment appsv1.Deployment, additionalEnvironment *v1.SecretEnvSource, adminPortalSecret *v1.Secret) error {
	environmentConfigurationSecret := v1.Secret{}
	if additionalEnvironment != nil {
		err := r.Client.Get(context.TODO(), types.NamespacedName{Name: additionalEnvironment.Name, Namespace: desiredDeployment.Namespace}, &environmentConfigurationSecret)
		if err != nil {
			return err
		}
	}
	adminPortalConfigurationSecret := v1.Secret{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: adminPortalSecret.Name, Namespace: desiredDeployment.Namespace}, &adminPortalConfigurationSecret)
	if err != nil {
		return err
	}

	existingDeployment := appsv1.Deployment{}
	err = r.Client.Get(context.TODO(), r.namespacedName(&desiredDeployment), &existingDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			if additionalEnvironment != nil {
				desiredDeployment.Spec.Template.Annotations[ENV_SECRET_RESVER_ANNOTATION] = string(environmentConfigurationSecret.ResourceVersion)
			}
			desiredDeployment.Spec.Template.Annotations[ADM_PORTAL_SECRET_RESVER_ANNOTATION] = string(adminPortalConfigurationSecret.ResourceVersion)
			err = r.Client.Create(context.TODO(), &desiredDeployment)
			r.Logger.Info("Creating Deployment...")
			return err
		}
		return err
	}

	changed := false

	if !reflect.DeepEqual(existingDeployment.Spec.Replicas, desiredDeployment.Spec.Replicas) {
		existingDeployment.Spec.Replicas = desiredDeployment.Spec.Replicas
		changed = true
	}
	if !reflect.DeepEqual(existingDeployment.Spec.Template.Labels, desiredDeployment.Spec.Template.Labels) {
		existingDeployment.Spec.Template.Labels = desiredDeployment.Spec.Template.Labels
		changed = true
	}
	if !reflect.DeepEqual(existingDeployment.Spec.Template.Spec.Containers[0].Image, desiredDeployment.Spec.Template.Spec.Containers[0].Image) {
		existingDeployment.Spec.Template.Spec.Containers[0].Image = desiredDeployment.Spec.Template.Spec.Containers[0].Image
		changed = true
	}
	if !reflect.DeepEqual(existingDeployment.Spec.Template.Spec.ServiceAccountName, desiredDeployment.Spec.Template.Spec.ServiceAccountName) {
		changed = true
		existingDeployment.Spec.Template.Spec.ServiceAccountName = desiredDeployment.Spec.Template.Spec.ServiceAccountName
	}

	if additionalEnvironment != nil {
		observedEnvSecretResourceGeneration := existingDeployment.Annotations[ENV_SECRET_RESVER_ANNOTATION]
		if !reflect.DeepEqual(observedEnvSecretResourceGeneration, string(environmentConfigurationSecret.Generation)) {
			changed = true
			existingDeployment.Spec.Template.Annotations[ENV_SECRET_RESVER_ANNOTATION] = string(environmentConfigurationSecret.ResourceVersion)
		}
	}

	observedAdminPortalSecretResourceGeneration := existingDeployment.Annotations[ADM_PORTAL_SECRET_RESVER_ANNOTATION]
	if !reflect.DeepEqual(observedAdminPortalSecretResourceGeneration, string(adminPortalConfigurationSecret.ResourceVersion)) {
		changed = true
		existingDeployment.Spec.Template.Annotations[ADM_PORTAL_SECRET_RESVER_ANNOTATION] = string(adminPortalConfigurationSecret.ResourceVersion)
	}

	if changed {
		err = r.Client.Update(context.TODO(), &existingDeployment)
		r.Logger.Info("Updating Deployment...")
		return err
	}

	return nil
}

func (r *Reconciler) reconcileAdminPortalEndpointSecret(desiredAdminPortalSecret v1.Secret) error {
	existingAdminPortalSecret := v1.Secret{}
	err := r.Client.Get(context.TODO(), r.namespacedName(&desiredAdminPortalSecret), &existingAdminPortalSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), &desiredAdminPortalSecret)
			r.Logger.Info("Creating Admin Portal Endpoint Secret...")
		}
		return err
	} else {
		existingAdminPortalSecretStringData := k8sutils.SecretStringDataFromData(existingAdminPortalSecret)
		if !reflect.DeepEqual(existingAdminPortalSecretStringData, desiredAdminPortalSecret.StringData) {
			existingAdminPortalSecret.StringData = desiredAdminPortalSecret.StringData
			err = r.Client.Update(context.TODO(), &existingAdminPortalSecret)
			r.Logger.Info("Updating Admin Portal Endpoint Secret...")
			return err
		}
	}
	return nil
}

func (r *Reconciler) reconcileService(desiredService v1.Service) error {
	existingService := v1.Service{}
	err := r.Client.Get(context.TODO(), r.namespacedName(&desiredService), &existingService)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), &desiredService)
			r.Logger.Info("Creating Service...")
		}
		return err
	} else {
		if !reflect.DeepEqual(existingService.Spec.Ports, desiredService.Spec.Ports) {
			existingService.Spec.Ports = desiredService.Spec.Ports

			err = r.Client.Update(context.TODO(), &existingService)
			r.Logger.Info("Updating Service...")
			if err != nil {
				return err
			}
		}
		return err
	}

}

func (r *Reconciler) reconcileIngress(desiredIngress extensions.Ingress) error {
	existingIngress := extensions.Ingress{}
	err := r.Client.Get(context.TODO(), r.namespacedName(&desiredIngress), &existingIngress)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), &desiredIngress)
			r.Logger.Info("Creating Ingress...")
		}
		return err
	} else {
		if !reflect.DeepEqual(existingIngress.Spec, desiredIngress.Spec) {
			existingIngress.Spec = desiredIngress.Spec
			err = r.Client.Update(context.TODO(), &existingIngress)
			r.Logger.Info("Updating Ingress...")
			if err != nil {
				return err
			}
		}
		return err
	}
}
