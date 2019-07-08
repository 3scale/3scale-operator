package operator

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator/resourcemerge"
	"github.com/3scale/3scale-operator/pkg/common"
	"github.com/google/go-cmp/cmp"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BackendReconciler struct {
	BaseAPIManagerLogicReconciler
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ LogicReconciler = &BackendReconciler{}

func NewBackendReconciler(baseAPIManagerLogicReconciler BaseAPIManagerLogicReconciler) BackendReconciler {
	return BackendReconciler{
		BaseAPIManagerLogicReconciler: baseAPIManagerLogicReconciler,
	}
}

func (r *BackendReconciler) Reconcile() (reconcile.Result, error) {
	backend, err := r.backend()
	if err != nil {
		return reconcile.Result{}, err
	}

	// // TODO finish reconciliations
	r.reconcileCronDeploymentConfig(backend.CronDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileListenerDeploymentConfig(backend.ListenerDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileListenerService(backend.ListenerService())
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.reconcileListenerRoute(backend.ListenerRoute())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileWorkerDeploymentConfig(backend.WorkerDeploymentConfig())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileEnvironmentConfigMap(backend.EnvironmentConfigMap())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileInternalAPISecretForSystem(backend.InternalAPISecretForSystem())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileRedisSecret(backend.RedisSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	r.reconcileListenerSecret(backend.ListenerSecret())
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *BackendReconciler) backend() (*component.Backend, error) {
	optsProvider := OperatorBackendOptionsProvider{APIManagerSpec: &r.apiManager.Spec, Namespace: r.apiManager.Namespace, Client: r.Client()}
	opts, err := optsProvider.GetBackendOptions()
	if err != nil {
		return nil, err
	}
	return component.NewBackend(opts), nil
}

func (r *BackendReconciler) reconcileDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	err := r.InitializeAsAPIManagerObject(desiredDeploymentConfig)
	if err != nil {
		return err
	}

	objectInfo := r.ObjectInfo(desiredDeploymentConfig)
	existingDeploymentConfig := &appsv1.DeploymentConfig{}
	err = r.Client().Get(context.TODO(), r.NamespacedNameWithAPIManagerNamespace(desiredDeploymentConfig), existingDeploymentConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredDeploymentConfig)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	// Set in the desired some fields that are automatically set
	// by Kubernetes controllers as defaults that are not defined in
	// our logic
	// TODO

	// reconcile update
	needsUpdate, err := r.ensureDeploymentConfig(existingDeploymentConfig, desiredDeploymentConfig)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating DeploymentConfig %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingDeploymentConfig)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating DeploymentConfig %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *BackendReconciler) reconcileCronDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *BackendReconciler) ensureDeploymentConfig(updated, desired *appsv1.DeploymentConfig) (bool, error) {
	changed := false

	objectMetaChanged, err := resourcemerge.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta, r.apiManager, r.Scheme())
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// Reconcile PodTemplateSpec labels
	if resourcemerge.EnsureLabels(updated.Spec.Template.ObjectMeta.Labels, desired.Spec.Template.ObjectMeta.Labels) {
		changed = true
	}

	r.ensurePodContainers(updated.Spec.Template.Spec.Containers, desired.Spec.Template.Spec.Containers)
	if !reflect.DeepEqual(updated.Spec.Template.Spec.Containers, desired.Spec.Template.Spec.Containers) {
		fmt.Println(cmp.Diff(updated.Spec.Template.Spec.Containers, desired.Spec.Template.Spec.Containers, cmpopts.IgnoreUnexported(resource.Quantity{})))
		updated.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		changed = true
	}

	r.ensurePodContainers(updated.Spec.Template.Spec.InitContainers, desired.Spec.Template.Spec.InitContainers)
	if !reflect.DeepEqual(updated.Spec.Template.Spec.InitContainers, desired.Spec.Template.Spec.InitContainers) {
		fmt.Println(cmp.Diff(updated.Spec.Template.Spec.InitContainers, desired.Spec.Template.Spec.InitContainers, cmpopts.IgnoreUnexported(resource.Quantity{})))
		updated.Spec.Template.Spec.InitContainers = desired.Spec.Template.Spec.InitContainers
		changed = true
	}

	r.ensureDeploymentConfigStrategy(&updated.Spec.Strategy, &desired.Spec.Strategy)
	if !reflect.DeepEqual(updated.Spec.Strategy, desired.Spec.Strategy) {
		updated.Spec.Strategy = desired.Spec.Strategy
		changed = true
	}

	return changed, nil
}

func (r *BackendReconciler) ensureDeploymentConfigStrategy(updated, desired *appsv1.DeploymentStrategy) {
	if desired.ActiveDeadlineSeconds == nil {
		desired.ActiveDeadlineSeconds = updated.ActiveDeadlineSeconds
	}
}

func (r *BackendReconciler) ensurePodContainers(updated, desired []v1.Container) {
	updatedContainerMap := map[string]*v1.Container{}
	for idx := range updated {
		container := &updated[idx]
		updatedContainerMap[container.Name] = container
	}

	for idx := range desired {
		desiredContainer := &desired[idx]
		if updatedContainer, ok := updatedContainerMap[desiredContainer.Name]; ok {
			desiredContainer.Image = updatedContainer.Image

			if desiredContainer.TerminationMessagePath == "" {
				desiredContainer.TerminationMessagePath = updatedContainer.TerminationMessagePath
			}

			if desiredContainer.TerminationMessagePolicy == v1.TerminationMessagePolicy("") {
				desiredContainer.TerminationMessagePolicy = updatedContainer.TerminationMessagePolicy
			}

			if desiredContainer.ImagePullPolicy == v1.PullPolicy("") {
				desiredContainer.ImagePullPolicy = updatedContainer.ImagePullPolicy
			}

			if desiredContainer.ReadinessProbe != nil && updatedContainer.ReadinessProbe != nil {
				r.ensureProbe(updatedContainer.ReadinessProbe, desiredContainer.ReadinessProbe)
			}
			if desiredContainer.LivenessProbe != nil && updatedContainer.LivenessProbe != nil {
				r.ensureProbe(updatedContainer.LivenessProbe, desiredContainer.LivenessProbe)
			}

			r.ensureResourceRequirements(&updatedContainer.Resources, &desiredContainer.Resources)
		}
	}

	sort.Slice(updated, func(i, j int) bool { return updated[i].Name < updated[j].Name })
	sort.Slice(desired, func(i, j int) bool { return desired[i].Name < desired[j].Name })
}

func (r *BackendReconciler) ensureResourceRequirements(updated, desired *v1.ResourceRequirements) {
	r.ensureResourceList(updated.Limits, desired.Limits)
	r.ensureResourceList(updated.Requests, desired.Requests)
}

// This method assigns the value of desired to updated when the resource quantities
// are not nil in both sides and are equal. The reason for this is
// because although they are equal internally are stored differently (the units
// might be expressed in a different way) and executing DeepEqual returns
// different even though they are "logically" equal
func (r *BackendReconciler) ensureResourceList(updated, desired v1.ResourceList) {
	if !desired.Cpu().IsZero() && !updated.Cpu().IsZero() &&
		desired.Cpu().Cmp(*updated.Cpu()) == 0 {
		desired[v1.ResourceCPU] = *updated.Cpu()
	}
	if !desired.Memory().IsZero() && !updated.Memory().IsZero() &&
		desired.Memory().Cmp(*updated.Memory()) == 0 {
		desired[v1.ResourceMemory] = *updated.Memory()
	}
	if !desired.Pods().IsZero() && !updated.Pods().IsZero() &&
		desired.Pods().Cmp(*updated.Pods()) == 0 {
		desired[v1.ResourcePods] = *updated.Pods()
	}
	if !desired.StorageEphemeral().IsZero() && !updated.StorageEphemeral().IsZero() &&
		desired.StorageEphemeral().Cmp(*updated.StorageEphemeral()) == 0 {
		desired[v1.ResourceEphemeralStorage] = *updated.StorageEphemeral()
	}
}

func (r *BackendReconciler) ensureProbe(updated, desired *v1.Probe) {
	if desired.TimeoutSeconds == 0 {
		desired.TimeoutSeconds = updated.TimeoutSeconds
	}

	if desired.PeriodSeconds == 0 {
		desired.PeriodSeconds = updated.PeriodSeconds
	}

	if desired.SuccessThreshold == 0 {
		desired.SuccessThreshold = updated.SuccessThreshold
	}

	if desired.FailureThreshold == 0 {
		desired.FailureThreshold = updated.FailureThreshold
	}

	if desired.Handler.HTTPGet != nil && updated.Handler.HTTPGet != nil {
		if desired.Handler.HTTPGet.Scheme == v1.URIScheme("") {
			desired.Handler.HTTPGet.Scheme = v1.URIScheme("HTTP")
		}
	}
}

func (r *BackendReconciler) ensurePodEnvVars(updated, desired []v1.EnvVar) bool {
	changed := false

	return changed
}

func (r *BackendReconciler) reconcileListenerDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *BackendReconciler) SetOwnerReference(obj common.KubernetesObject) error {
	err := controllerutil.SetControllerReference(r.apiManager, obj, r.Scheme())
	if err != nil {
		r.Logger().Error(err, "Error setting OwnerReference on object",
			"Kind", obj.GetObjectKind().GroupVersionKind().String(),
			"Namespace", obj.GetNamespace(),
			"Name", obj.GetName(),
		)
	}
	return err
}

func (r *BackendReconciler) ensureService(updated, desired *v1.Service) (bool, error) {
	changed := false

	objectMetaChanged, err := resourcemerge.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta, r.apiManager, r.Scheme())
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	desired.Spec.ClusterIP = updated.Spec.ClusterIP
	desired.Spec.Type = updated.Spec.Type
	desired.Spec.SessionAffinity = updated.Spec.SessionAffinity

	if !reflect.DeepEqual(updated.Spec, desired.Spec) {
		updated.Spec = desired.Spec
		changed = true
	}

	return changed, nil
}

func (r *BackendReconciler) reconcileListenerService(desiredService *v1.Service) error {
	return r.reconcileService(desiredService)
}

func (r *BackendReconciler) reconcileService(desiredService *v1.Service) error {
	err := r.InitializeAsAPIManagerObject(desiredService)
	if err != nil {
		return err
	}
	objectInfo := r.ObjectInfo(desiredService)

	existingService := &v1.Service{}
	err = r.Client().Get(context.TODO(), r.NamespacedNameWithAPIManagerNamespace(desiredService), existingService)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredService)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureService(existingService, desiredService)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating Service %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingService)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Service %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *BackendReconciler) reconcileRoute(desiredRoute *routev1.Route) error {
	err := r.InitializeAsAPIManagerObject(desiredRoute)
	if err != nil {
		return err
	}

	objectInfo := r.ObjectInfo(desiredRoute)

	existingRoute := &routev1.Route{}
	err = r.Client().Get(context.TODO(), r.NamespacedNameWithAPIManagerNamespace(desiredRoute), existingRoute)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredRoute)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureRoute(existingRoute, desiredRoute)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating Route %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingRoute)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Service %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *BackendReconciler) ensureRoute(updated, desired *routev1.Route) (bool, error) {
	changed := false

	objectMetaChanged, err := resourcemerge.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta, r.apiManager, r.Scheme())
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// Set in the desired some fields that are automatically set
	// by Kubernetes controllers as defaults that are not defined in
	// our logic
	desired.Spec.WildcardPolicy = updated.Spec.WildcardPolicy
	desired.Spec.To.Weight = updated.Spec.To.Weight

	if !reflect.DeepEqual(updated.Spec, desired.Spec) {
		updated.Spec = desired.Spec
		changed = true
	}

	return changed, nil
}

func (r *BackendReconciler) reconcileListenerRoute(desiredRoute *routev1.Route) error {
	return r.reconcileRoute(desiredRoute)
}

func (r *BackendReconciler) reconcileWorkerDeploymentConfig(desiredDeploymentConfig *appsv1.DeploymentConfig) error {
	return r.reconcileDeploymentConfig(desiredDeploymentConfig)
}

func (r *BackendReconciler) reconcileConfigMap(desiredConfigMap *v1.ConfigMap) error {
	err := r.InitializeAsAPIManagerObject(desiredConfigMap)
	if err != nil {
		return err
	}

	objectInfo := r.ObjectInfo(desiredConfigMap)
	existingConfigMap := &v1.ConfigMap{}
	err = r.Client().Get(context.TODO(), r.NamespacedNameWithAPIManagerNamespace(desiredConfigMap), existingConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredConfigMap)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureConfigMap(existingConfigMap, desiredConfigMap)
	if err != nil {
		return err
	}

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating ConfigMap %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingConfigMap)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Service %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *BackendReconciler) reconcileEnvironmentConfigMap(desiredConfigMap *v1.ConfigMap) error {
	return r.reconcileConfigMap(desiredConfigMap)
}

func (r *BackendReconciler) ensureConfigMap(updated, desired *v1.ConfigMap) (bool, error) {
	changed := false

	objectMetaChanged, err := resourcemerge.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta, r.apiManager, r.Scheme())
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// TODO should be the reconciliation of ConfigMap data a merge behavior
	// instead of a replace one?
	// TODO should we reconcile BinaryData too???
	if !reflect.DeepEqual(updated.Data, desired.Data) {
		updated.Data = desired.Data
		changed = true
	}

	return changed, nil
}

func (r *BackendReconciler) reconcileInternalAPISecretForSystem(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *BackendReconciler) reconcileRedisSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

func (r *BackendReconciler) reconcileListenerSecret(desiredSecret *v1.Secret) error {
	return r.reconcileSecret(desiredSecret)
}

// TODO should this be a shared Secret reconcile behaviour or could there be
// different secrets where we reconcile them in a different way?
// For example, secrets where we just want to check if they are installed it
// but we don't want to reconcile them. That would make sense for example
// for system-seed secret.
// Also in the case where we wanted to reconcile them but we wanted to do it
// in a different way
func (r *BackendReconciler) reconcileSecret(desiredSecret *v1.Secret) error {
	err := r.InitializeAsAPIManagerObject(desiredSecret)
	if err != nil {
		return err
	}

	objectInfo := r.ObjectInfo(desiredSecret)
	existingSecret := &v1.Secret{}
	err = r.Client().Get(context.TODO(), r.NamespacedNameWithAPIManagerNamespace(desiredSecret), existingSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			createErr := r.Client().Create(context.TODO(), desiredSecret)
			if createErr != nil {
				r.Logger().Error(createErr, fmt.Sprintf("Error creating object %s. Requeuing request...", objectInfo))
				return createErr
			}
			r.Logger().Info(fmt.Sprintf("Created object %s", objectInfo))
			return nil
		}
		return err
	}

	needsUpdate, err := r.ensureSecret(existingSecret, desiredSecret)

	if needsUpdate {
		r.Logger().Info(fmt.Sprintf("Updating Secret %s", objectInfo))
		err := r.Client().Update(context.TODO(), existingSecret)
		if err != nil {
			r.Logger().Error(err, fmt.Sprintf("error updating Service %s", objectInfo))
			return err
		}
	}

	return nil
}

func (r *BackendReconciler) ensureSecret(updated, desired *v1.Secret) (bool, error) {
	changed := false

	objectMetaChanged, err := resourcemerge.EnsureObjectMeta(&updated.ObjectMeta, &desired.ObjectMeta, r.apiManager, r.Scheme())
	if err != nil {
		return false, err
	}
	if objectMetaChanged {
		changed = true
	}

	// TODO writing on StringData has merge behaviour with existing Data content.
	// Do we want this or do we want to always overwrite everything? Take note that
	// in ConfigMap for example it seems there's no merge behavior by itself so
	// either we implement it by ourselves or we would have different behavior
	// between secret and configmap fields reconciliation
	// TODO there's also a case where a StringData field that is set to empty
	// is encoded as the empty string in the Data section but when there's another
	// update it is encoded to null, and that would be detected as a difference.
	// Would that be dangerous/a problem?
	updatedSecretStringData := getSecretStringDataFromData(updated.Data)
	if !reflect.DeepEqual(updatedSecretStringData, desired.StringData) {
		updated.StringData = desired.StringData
		changed = true
	}

	return changed, nil
}

func (r *BackendReconciler) SetNamespace(obj metav1.Object, namespace string) {
	obj.SetNamespace(namespace)
}

func (r *BackendReconciler) ObjectInfo(obj common.KubernetesObject) string {
	return fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
}
