// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIManager) DeepCopyInto(out *APIManager) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIManager.
func (in *APIManager) DeepCopy() *APIManager {
	if in == nil {
		return nil
	}
	out := new(APIManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *APIManager) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIManagerCommonSpec) DeepCopyInto(out *APIManagerCommonSpec) {
	*out = *in
	if in.AppLabel != nil {
		in, out := &in.AppLabel, &out.AppLabel
		*out = new(string)
		**out = **in
	}
	if in.TenantName != nil {
		in, out := &in.TenantName, &out.TenantName
		*out = new(string)
		**out = **in
	}
	if in.ImageStreamTagImportInsecure != nil {
		in, out := &in.ImageStreamTagImportInsecure, &out.ImageStreamTagImportInsecure
		*out = new(bool)
		**out = **in
	}
	if in.ResourceRequirementsEnabled != nil {
		in, out := &in.ResourceRequirementsEnabled, &out.ResourceRequirementsEnabled
		*out = new(bool)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIManagerCommonSpec.
func (in *APIManagerCommonSpec) DeepCopy() *APIManagerCommonSpec {
	if in == nil {
		return nil
	}
	out := new(APIManagerCommonSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIManagerCondition) DeepCopyInto(out *APIManagerCondition) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIManagerCondition.
func (in *APIManagerCondition) DeepCopy() *APIManagerCondition {
	if in == nil {
		return nil
	}
	out := new(APIManagerCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIManagerList) DeepCopyInto(out *APIManagerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]APIManager, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIManagerList.
func (in *APIManagerList) DeepCopy() *APIManagerList {
	if in == nil {
		return nil
	}
	out := new(APIManagerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *APIManagerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIManagerSpec) DeepCopyInto(out *APIManagerSpec) {
	*out = *in
	in.APIManagerCommonSpec.DeepCopyInto(&out.APIManagerCommonSpec)
	if in.Apicast != nil {
		in, out := &in.Apicast, &out.Apicast
		*out = new(ApicastSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Backend != nil {
		in, out := &in.Backend, &out.Backend
		*out = new(BackendSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.System != nil {
		in, out := &in.System, &out.System
		*out = new(SystemSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Zync != nil {
		in, out := &in.Zync, &out.Zync
		*out = new(ZyncSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.HighAvailability != nil {
		in, out := &in.HighAvailability, &out.HighAvailability
		*out = new(HighAvailabilitySpec)
		**out = **in
	}
	if in.PodDisruptionBudget != nil {
		in, out := &in.PodDisruptionBudget, &out.PodDisruptionBudget
		*out = new(PodDisruptionBudgetSpec)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIManagerSpec.
func (in *APIManagerSpec) DeepCopy() *APIManagerSpec {
	if in == nil {
		return nil
	}
	out := new(APIManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIManagerStatus) DeepCopyInto(out *APIManagerStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]APIManagerCondition, len(*in))
		copy(*out, *in)
	}
	in.Deployments.DeepCopyInto(&out.Deployments)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIManagerStatus.
func (in *APIManagerStatus) DeepCopy() *APIManagerStatus {
	if in == nil {
		return nil
	}
	out := new(APIManagerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApicastProductionSpec) DeepCopyInto(out *ApicastProductionSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApicastProductionSpec.
func (in *ApicastProductionSpec) DeepCopy() *ApicastProductionSpec {
	if in == nil {
		return nil
	}
	out := new(ApicastProductionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApicastSpec) DeepCopyInto(out *ApicastSpec) {
	*out = *in
	if in.ApicastManagementAPI != nil {
		in, out := &in.ApicastManagementAPI, &out.ApicastManagementAPI
		*out = new(string)
		**out = **in
	}
	if in.OpenSSLVerify != nil {
		in, out := &in.OpenSSLVerify, &out.OpenSSLVerify
		*out = new(bool)
		**out = **in
	}
	if in.IncludeResponseCodes != nil {
		in, out := &in.IncludeResponseCodes, &out.IncludeResponseCodes
		*out = new(bool)
		**out = **in
	}
	if in.RegistryURL != nil {
		in, out := &in.RegistryURL, &out.RegistryURL
		*out = new(string)
		**out = **in
	}
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.ProductionSpec != nil {
		in, out := &in.ProductionSpec, &out.ProductionSpec
		*out = new(ApicastProductionSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.StagingSpec != nil {
		in, out := &in.StagingSpec, &out.StagingSpec
		*out = new(ApicastStagingSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApicastSpec.
func (in *ApicastSpec) DeepCopy() *ApicastSpec {
	if in == nil {
		return nil
	}
	out := new(ApicastSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApicastStagingSpec) DeepCopyInto(out *ApicastStagingSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApicastStagingSpec.
func (in *ApicastStagingSpec) DeepCopy() *ApicastStagingSpec {
	if in == nil {
		return nil
	}
	out := new(ApicastStagingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendCronSpec) DeepCopyInto(out *BackendCronSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendCronSpec.
func (in *BackendCronSpec) DeepCopy() *BackendCronSpec {
	if in == nil {
		return nil
	}
	out := new(BackendCronSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendListenerSpec) DeepCopyInto(out *BackendListenerSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendListenerSpec.
func (in *BackendListenerSpec) DeepCopy() *BackendListenerSpec {
	if in == nil {
		return nil
	}
	out := new(BackendListenerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendRedisPersistentVolumeClaimSpec) DeepCopyInto(out *BackendRedisPersistentVolumeClaimSpec) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendRedisPersistentVolumeClaimSpec.
func (in *BackendRedisPersistentVolumeClaimSpec) DeepCopy() *BackendRedisPersistentVolumeClaimSpec {
	if in == nil {
		return nil
	}
	out := new(BackendRedisPersistentVolumeClaimSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendSpec) DeepCopyInto(out *BackendSpec) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.RedisImage != nil {
		in, out := &in.RedisImage, &out.RedisImage
		*out = new(string)
		**out = **in
	}
	if in.RedisPersistentVolumeClaimSpec != nil {
		in, out := &in.RedisPersistentVolumeClaimSpec, &out.RedisPersistentVolumeClaimSpec
		*out = new(BackendRedisPersistentVolumeClaimSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.ListenerSpec != nil {
		in, out := &in.ListenerSpec, &out.ListenerSpec
		*out = new(BackendListenerSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.WorkerSpec != nil {
		in, out := &in.WorkerSpec, &out.WorkerSpec
		*out = new(BackendWorkerSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.CronSpec != nil {
		in, out := &in.CronSpec, &out.CronSpec
		*out = new(BackendCronSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendSpec.
func (in *BackendSpec) DeepCopy() *BackendSpec {
	if in == nil {
		return nil
	}
	out := new(BackendSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BackendWorkerSpec) DeepCopyInto(out *BackendWorkerSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BackendWorkerSpec.
func (in *BackendWorkerSpec) DeepCopy() *BackendWorkerSpec {
	if in == nil {
		return nil
	}
	out := new(BackendWorkerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeprecatedSystemS3Spec) DeepCopyInto(out *DeprecatedSystemS3Spec) {
	*out = *in
	out.AWSCredentials = in.AWSCredentials
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeprecatedSystemS3Spec.
func (in *DeprecatedSystemS3Spec) DeepCopy() *DeprecatedSystemS3Spec {
	if in == nil {
		return nil
	}
	out := new(DeprecatedSystemS3Spec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HighAvailabilitySpec) DeepCopyInto(out *HighAvailabilitySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HighAvailabilitySpec.
func (in *HighAvailabilitySpec) DeepCopy() *HighAvailabilitySpec {
	if in == nil {
		return nil
	}
	out := new(HighAvailabilitySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodDisruptionBudgetSpec) DeepCopyInto(out *PodDisruptionBudgetSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodDisruptionBudgetSpec.
func (in *PodDisruptionBudgetSpec) DeepCopy() *PodDisruptionBudgetSpec {
	if in == nil {
		return nil
	}
	out := new(PodDisruptionBudgetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemAppSpec) DeepCopyInto(out *SystemAppSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemAppSpec.
func (in *SystemAppSpec) DeepCopy() *SystemAppSpec {
	if in == nil {
		return nil
	}
	out := new(SystemAppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemDatabaseSpec) DeepCopyInto(out *SystemDatabaseSpec) {
	*out = *in
	if in.MySQL != nil {
		in, out := &in.MySQL, &out.MySQL
		*out = new(SystemMySQLSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.PostgreSQL != nil {
		in, out := &in.PostgreSQL, &out.PostgreSQL
		*out = new(SystemPostgreSQLSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemDatabaseSpec.
func (in *SystemDatabaseSpec) DeepCopy() *SystemDatabaseSpec {
	if in == nil {
		return nil
	}
	out := new(SystemDatabaseSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemFileStorageSpec) DeepCopyInto(out *SystemFileStorageSpec) {
	*out = *in
	if in.PVC != nil {
		in, out := &in.PVC, &out.PVC
		*out = new(SystemPVCSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.DeprecatedS3 != nil {
		in, out := &in.DeprecatedS3, &out.DeprecatedS3
		*out = new(DeprecatedSystemS3Spec)
		**out = **in
	}
	if in.S3 != nil {
		in, out := &in.S3, &out.S3
		*out = new(SystemS3Spec)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemFileStorageSpec.
func (in *SystemFileStorageSpec) DeepCopy() *SystemFileStorageSpec {
	if in == nil {
		return nil
	}
	out := new(SystemFileStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemMySQLPVCSpec) DeepCopyInto(out *SystemMySQLPVCSpec) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemMySQLPVCSpec.
func (in *SystemMySQLPVCSpec) DeepCopy() *SystemMySQLPVCSpec {
	if in == nil {
		return nil
	}
	out := new(SystemMySQLPVCSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemMySQLSpec) DeepCopyInto(out *SystemMySQLSpec) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.PersistentVolumeClaimSpec != nil {
		in, out := &in.PersistentVolumeClaimSpec, &out.PersistentVolumeClaimSpec
		*out = new(SystemMySQLPVCSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemMySQLSpec.
func (in *SystemMySQLSpec) DeepCopy() *SystemMySQLSpec {
	if in == nil {
		return nil
	}
	out := new(SystemMySQLSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemPVCSpec) DeepCopyInto(out *SystemPVCSpec) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemPVCSpec.
func (in *SystemPVCSpec) DeepCopy() *SystemPVCSpec {
	if in == nil {
		return nil
	}
	out := new(SystemPVCSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemPostgreSQLPVCSpec) DeepCopyInto(out *SystemPostgreSQLPVCSpec) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemPostgreSQLPVCSpec.
func (in *SystemPostgreSQLPVCSpec) DeepCopy() *SystemPostgreSQLPVCSpec {
	if in == nil {
		return nil
	}
	out := new(SystemPostgreSQLPVCSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemPostgreSQLSpec) DeepCopyInto(out *SystemPostgreSQLSpec) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.PersistentVolumeClaimSpec != nil {
		in, out := &in.PersistentVolumeClaimSpec, &out.PersistentVolumeClaimSpec
		*out = new(SystemPostgreSQLPVCSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemPostgreSQLSpec.
func (in *SystemPostgreSQLSpec) DeepCopy() *SystemPostgreSQLSpec {
	if in == nil {
		return nil
	}
	out := new(SystemPostgreSQLSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemRedisPersistentVolumeClaimSpec) DeepCopyInto(out *SystemRedisPersistentVolumeClaimSpec) {
	*out = *in
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemRedisPersistentVolumeClaimSpec.
func (in *SystemRedisPersistentVolumeClaimSpec) DeepCopy() *SystemRedisPersistentVolumeClaimSpec {
	if in == nil {
		return nil
	}
	out := new(SystemRedisPersistentVolumeClaimSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemS3Spec) DeepCopyInto(out *SystemS3Spec) {
	*out = *in
	out.ConfigurationSecretRef = in.ConfigurationSecretRef
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemS3Spec.
func (in *SystemS3Spec) DeepCopy() *SystemS3Spec {
	if in == nil {
		return nil
	}
	out := new(SystemS3Spec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemSidekiqSpec) DeepCopyInto(out *SystemSidekiqSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemSidekiqSpec.
func (in *SystemSidekiqSpec) DeepCopy() *SystemSidekiqSpec {
	if in == nil {
		return nil
	}
	out := new(SystemSidekiqSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemSpec) DeepCopyInto(out *SystemSpec) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.MemcachedImage != nil {
		in, out := &in.MemcachedImage, &out.MemcachedImage
		*out = new(string)
		**out = **in
	}
	if in.RedisImage != nil {
		in, out := &in.RedisImage, &out.RedisImage
		*out = new(string)
		**out = **in
	}
	if in.RedisPersistentVolumeClaimSpec != nil {
		in, out := &in.RedisPersistentVolumeClaimSpec, &out.RedisPersistentVolumeClaimSpec
		*out = new(SystemRedisPersistentVolumeClaimSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.FileStorageSpec != nil {
		in, out := &in.FileStorageSpec, &out.FileStorageSpec
		*out = new(SystemFileStorageSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.DatabaseSpec != nil {
		in, out := &in.DatabaseSpec, &out.DatabaseSpec
		*out = new(SystemDatabaseSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.AppSpec != nil {
		in, out := &in.AppSpec, &out.AppSpec
		*out = new(SystemAppSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.SidekiqSpec != nil {
		in, out := &in.SidekiqSpec, &out.SidekiqSpec
		*out = new(SystemSidekiqSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemSpec.
func (in *SystemSpec) DeepCopy() *SystemSpec {
	if in == nil {
		return nil
	}
	out := new(SystemSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZyncAppSpec) DeepCopyInto(out *ZyncAppSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZyncAppSpec.
func (in *ZyncAppSpec) DeepCopy() *ZyncAppSpec {
	if in == nil {
		return nil
	}
	out := new(ZyncAppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZyncQueSpec) DeepCopyInto(out *ZyncQueSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZyncQueSpec.
func (in *ZyncQueSpec) DeepCopy() *ZyncQueSpec {
	if in == nil {
		return nil
	}
	out := new(ZyncQueSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZyncSpec) DeepCopyInto(out *ZyncSpec) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.PostgreSQLImage != nil {
		in, out := &in.PostgreSQLImage, &out.PostgreSQLImage
		*out = new(string)
		**out = **in
	}
	if in.AppSpec != nil {
		in, out := &in.AppSpec, &out.AppSpec
		*out = new(ZyncAppSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.QueSpec != nil {
		in, out := &in.QueSpec, &out.QueSpec
		*out = new(ZyncQueSpec)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZyncSpec.
func (in *ZyncSpec) DeepCopy() *ZyncSpec {
	if in == nil {
		return nil
	}
	out := new(ZyncSpec)
	in.DeepCopyInto(out)
	return out
}
