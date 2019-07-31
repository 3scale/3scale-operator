package apicast

import (
	"net/url"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Apicast struct {
	namespace              string
	deploymentName         string
	serviceName            string
	replicas               int32
	appLabel               string
	serviceAccountName     string
	image                  string
	exposedHostname        string
	ownerReference         *metav1.OwnerReference
	adminPortalCredentials ApicastAdminPortalCredentials
	additionalEnvironment  *v1.SecretEnvSource
}

type ApicastAdminPortalCredentials struct {
	URL         string
	AccessToken string
}

func (a *Apicast) deploymentEnvFromSource() []v1.EnvFromSource {
	envFromSource := []v1.EnvFromSource{}
	if a.additionalEnvironment != nil {
		envFromSource = append(envFromSource, v1.EnvFromSource{
			SecretRef: a.additionalEnvironment,
		})
	}

	envFromSource = append(envFromSource, v1.EnvFromSource{
		SecretRef: &v1.SecretEnvSource{
			LocalObjectReference: v1.LocalObjectReference{
				Name: a.apicastAdminPortalSecretName(),
			},
		},
	})
	return envFromSource
}

func (a *Apicast) apicastAdminPortalSecretName() string {
	return a.deploymentName + "-admin-portal-endpoint"
}

func (a *Apicast) Deployment() *appsv1.Deployment {

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.deploymentName,
			Namespace: a.namespace,
			Labels:    a.commonLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: a.deploymentLabelSelector(),
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      a.deploymentLabelSelector(),
					Annotations: a.podAnnotations(),
				},
				Spec: v1.PodSpec{
					ServiceAccountName: a.serviceAccountName,
					Containers: []v1.Container{
						v1.Container{
							Name: a.deploymentName,
							Ports: []v1.ContainerPort{
								v1.ContainerPort{Name: "proxy", ContainerPort: 8080, Protocol: v1.ProtocolTCP},
								v1.ContainerPort{Name: "management", ContainerPort: 8090, Protocol: v1.ProtocolTCP},
								v1.ContainerPort{Name: "metrics", ContainerPort: 9421, Protocol: v1.ProtocolTCP},
							},
							Image:           a.image,
							ImagePullPolicy: v1.PullAlways, // This is different than the currently used which is IfNotPresent
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							LivenessProbe:  a.livenessProbe(),
							ReadinessProbe: a.readinessProbe(),
							// Env takes precedence with respect to EnvFrom on duplicated
							// var values
							EnvFrom: a.deploymentEnvFromSource(),
						},
					},
				},
			},
			Replicas: &a.replicas, // TODO set to nil?
		},
	}
	if a.ownerReference != nil {
		addOwnerRefToObject(deployment, *a.ownerReference)
	}
	return deployment
}

func (a *Apicast) deploymentLabelSelector() map[string]string {
	return map[string]string{
		"deployment": a.deploymentName,
	}
}

func (a *Apicast) commonLabels() map[string]string {
	return map[string]string{
		"app":                  a.appLabel,
		"threescale_component": "apicast",
	}
}

func (a *Apicast) podAnnotations() map[string]string {
	return map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9421",
	}
}

func (a *Apicast) Service() *v1.Service {
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.serviceName,
			Namespace: a.namespace,
			Labels:    a.commonLabels(),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{Name: "proxy", Port: 8080, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)},
				v1.ServicePort{Name: "management", Port: 8090, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8090)},
			},
			Selector: a.deploymentLabelSelector(),
		},
	}

	if a.ownerReference != nil {
		addOwnerRefToObject(service, *a.ownerReference)
	}

	return service
}

func (a *Apicast) livenessProbe() *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/status/live",
				Port: intstr.FromInt(8090),
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
	}
}

func (a *Apicast) readinessProbe() *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/status/ready",
				Port: intstr.FromInt(8090),
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       30,
	}
}

func (a *Apicast) Ingress() *extensions.Ingress {
	ingress := &extensions.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.deploymentName,
			Namespace: a.namespace,
			Labels:    a.commonLabels(),
		},
		Spec: extensions.IngressSpec{
			Rules: []extensions.IngressRule{
				{
					Host: a.exposedHostname,
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Backend: extensions.IngressBackend{
										ServiceName: a.deploymentName,
										ServicePort: intstr.FromString("proxy"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if a.ownerReference != nil {
		addOwnerRefToObject(ingress, *a.ownerReference)
	}

	return ingress
}

func (a *Apicast) AdminPortalEndpointSecret() (v1.Secret, error) {
	parsedURL, err := url.Parse(a.adminPortalCredentials.URL)
	if err != nil {
		return v1.Secret{}, err
	}
	parsedURL.User = url.UserPassword("", a.adminPortalCredentials.AccessToken)
	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    a.commonLabels(),
			Name:      a.apicastAdminPortalSecretName(),
			Namespace: a.namespace,
		},
		StringData: map[string]string{
			"THREESCALE_PORTAL_ENDPOINT": parsedURL.String(),
		},
	}
	if a.ownerReference != nil {
		addOwnerRefToObject(&secret, *a.ownerReference)
	}
	return secret, err
}

func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}

func (a *Apicast) ConfigMap() *v1.ConfigMap {
	configMap := &v1.ConfigMap{}
	return configMap
}
