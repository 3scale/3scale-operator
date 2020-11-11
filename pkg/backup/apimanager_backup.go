package backup

import (
	"fmt"
	"strings"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const BackupPVCMountPath = "/backup"
const SystemFileStoragePVCMountPath = "/system-filestorage-pvc"
const APIManagerSerializedBackupFileName = "apimanager-backup.json"

var secretsToBackup map[string]string = map[string]string{
	"SystemSMTP":          "system-smtp",
	"SystemSeed":          "system-seed",
	"SystemDatabase":      "system-database",
	"BackendInternalAPI":  "backend-internal-api",
	"BackendRedis":        "backend-redis",
	"BackendListener":     "backend-listener",
	"SystemEventsHook":    "system-events-hook",
	"SystemApp":           "system-app",
	"SystemRecaptcha":     "system-recaptcha",
	"SystemRedis":         "system-redis",
	"SystemMemcached":     "system-memcache", // TODO should we backup/restore this one?
	"Zync":                "zync",
	"SystemMasterAPIcast": "system-master-apicast",
}

var configMapsToBackup map[string]string = map[string]string{
	"SystemEnvironment":  "system-environment",
	"APIcastEnvironment": "apicast-environment",
}

type APIManagerBackup struct {
	options *APIManagerBackupOptions
}

func NewAPIManagerBackup(options *APIManagerBackupOptions) *APIManagerBackup {
	return &APIManagerBackup{
		options: options,
	}
}

func (b *APIManagerBackup) Validate() error {
	return nil
}

func (b *APIManagerBackup) APIManager() *appsv1alpha1.APIManager {
	return b.options.APIManager
}

func (b *APIManagerBackup) BackupDestinationPVC() *v1.PersistentVolumeClaim {
	if b.options.APIManagerBackupPVCOptions == nil {
		return nil
	}

	volName := ""
	if b.options.APIManagerBackupPVCOptions.BackupDestinationPVC.VolumeName != nil {
		volName = *b.options.APIManagerBackupPVCOptions.BackupDestinationPVC.VolumeName
	}

	res := &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.options.APIManagerBackupName,
			Namespace: b.options.Namespace,
			Annotations: map[string]string{
				"apiManagerName":       b.APIManager().Name,
				"apiManagerUID":        string(b.APIManager().GetUID()),
				"apiManagerBackupName": b.options.APIManagerBackupName,
				"apiManagerBackupUID":  string(b.options.APIManagerBackupUID),
			},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			StorageClassName: b.options.APIManagerBackupPVCOptions.BackupDestinationPVC.StorageClass,
			VolumeName:       volName, // TODO maybe make directly the pvc volume name option a string instead of a pointer to string? if we do that we would not be able to differentiate between not passed and passed but empty
		},
	}

	if b.options.APIManagerBackupPVCOptions.BackupDestinationPVC.StorageRequests != nil {
		res.Spec.Resources.Requests = v1.ResourceList{
			v1.ResourceStorage: *b.options.APIManagerBackupPVCOptions.BackupDestinationPVC.StorageRequests,
		}
	}

	return res
}

func (b *APIManagerBackup) BackupSecretsAndConfigMapsToPVCJob() *batchv1.Job {
	if b.options.APIManagerBackupPVCOptions == nil {
		return nil
	}

	jobName, err := helper.UIDBasedJobName("backup-cfgmaps-secrets", b.options.APIManagerBackupUID)
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
						b.pvcBackupDestinationPodVolume(),
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "backup-cfgmaps-secrets",
							Image: b.options.OCCLIImageURL,
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"-c",
								"-e",
								b.backupSecretsAndConfigMapsContainerArgs(),
							},
							//Env: []v1.EnvVar{},
							VolumeMounts: []v1.VolumeMount{
								b.pvcBackupDestinationContainerVolumeMount(),
							},
						},
					},
					ServiceAccountName: "3scale-operator",     // TODO create our own SA, Role and RoleBinding to do just what we need
					RestartPolicy:      v1.RestartPolicyNever, // Only "Never" or "OnFailure" are accepted in Kubernetes Jobs
				},
			},
		},
	}
}

func (b *APIManagerBackup) BackupAPIManagerCustomResourceToPVCJob() *batchv1.Job {
	if b.options.APIManagerBackupPVCOptions == nil {
		return nil
	}

	jobName, err := helper.UIDBasedJobName("backup-apimanager-cr", b.options.APIManagerBackupUID)
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
						b.pvcBackupDestinationPodVolume(),
					},
					Containers: []v1.Container{
						v1.Container{
							Name:  "backup-apimanager-cr",
							Image: b.options.OCCLIImageURL,
							Command: []string{
								"/bin/bash",
							},
							Args: []string{
								"-c",
								"-e",
								b.backupAPIManagerCustomResourceContainerArgs(),
							},
							//Env: []v1.EnvVar{},
							VolumeMounts: []v1.VolumeMount{
								b.pvcBackupDestinationContainerVolumeMount(),
							},
						},
					},
					ServiceAccountName: "3scale-operator",     // TODO create our own SA, Role and RoleBinding to do just what we need
					RestartPolicy:      v1.RestartPolicyNever, // Only "Never" or "OnFailure" are accepted in Kubernetes Jobs
				},
			},
		},
	}
}

func (b *APIManagerBackup) BackupSystemFileStoragePVCToPVCJob() *batchv1.Job {
	if b.options.APIManagerBackupPVCOptions == nil {
		return nil
	}

	jobName, err := helper.UIDBasedJobName("backup-system-fs-pvc", b.options.APIManagerBackupUID)
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
						b.pvcBackupDestinationPodVolume(),
						b.systemFileStoragePodVolume(),
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
								b.backupSystemFilestoragePVCContainerArgs(),
							},
							//Env: []v1.EnvVar{},
							VolumeMounts: []v1.VolumeMount{
								b.pvcBackupDestinationContainerVolumeMount(),
								b.systemFileStorageContainerVolumeMount(),
							},
						},
					},
					ServiceAccountName: "3scale-operator",     // TODO create our own SA, Role and RoleBinding to do just what we need
					RestartPolicy:      v1.RestartPolicyNever, // Only "Never" or "OnFailure" are accepted in Kubernetes Jobs
				},
			},
		},
	}
}

func (b *APIManagerBackup) systemFileStoragePodVolume() v1.Volume {
	return v1.Volume{
		Name: "system-storage",
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: "system-storage",
				ReadOnly:  true,
			},
		},
	}
}

func (b *APIManagerBackup) systemFileStorageContainerVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      "system-storage",
		MountPath: SystemFileStoragePVCMountPath,
	}
}

func (b *APIManagerBackup) pvcBackupDestinationPodVolume() v1.Volume {
	return v1.Volume{
		Name: b.BackupDestinationPVC().Name,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: b.BackupDestinationPVC().Name,
			},
		},
	}
}

func (b *APIManagerBackup) pvcBackupDestinationContainerVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      b.BackupDestinationPVC().Name,
		MountPath: BackupPVCMountPath,
	}
}

func (b *APIManagerBackup) backupSecretsAndConfigMapsContainerArgs() string {
	pythonCleanupSubscriptContent := b.pythonCleanupK8sObjectScript()
	return fmt.Sprintf(`
SECRETS="%s";
CONFIGMAPS="%s";
BASEPATH="%s";
PYTHON_CLEANUP_SUBSCRIPT="%s"
mkdir -p $BASEPATH/secrets;
mkdir -p $BASEPATH/configmaps;
for i in $(echo -n $SECRETS); do oc get secret -o json $i | python -c "${PYTHON_CLEANUP_SUBSCRIPT}" > $BASEPATH/secrets/$i.json; done;
for i in $(echo -n $CONFIGMAPS); do oc get configmap -o json $i | python -c "${PYTHON_CLEANUP_SUBSCRIPT}" > $BASEPATH/configmaps/$i.json; done;
`,
		strings.Join(helper.SortedMapStringStringValues(secretsToBackup), " "),
		strings.Join(helper.SortedMapStringStringValues(configMapsToBackup), " "),
		BackupPVCMountPath,
		pythonCleanupSubscriptContent,
	)
}

func (b *APIManagerBackup) backupAPIManagerCustomResourceContainerArgs() string {
	pythonCleanupSubscriptContent := b.pythonCleanupK8sObjectScript()
	return fmt.Sprintf(`
BASEPATH="%s";
PYTHON_CLEANUP_SUBSCRIPT="%s"
APIMANAGER_NAME="%s"
APIMANAGER_BACKUP_FILENAME="%s"
APIMANAGER_SUBDIR="$BASEPATH/apimanager";
mkdir -p ${APIMANAGER_SUBDIR};
oc get apimanager -o json ${APIMANAGER_NAME} | python -c "${PYTHON_CLEANUP_SUBSCRIPT}" > ${APIMANAGER_SUBDIR}/${APIMANAGER_BACKUP_FILENAME};
`,
		BackupPVCMountPath,
		pythonCleanupSubscriptContent,
		b.options.APIManagerName,
		APIManagerSerializedBackupFileName,
	)
}

func (b *APIManagerBackup) pythonCleanupK8sObjectScript() string {
	return `
import sys, json

parsed=json.load(sys.stdin)
if 'status' in parsed:
  del parsed['status']

metadataAttrsToDelete = ['ownerReferences', 'selfLink', 'uid', 'resourceVersion', 'creationTimestamp', 'namespace', 'clusterName', 'generation']
metadata=parsed['metadata']
for metadataAttr in metadataAttrsToDelete:
  if metadataAttr in metadata:
    del metadata[metadataAttr]

print(json.dumps(parsed, indent=4, sort_keys=True))
`
}

func (b *APIManagerBackup) backupSystemFilestoragePVCContainerArgs() string {
	return fmt.Sprintf(`
BASEPATH='%s';
SYSTEM_FILESTORAGE_PVC_DIR='%s'
PVC_BACKUP_FILESTORAGE_SUBDIR="${BASEPATH}/system-filestorage-pvc";
mkdir -p ${PVC_BACKUP_FILESTORAGE_SUBDIR};
rsync -av ${SYSTEM_FILESTORAGE_PVC_DIR}/ ${PVC_BACKUP_FILESTORAGE_SUBDIR};
`,
		BackupPVCMountPath,
		SystemFileStoragePVCMountPath,
	)
}
