package envtest

import (
	"context"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/3scale/3scale-operator/pkg/reconcilers"
)

func baseDeployment(name, ns string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			// Strategy omitted — K8s defaults to RollingUpdate 25%/25%;
			// normalizeDeploymentDefaults fills in the same values on desired.
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "main",
						Image: "nginx:1.25",
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.FromInt32(8080),
									// Scheme omitted — K8s defaults to "HTTP"
								},
							},
							// Integer fields omitted — K8s defaults them to 1/10/1/3
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.FromInt32(8080),
								},
							},
						},
					}},
				},
			},
		},
	}
}

var _ = Describe("DeploymentMutator normalization", func() {
	ctx := context.Background()

	createAndGet := func(desired *appsv1.Deployment) *appsv1.Deployment {
		Expect(k8sClient.Create(ctx, desired.DeepCopy())).To(Succeed())
		existing := &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}, existing)).To(Succeed())
		return existing
	}

	assertNoUpdate := func(existing, desired *appsv1.Deployment, mutators ...reconcilers.DMutateFn) {
		changed, err := reconcilers.DeploymentMutator(mutators...)(existing, desired)
		Expect(err).NotTo(HaveOccurred())
		Expect(changed).To(BeFalse())
	}

	Context("normalizeDeploymentDefaults", func() {
		It("does not detect change for Strategy type default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("strat-type", testNamespace)
				// Strategy entirely omitted — K8s defaults Type=RollingUpdate, MaxUnavailable=25%, MaxSurge=25%
				d.Spec.Strategy = appsv1.DeploymentStrategy{}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentStrategyMutator)
		})

		It("does not detect change for MaxUnavailable default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("strat-maxunavail", testNamespace)
				maxSurge := intstr.FromString("25%")
				d.Spec.Strategy = appsv1.DeploymentStrategy{
					Type: appsv1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &appsv1.RollingUpdateDeployment{
						MaxSurge: &maxSurge,
						// MaxUnavailable omitted — K8s defaults to 25%
					},
				}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentStrategyMutator)
		})

		It("does not detect change for MaxSurge default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("strat-maxsurge", testNamespace)
				maxUnavailable := intstr.FromString("25%")
				d.Spec.Strategy = appsv1.DeploymentStrategy{
					Type: appsv1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &appsv1.RollingUpdateDeployment{
						MaxUnavailable: &maxUnavailable,
						// MaxSurge omitted — K8s defaults to 25%
					},
				}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentStrategyMutator)
		})
	})

	Context("normalizeProbeDefaults", func() {
		It("does not detect change for HTTPGet scheme default", func() {
			desired := baseDeployment("probe-scheme", testNamespace)
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentProbesMutator)
		})

		It("does not detect change for TimeoutSeconds default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("probe-timeout", testNamespace)
				d.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
					ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8080)}},
					PeriodSeconds:    10,
					SuccessThreshold: 1,
					FailureThreshold: 3,
					// TimeoutSeconds omitted — K8s defaults to 1
				}
				d.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
					ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8080)}},
					PeriodSeconds:    10,
					SuccessThreshold: 1,
					FailureThreshold: 3,
				}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentProbesMutator)
		})

		It("does not detect change for PeriodSeconds default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("probe-period", testNamespace)
				d.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
					ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8080)}},
					TimeoutSeconds:   1,
					SuccessThreshold: 1,
					FailureThreshold: 3,
					// PeriodSeconds omitted — K8s defaults to 10
				}
				d.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
					ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8080)}},
					TimeoutSeconds:   1,
					SuccessThreshold: 1,
					FailureThreshold: 3,
				}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentProbesMutator)
		})

		It("does not detect change for SuccessThreshold default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("probe-success", testNamespace)
				d.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
					ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8080)}},
					TimeoutSeconds:   1,
					PeriodSeconds:    10,
					FailureThreshold: 3,
					// SuccessThreshold omitted — K8s defaults to 1
				}
				d.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
					ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8080)}},
					TimeoutSeconds:   1,
					PeriodSeconds:    10,
					FailureThreshold: 3,
				}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentProbesMutator)
		})

		It("does not detect change for FailureThreshold default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("probe-failure", testNamespace)
				d.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
					ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8080)}},
					TimeoutSeconds:   1,
					PeriodSeconds:    10,
					SuccessThreshold: 1,
					// FailureThreshold omitted — K8s defaults to 3
				}
				d.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
					ProbeHandler:     corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt32(8080)}},
					TimeoutSeconds:   1,
					PeriodSeconds:    10,
					SuccessThreshold: 1,
				}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentProbesMutator)
		})
	})

	Context("normalizeContainerDefaults", func() {
		It("does not detect change for TerminationMessagePath default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("ctr-termpath", testNamespace)
				d.Spec.Template.Spec.InitContainers = []corev1.Container{{
					Name:                     "init",
					Image:                    "busybox:1.36",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					ImagePullPolicy:          corev1.PullIfNotPresent,
					// TerminationMessagePath omitted — K8s defaults to "/dev/termination-log"
				}}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentPodInitContainerMutator)
		})

		It("does not detect change for TerminationMessagePolicy default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("ctr-termpolicy", testNamespace)
				d.Spec.Template.Spec.InitContainers = []corev1.Container{{
					Name:                   "init",
					Image:                  "busybox:1.36",
					TerminationMessagePath: "/dev/termination-log",
					ImagePullPolicy:        corev1.PullIfNotPresent,
					// TerminationMessagePolicy omitted — K8s defaults to "File"
				}}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentPodInitContainerMutator)
		})

		It("does not detect change for ImagePullPolicy default", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("ctr-pullpolicy", testNamespace)
				d.Spec.Template.Spec.InitContainers = []corev1.Container{{
					Name:                     "init",
					Image:                    "busybox:1.36",
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					// ImagePullPolicy omitted — K8s infers "IfNotPresent" from tagged image
				}}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentPodInitContainerMutator)
		})

		It("does not detect change for empty VolumeMounts normalised to nil", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("ctr-volumemounts", testNamespace)
				d.Spec.Template.Spec.InitContainers = []corev1.Container{{
					Name:                     "init",
					Image:                    "busybox:1.36",
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
					ImagePullPolicy:          corev1.PullIfNotPresent,
					VolumeMounts:             []corev1.VolumeMount{}, // empty slice — K8s stores as nil
				}}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentPodInitContainerMutator)
		})
	})

	Context("normalizePodSpecDefaults", func() {
		It("does not detect change for RestartPolicy default", func() {
			desired := baseDeployment("ps-restartpol", testNamespace)
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, func(desired, existing *appsv1.Deployment) (bool, error) {
				if existing.Spec.Template.Spec.RestartPolicy != desired.Spec.Template.Spec.RestartPolicy {
					existing.Spec.Template.Spec.RestartPolicy = desired.Spec.Template.Spec.RestartPolicy
					return true, nil
				}
				return false, nil
			})
		})

		It("does not detect change for DNSPolicy default", func() {
			desired := baseDeployment("ps-dnspolicy", testNamespace)
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, func(desired, existing *appsv1.Deployment) (bool, error) {
				if existing.Spec.Template.Spec.DNSPolicy != desired.Spec.Template.Spec.DNSPolicy {
					existing.Spec.Template.Spec.DNSPolicy = desired.Spec.Template.Spec.DNSPolicy
					return true, nil
				}
				return false, nil
			})
		})

		It("does not detect change for SecurityContext default", func() {
			desired := baseDeployment("ps-secctx", testNamespace)
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, func(desired, existing *appsv1.Deployment) (bool, error) {
				if !reflect.DeepEqual(existing.Spec.Template.Spec.SecurityContext, desired.Spec.Template.Spec.SecurityContext) {
					existing.Spec.Template.Spec.SecurityContext = desired.Spec.Template.Spec.SecurityContext
					return true, nil
				}
				return false, nil
			})
		})

		It("does not detect change for TerminationGracePeriodSeconds default", func() {
			desired := baseDeployment("ps-termperiod", testNamespace)
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, func(desired, existing *appsv1.Deployment) (bool, error) {
				if !reflect.DeepEqual(existing.Spec.Template.Spec.TerminationGracePeriodSeconds, desired.Spec.Template.Spec.TerminationGracePeriodSeconds) {
					existing.Spec.Template.Spec.TerminationGracePeriodSeconds = desired.Spec.Template.Spec.TerminationGracePeriodSeconds
					return true, nil
				}
				return false, nil
			})
		})

		It("does not detect change for SchedulerName default", func() {
			desired := baseDeployment("ps-scheduler", testNamespace)
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, func(desired, existing *appsv1.Deployment) (bool, error) {
				if existing.Spec.Template.Spec.SchedulerName != desired.Spec.Template.Spec.SchedulerName {
					existing.Spec.Template.Spec.SchedulerName = desired.Spec.Template.Spec.SchedulerName
					return true, nil
				}
				return false, nil
			})
		})
	})

	Context("normalizeVolumeDefaults", func() {
		It("does not detect change for ConfigMap DefaultMode", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("vol-configmap", testNamespace)
				d.Spec.Template.Spec.Volumes = []corev1.Volume{{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "my-config"},
							// DefaultMode omitted — K8s defaults to 0644 (420)
						},
					},
				}}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentVolumesMutator)
		})

		It("does not detect change for Secret DefaultMode", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("vol-secret", testNamespace)
				d.Spec.Template.Spec.Volumes = []corev1.Volume{{
					Name: "secret",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "my-secret",
							// DefaultMode omitted — K8s defaults to 0644 (420)
						},
					},
				}}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentVolumesMutator)
		})

		It("does not detect change for DownwardAPI DefaultMode", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("vol-downapi", testNamespace)
				d.Spec.Template.Spec.Volumes = []corev1.Volume{{
					Name: "downapi",
					VolumeSource: corev1.VolumeSource{
						DownwardAPI: &corev1.DownwardAPIVolumeSource{
							Items: []corev1.DownwardAPIVolumeFile{{
								Path: "labels",
								FieldRef: &corev1.ObjectFieldSelector{
									APIVersion: "v1", // K8s defaults to "v1" if empty
									FieldPath:  "metadata.labels",
								},
							}},
							// DefaultMode omitted — K8s defaults to 0644 (420)
						},
					},
				}}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentVolumesMutator)
		})

		It("does not detect change for Projected DefaultMode", func() {
			desired := func() *appsv1.Deployment {
				d := baseDeployment("vol-projected", testNamespace)
				d.Spec.Template.Spec.Volumes = []corev1.Volume{{
					Name: "projected",
					VolumeSource: corev1.VolumeSource{
						Projected: &corev1.ProjectedVolumeSource{
							Sources: []corev1.VolumeProjection{{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{Name: "my-config"},
								},
							}},
							// DefaultMode omitted — K8s defaults to 0644 (420)
						},
					},
				}}
				return d
			}()
			existing := createAndGet(desired)
			assertNoUpdate(existing, desired, reconcilers.DeploymentVolumesMutator)
		})
	})
})
