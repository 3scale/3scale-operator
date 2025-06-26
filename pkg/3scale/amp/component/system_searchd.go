package component

import (
	"context"

	"github.com/3scale/3scale-operator/pkg/reconcilers"
	k8sappsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SystemSearchdDeploymentName = "system-searchd"
	SystemSearchdPVCName        = "system-searchd-manticore"
	SystemSearchdServiceName    = "system-searchd"
	SystemSearchdDBVolumeName   = "system-searchd-database"

	// 3scale 2.14 -> 2.15 (manticore)
	SystemSearchdReindexJobName = "system-searchd-manticore-reindex"
)

type SystemSearchd struct {
	Options *SystemSearchdOptions
}

func NewSystemSearchd(options *SystemSearchdOptions) *SystemSearchd {
	return &SystemSearchd{Options: options}
}

func (s *SystemSearchd) Service() *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSearchdServiceName,
			Labels: s.Options.Labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name:       "searchd",
					Protocol:   v1.ProtocolTCP,
					Port:       9306,
					TargetPort: intstr.FromInt32(9306),
				},
			},
			Selector: map[string]string{reconcilers.DeploymentLabelSelector: "system-searchd"},
		},
	}
}

func (s *SystemSearchd) Deployment(ctx context.Context, k8sclient client.Client, operatorNamespace string, containerImage string) (*k8sappsv1.Deployment, error) {
	var searchdReplicas int32 = 1
	watchedSecretAnnotations, err := ComputeWatchedSecretAnnotations(ctx, k8sclient, SystemSearchdDeploymentName, operatorNamespace, s)
	if err != nil {
		return nil, err
	}

	return &k8sappsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: reconcilers.DeploymentAPIVersion, Kind: reconcilers.DeploymentKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSearchdDeploymentName,
			Labels: s.Options.Labels,
		},
		Spec: k8sappsv1.DeploymentSpec{
			Replicas: &searchdReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					reconcilers.DeploymentLabelSelector: SystemSearchdDeploymentName,
				},
			},
			Strategy: k8sappsv1.DeploymentStrategy{
				Type: k8sappsv1.RecreateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      s.Options.PodTemplateLabels,
					Annotations: s.searchdPodAnnotations(watchedSecretAnnotations),
				},
				Spec: v1.PodSpec{
					InitContainers:     s.searchdInit(containerImage),
					Affinity:           s.Options.Affinity,
					Tolerations:        s.Options.Tolerations,
					ServiceAccountName: "amp",
					Volumes:            s.searchdVolume(),
					Containers: []v1.Container{
						{
							Name:            SystemSearchdDeploymentName,
							Image:           containerImage,
							ImagePullPolicy: v1.PullIfNotPresent,
							VolumeMounts:    s.searchDVolumeMounts(),
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt32(9306),
									},
								},
								InitialDelaySeconds: 60,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.FromInt32(9306),
									},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      10,
								PeriodSeconds:       30,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							Resources: s.Options.ContainerResourceRequirements,
						},
					},
					PriorityClassName:         s.Options.PriorityClassName,
					TopologySpreadConstraints: s.Options.TopologySpreadConstraints,
				},
			},
		},
	}, nil
}

func (s *SystemSearchd) PVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSearchdPVCName,
			Labels: s.Options.Labels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: s.Options.PVCOptions.StorageClass,
			VolumeName:       s.Options.PVCOptions.VolumeName,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.PersistentVolumeAccessMode("ReadWriteOnce"),
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: s.Options.PVCOptions.StorageRequests,
				},
			},
		},
	}
}

// ReindexingJob returns the job to run manticore reindexing command. This will be removed for 2.16.
// 3scale 2.14 -> 2.15 (manticore)
func (s *SystemSearchd) ReindexingJob(containerImage string, system *System) *batchv1.Job {
	var completions int32 = 1

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   SystemSearchdReindexJobName,
			Labels: s.Options.Labels,
		},
		Spec: batchv1.JobSpec{
			Completions: &completions,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: s.searchdInit(containerImage),
					Containers: []v1.Container{
						{
							Name:            SystemSearchdReindexJobName,
							Image:           containerImage,
							Args:            []string{"bash", "-c", "bundle exec rake searchd:optimal_index"},
							Env:             system.buildSystemBaseEnv(),
							Resources:       s.Options.ContainerResourceRequirements,
							ImagePullPolicy: v1.PullIfNotPresent,
							VolumeMounts:    s.searchdManticoreVolumeMounts(),
						},
					},
					Volumes:            s.searchdJobVolume(),
					RestartPolicy:      v1.RestartPolicyNever,
					ServiceAccountName: "amp",
					PriorityClassName:  s.Options.PriorityClassName,
				},
			},
		},
	}
}

func (s *SystemSearchd) searchdInit(containerImage string) []v1.Container {
	if s.Options.SearchdDbTLSEnabled {
		return []v1.Container{
			{
				Name:  "set-permissions",
				Image: containerImage, // Minimal image for chmod
				Command: []string{
					"sh",
					"-c",
					"cp /tls/* /writable-tls/ && chmod 0600 /writable-tls/*",
				},
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "tls-secret",
						MountPath: "/tls",
						ReadOnly:  true,
					},
					{
						Name:      "writable-tls",
						MountPath: "/writable-tls",
						ReadOnly:  false, // Writable emptyDir volume
					},
				},
			},
		}
	} else {
		return []v1.Container{}
	}
}

func (s *SystemSearchd) searchDVolumeMounts() []v1.VolumeMount {
	if s.Options.SearchdDbTLSEnabled {
		return []v1.VolumeMount{
			{
				ReadOnly:  false,
				Name:      SystemSearchdDBVolumeName,
				MountPath: "/var/lib/searchd",
			},
			{
				Name:      "writable-tls", // Reuse the same volume in the main container if needed
				MountPath: "/tls",
				ReadOnly:  true,
			},
		}
	} else {
		return []v1.VolumeMount{
			{
				ReadOnly:  false,
				Name:      SystemSearchdDBVolumeName,
				MountPath: "/var/lib/searchd",
			},
		}
	}
}

func (s *SystemSearchd) searchdManticoreVolumeMounts() []v1.VolumeMount {
	if s.Options.SearchdDbTLSEnabled {
		return []v1.VolumeMount{
			{
				Name:      "writable-tls", // Reuse the same volume in the main container if needed
				MountPath: "/tls",
				ReadOnly:  true,
			},
		}
	} else {
		return []v1.VolumeMount{}
	}
}

func (s *SystemSearchd) searchdJobVolume() []v1.Volume {
	if s.Options.SearchdDbTLSEnabled {
		return []v1.Volume{
			{
				Name: "tls-secret",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: SystemSecretSystemDatabaseSecretName, // Name of the secret containing the TLS certs
						Items: []v1.KeyToPath{
							{
								Key:  SystemSecretSslCa,
								Path: "ca.crt", // Map the secret key to the ca.crt file in the container
							},
							{
								Key:  SystemSecretSslCert,
								Path: "tls.crt", // Map the secret key to the tls.crt file in the container
							},
							{
								Key:  SystemSecretSslKey,
								Path: "tls.key", // Map the secret key to the tls.key file in the container
							},
						},
					},
				},
			},
			{
				Name: "writable-tls",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		}
	} else {
		return []v1.Volume{}
	}
}

func (s *SystemSearchd) searchdVolume() []v1.Volume {
	if s.Options.SearchdDbTLSEnabled {
		return []v1.Volume{
			{
				Name: SystemSearchdDBVolumeName,
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: SystemSearchdPVCName,
						ReadOnly:  false,
					},
				},
			},
			{
				Name: "tls-secret",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: SystemSecretSystemDatabaseSecretName, // Name of the secret containing the TLS certs
						Items: []v1.KeyToPath{
							{
								Key:  SystemSecretSslCa,
								Path: "ca.crt", // Map the secret key to the ca.crt file in the container
							},
							{
								Key:  SystemSecretSslCert,
								Path: "tls.crt", // Map the secret key to the tls.crt file in the container
							},
							{
								Key:  SystemSecretSslKey,
								Path: "tls.key", // Map the secret key to the tls.key file in the container
							},
						},
					},
				},
			},
			{
				Name: "writable-tls",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		}
	} else {
		return []v1.Volume{
			{
				Name: SystemSearchdDBVolumeName,
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: SystemSearchdPVCName,
						ReadOnly:  false,
					},
				},
			},
		}
	}
}

func (s *SystemSearchd) searchdPodAnnotations(watchedSecretAnnotations map[string]string) map[string]string {
	annotations := s.Options.PodTemplateAnnotations

	if annotations == nil {
		annotations = make(map[string]string)
	}
	for key, val := range watchedSecretAnnotations {
		annotations[key] = val
	}

	for key, val := range s.Options.PodTemplateAnnotations {
		annotations[key] = val
	}

	return annotations
}
