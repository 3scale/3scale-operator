package restore

import (
	"fmt"
	"strings"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/backup"
	"github.com/3scale/3scale-operator/pkg/helper"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIManagerRestore struct {
	options *APIManagerRestoreOptions
}

func NewAPIManagerRestore(options *APIManagerRestoreOptions) *APIManagerRestore {
	return &APIManagerRestore{
		options: options,
	}
}

const (
	RestorePVCMountPath           = "/backup"
	SystemFileStoragePVCMountPath = "/system-filestorage-pvc"
)

var secretsToRestore map[string]string = map[string]string{
	"SystemSMTP": "system-smtp",
	"SystemSeed": "system-seed",
	// "SystemDatabase":      "system-database", // We decided this is done by the user. Reason for that is that he might want to restore it in another place
	"BackendInternalAPI": "backend-internal-api",
	// "BackendRedis": "backend-redis", // We decided this is done by the user. Reason for that is that he might want to restore it in another place
	"BackendListener":  "backend-listener",
	"SystemEventsHook": "system-events-hook",
	"SystemApp":        "system-app",
	"SystemRecaptcha":  "system-recaptcha",
	// "SystemRedis":         "system-redis", // We decided this is done by the user. Reason for that is that he might want to restore it in another place
	// "SystemMemcached":     "system-memcache", // TODO should we backup/restore this one?
	"Zync":                "zync",
	"SystemMasterAPIcast": "system-master-apicast",
}

var configMapsToRestore map[string]string = map[string]string{
	"SystemEnvironment":  "system-environment",
	"APIcastEnvironment": "apicast-environment",
}

func (b *APIManagerRestore) restoreSourcePVCContainerVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      b.options.APIManagerRestorePVCOptions.PersistentVolumeClaimVolumeSource.ClaimName,
		MountPath: RestorePVCMountPath,
	}
}

func (b *APIManagerRestore) restoreSourcePVCPodVolume() v1.Volume {
	return v1.Volume{
		Name: b.options.APIManagerRestorePVCOptions.PersistentVolumeClaimVolumeSource.ClaimName,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &b.options.APIManagerRestorePVCOptions.PersistentVolumeClaimVolumeSource,
		},
	}
}

func (b *APIManagerRestore) systemFileStoragePVCContainerVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      component.SystemFileStoragePVCName,
		MountPath: SystemFileStoragePVCMountPath,
	}
}

func (b *APIManagerRestore) systemFileStoragePVCPodVolume() v1.Volume {
	return v1.Volume{
		Name: component.SystemFileStoragePVCName,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: component.SystemFileStoragePVCName,
			},
		},
	}
}

func (b *APIManagerRestore) RestoreSecretsAndConfigMapsFromPVCJob() *batchv1.Job {
	if b.options.APIManagerRestorePVCOptions == nil {
		return nil
	}

	jobName, err := helper.UIDBasedJobName("restore-cfgmaps-secrets", b.options.APIManagerRestoreUID)
	if err != nil {
		panic(err)
	}

	var completions int32 = 1
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: b.options.Namespace,
		},
		Spec: batchv1.JobSpec{
			Completions: &completions,
			// TODO BackoffLimit field controls how many times the job is retried
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						b.restoreSourcePVCPodVolume(),
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "restore-cfgmaps-secrets",
							Image: b.options.OCCLIImageURL,
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"-c",
								"-e",
								b.restoreSecretsAndConfigMapsContainerArgs(),
							},
							//Env: []v1.EnvVar{},
							VolumeMounts: []v1.VolumeMount{
								b.restoreSourcePVCContainerVolumeMount(),
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever, // Only "Never" or "OnFailure" are accepted in Kubernetes Jobs
				},
			},
		},
	}
}

func (b *APIManagerRestore) RestoreSystemFileStoragePVCFromPVCJob() *batchv1.Job {
	if b.options.APIManagerRestorePVCOptions == nil {
		return nil
	}

	jobName, err := helper.UIDBasedJobName("restore-system-fs", b.options.APIManagerRestoreUID)
	if err != nil {
		panic(err)
	}

	var completions int32 = 1
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: b.options.Namespace,
		},
		Spec: batchv1.JobSpec{
			Completions: &completions,
			// TODO BackoffLimit field controls how many times the job is retried. Should we limit to 1?
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						b.restoreSourcePVCPodVolume(),
						b.systemFileStoragePVCPodVolume(),
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "backup-system-filestorage-pvc",
							Image: b.options.OCCLIImageURL,
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"-c",
								"-e",
								b.restoreSystemFilestoragePVCContainerArgs(),
							},
							//Env: []v1.EnvVar{},
							VolumeMounts: []v1.VolumeMount{
								b.restoreSourcePVCContainerVolumeMount(),
								b.systemFileStoragePVCContainerVolumeMount(),
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever, // Only "Never" or "OnFailure" are accepted in Kubernetes Jobs
				},
			},
		},
	}
}

func (b *APIManagerRestore) CreateAPIManagerSharedSecretJob() *batchv1.Job {
	if b.options.APIManagerRestorePVCOptions == nil {
		return nil
	}

	jobName, err := helper.UIDBasedJobName("restore-apm-tosecret", b.options.APIManagerRestoreUID)
	if err != nil {
		panic(err)
	}

	var completions int32 = 1
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: b.options.Namespace,
		},
		Spec: batchv1.JobSpec{
			Completions: &completions,
			// TODO BackoffLimit field controls how many times the job is retried
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						b.restoreSourcePVCPodVolume(),
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "job",
							Image: b.options.OCCLIImageURL,
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"-c",
								"-e",
								b.createAPIManagerSharedSecretContainerArgs(),
							},
							//Env: []v1.EnvVar{},
							VolumeMounts: []v1.VolumeMount{
								b.restoreSourcePVCContainerVolumeMount(),
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever, // Only "Never" or "OnFailure" are accepted in Kubernetes Jobs
				},
			},
		},
	}
}

func (b *APIManagerRestore) ZyncResyncDomainsJob() *batchv1.Job {
	if b.options.APIManagerRestorePVCOptions == nil {
		return nil
	}

	jobName, err := helper.UIDBasedJobName("resync-domains", b.options.APIManagerRestoreUID)
	if err != nil {
		panic(err)
	}

	var completions int32 = 1
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: b.options.Namespace,
		},
		Spec: batchv1.JobSpec{
			Completions: &completions,
			// TODO BackoffLimit field controls how many times the job is retried
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						v1.Container{
							Name:  "job",
							Image: b.options.OCCLIImageURL,
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"-c",
								"-e",
								b.zyncResyncDomainsContainerArgs(),
							},
							//Env: []v1.EnvVar{},
						},
					},
					RestartPolicy: v1.RestartPolicyNever, // Only "Never" or "OnFailure" are accepted in Kubernetes Jobs
				},
			},
		},
	}
}

func (b *APIManagerRestore) SystemStoragePVC(restoreInfo *RuntimeAPIManagerRestoreInfo) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.SystemFileStoragePVCName,
			Namespace: b.options.Namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: restoreInfo.PVCStorageClass,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					// We hardcode the size due to in APIManager is hardcoded to 100Mi. If in
					// the future this changes we should change it here too or update the
					// logic here
					v1.ResourceStorage: resource.MustParse("100Mi"),
				},
			},
		},
	}
}

func (b *APIManagerRestore) SecretToShareName() string {
	return fmt.Sprintf("%s-serialized-apimanager", b.options.APIManagerRestoreName)
}

func (b *APIManagerRestore) restoreSystemFilestoragePVCContainerArgs() string {
	// We could use rsync -av but the problem is that
	// rsync tries to change the attributes of the destination directory
	// and cannot do it due to the volume permissions
	//rsync: chgrp "/system-filestorage-pvc/." failed: Operation not permitted (1)
	// ./
	// An alternative using the "-a" flag could be using * but then could be problems
	// like reaching glob extension limit, or not taking into account dot files in the
	// main directory etc...
	// So it seems there's no "perfect" solution
	return fmt.Sprintf(`
	BASEPATH='%s';
	SYSTEM_FILESTORAGE_PVC_DIR='%s'
  PVC_BACKUP_FILESTORAGE_SUBDIR="${BASEPATH}/system-filestorage-pvc";
	rsync -rlv ${PVC_BACKUP_FILESTORAGE_SUBDIR}/* ${SYSTEM_FILESTORAGE_PVC_DIR}/;
`,
		RestorePVCMountPath,
		SystemFileStoragePVCMountPath,
	)
}

func (b *APIManagerRestore) createAPIManagerSharedSecretContainerArgs() string {
	return fmt.Sprintf(`
  BASEPATH='%s';
  APIMANAGER_BACKUP_SUBDDIR="${BASEPATH}/apimanager";
  SECRET_TO_SHARE='%s';
  APIMANAGER_BACKUP_FILENAME="%s";
  oc create secret generic ${SECRET_TO_SHARE} --from-file=${APIMANAGER_BACKUP_SUBDDIR}/${APIMANAGER_BACKUP_FILENAME};
`,
		RestorePVCMountPath,
		b.SecretToShareName(),
		backup.APIManagerSerializedBackupFileName,
	)
}

func (b *APIManagerRestore) restoreSecretsAndConfigMapsContainerArgs() string {
	return fmt.Sprintf(`
	SECRETS='%s';
	CONFIGMAPS='%s';
	BASEPATH='%s';
	for i in $(echo -n $SECRETS); do
		res=$(oc get secret ${i} --ignore-not-found=true)
		if [ -z "${res}" ]; then
			oc create -f ${BASEPATH}/secrets/${i}.json
		else
			echo "Secret ${i} already exists. Skipping restore of the secret"
		fi
	done;
	
	for i in $(echo -n $CONFIGMAPS); do
		res=$(oc get configmap ${i} --ignore-not-found=true);
		if [ -z "${res}" ]; then
			oc create -f ${BASEPATH}/configmaps/${i}.json;
		else
			echo "ConfigMap '${i}' already exists. Skipping restore of the ConfigMap";
		fi
	done;
`,
		strings.Join(helper.SortedMapStringStringValues(secretsToRestore), " "),
		strings.Join(helper.SortedMapStringStringValues(configMapsToRestore), " "),
		RestorePVCMountPath,
	)
}

func (b *APIManagerRestore) zyncResyncDomainsContainerArgs() string {
	return `
	dcname="system-sidekiq"
	dcpods=$(oc get pods --ignore-not-found=true -l deploymentconfig=${dcname} --no-headers=true -o custom-columns=:metadata.name)
	if [ -z "${dcpods}" ]; then
		echo "No pods found for Deployment ${dcname}"
		exit 1
	fi
	podname=$(echo -n $dcpods | awk '{print $1}')
	oc exec ${podname} bash -- -c "bundle exec rake zync:resync:domains"
`
}
