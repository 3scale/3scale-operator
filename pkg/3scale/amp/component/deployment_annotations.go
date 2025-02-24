package component

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/3scale/3scale-operator/pkg/helper"

	"github.com/go-logr/logr"
	k8sappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func ComputeWatchedSecretAnnotations(ctx context.Context, client k8sclient.Client, deploymentName, watchNS string, component interface{}) (map[string]string, error) {
	// First get the initial annotations
	uncheckedAnnotations, err := getWatchedSecretAnnotations(ctx, client, deploymentName, watchNS, component)
	if err != nil {
		return nil, err
	}

	// Then get the deployment (if it exists) to compare the existing annotations
	deployment := &k8sappsv1.Deployment{}
	deploymentKey := k8sclient.ObjectKey{
		Name:      deploymentName,
		Namespace: watchNS,
	}
	err = client.Get(ctx, deploymentKey, deployment)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	// If the deployment doesn't exist yet then just return the uncheckedAnnotations because there's nothing to compare to
	if apierrors.IsNotFound(err) {
		return uncheckedAnnotations, nil
	}

	// Next get the master hashed secret (if it exists) to compare the secret hashes
	hashedSecret := &corev1.Secret{}
	hashedSecretKey := k8sclient.ObjectKey{
		Name:      HashedSecretName,
		Namespace: watchNS,
	}
	err = client.Get(ctx, hashedSecretKey, hashedSecret)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	// If the master hashed secret doesn't exist yet then just return the uncheckedAnnotations because there's nothing to compare to
	if apierrors.IsNotFound(err) {
		return uncheckedAnnotations, nil
	}

	existingPodAnnotations := deployment.Spec.Template.Annotations

	// First check if the annotations match, if they do then there's no need to check the hash
	if !reflect.DeepEqual(existingPodAnnotations, uncheckedAnnotations) {
		reconciledAnnotations := map[string]string{}

		// Loop through the annotations to see if the secret data has actually changed
		for key, resourceVersion := range uncheckedAnnotations {
			if existingPodAnnotations[key] == "" {
				reconciledAnnotations[key] = resourceVersion // If this is the first time adding the annotation then use the new resourceVersion
			} else if existingPodAnnotations[key] != resourceVersion && HasSecretHashChanged(ctx, client, key, hashedSecret, watchNS, component) {
				reconciledAnnotations[key] = resourceVersion // Else if the resourceVersions don't match and the hash has changed then use the new resourceVersion
			} else {
				reconciledAnnotations[key] = existingPodAnnotations[key] // Otherwise keep the existing resourceVersion
			}
		}

		return reconciledAnnotations, nil
	}
	return uncheckedAnnotations, nil // No difference with existing annotations so can return uncheckedAnnotations
}

func getWatchedSecretAnnotations(ctx context.Context, client k8sclient.Client, deploymentName string, namespace string, component interface{}) (map[string]string, error) {
	annotations := map[string]string{}

	switch c := component.(type) {
	case *Apicast:
		// HTTPs Certificate Secret
		// OpenTelemetry Config Secret
		// Custom Policy Secret(s)
		// Custom Env Secret(s)

		apicast := c

		if deploymentName == ApicastProductionName {
			if apicast.Options.ProductionHTTPSCertificateSecretName != nil && *apicast.Options.ProductionHTTPSCertificateSecretName != "" {
				httpCertSecret := &corev1.Secret{}
				httpCertSecretKey := k8sclient.ObjectKey{
					Name:      *apicast.Options.ProductionHTTPSCertificateSecretName,
					Namespace: apicast.Options.Namespace,
				}
				err := client.Get(ctx, httpCertSecretKey, httpCertSecret)
				if err != nil {
					return nil, err
				}
				if helper.IsSecretWatchedBy3scale(httpCertSecret) {
					annotationKey := fmt.Sprintf("%s%s", HttpsCertSecretResverAnnotationPrefix, httpCertSecret.Name)
					annotations[annotationKey] = httpCertSecret.ResourceVersion
				}
			}

			if &apicast.Options.ProductionOpentelemetry != nil && apicast.Options.ProductionOpentelemetry.Enabled {
				if &apicast.Options.ProductionOpentelemetry.Secret != nil && apicast.Options.ProductionOpentelemetry.Secret.Name != "" {
					telemetryConfigSecret := &corev1.Secret{}
					telemetryConfigSecretKey := k8sclient.ObjectKey{
						Name:      apicast.Options.ProductionOpentelemetry.Secret.Name,
						Namespace: apicast.Options.Namespace,
					}
					err := client.Get(ctx, telemetryConfigSecretKey, telemetryConfigSecret)
					if err != nil {
						return nil, err
					}
					if helper.IsSecretWatchedBy3scale(telemetryConfigSecret) {
						annotationKey := fmt.Sprintf("%s%s", OpenTelemetrySecretResverAnnotationPrefix, telemetryConfigSecret.Name)
						annotations[annotationKey] = telemetryConfigSecret.ResourceVersion
					}

				}
			}

			for idx := range apicast.Options.ProductionCustomPolicies {
				// Secrets must exist and have the watched-by label
				// Annotation key includes the name of the secret
				if helper.IsSecretWatchedBy3scale(apicast.Options.ProductionCustomPolicies[idx].Secret) {
					annotationKey := fmt.Sprintf("%s%s", CustomPoliciesSecretResverAnnotationPrefix, apicast.Options.ProductionCustomPolicies[idx].Secret.Name)
					annotations[annotationKey] = apicast.Options.ProductionCustomPolicies[idx].Secret.ResourceVersion
				}
			}

			for idx := range apicast.Options.ProductionCustomEnvironments {
				// Secrets must exist and have the watched-by label
				// Annotation key includes the name of the secret
				if helper.IsSecretWatchedBy3scale(apicast.Options.ProductionCustomEnvironments[idx]) {
					annotationKey := fmt.Sprintf("%s%s", CustomEnvSecretResverAnnotationPrefix, apicast.Options.ProductionCustomEnvironments[idx].Name)
					annotations[annotationKey] = apicast.Options.ProductionCustomEnvironments[idx].ResourceVersion
				}
			}
		} else if deploymentName == ApicastStagingName {
			if apicast.Options.StagingHTTPSCertificateSecretName != nil && *apicast.Options.StagingHTTPSCertificateSecretName != "" {
				httpCertSecret := &corev1.Secret{}
				httpCertSecretKey := k8sclient.ObjectKey{
					Name:      *apicast.Options.StagingHTTPSCertificateSecretName,
					Namespace: apicast.Options.Namespace,
				}
				err := client.Get(ctx, httpCertSecretKey, httpCertSecret)
				if err != nil {
					return nil, err
				}
				if helper.IsSecretWatchedBy3scale(httpCertSecret) {
					annotationKey := fmt.Sprintf("%s%s", HttpsCertSecretResverAnnotationPrefix, httpCertSecret.Name)
					annotations[annotationKey] = httpCertSecret.ResourceVersion
				}
			}

			if &apicast.Options.StagingOpentelemetry != nil && apicast.Options.StagingOpentelemetry.Enabled {
				if &apicast.Options.StagingOpentelemetry.Secret != nil && apicast.Options.StagingOpentelemetry.Secret.Name != "" {
					telemetryConfigSecret := &corev1.Secret{}
					telemetryConfigSecretKey := k8sclient.ObjectKey{
						Name:      apicast.Options.StagingOpentelemetry.Secret.Name,
						Namespace: apicast.Options.Namespace,
					}
					err := client.Get(ctx, telemetryConfigSecretKey, telemetryConfigSecret)
					if err != nil {
						return nil, err
					}
					if helper.IsSecretWatchedBy3scale(telemetryConfigSecret) {
						annotationKey := fmt.Sprintf("%s%s", OpenTelemetrySecretResverAnnotationPrefix, telemetryConfigSecret.Name)
						annotations[annotationKey] = telemetryConfigSecret.ResourceVersion
					}
				}
			}

			for idx := range apicast.Options.StagingCustomPolicies {
				// Secrets must exist and have the watched-by label
				// Annotation key includes the name of the secret
				if helper.IsSecretWatchedBy3scale(apicast.Options.StagingCustomPolicies[idx].Secret) {
					annotationKey := fmt.Sprintf("%s%s", CustomPoliciesSecretResverAnnotationPrefix, apicast.Options.StagingCustomPolicies[idx].Secret.Name)
					annotations[annotationKey] = apicast.Options.StagingCustomPolicies[idx].Secret.ResourceVersion
				}
			}

			for idx := range apicast.Options.StagingCustomEnvironments {
				// Secrets must exist and have the watched-by label
				// Annotation key includes the name of the secret
				if helper.IsSecretWatchedBy3scale(apicast.Options.StagingCustomEnvironments[idx]) {
					annotationKey := fmt.Sprintf("%s%s", CustomEnvSecretResverAnnotationPrefix, apicast.Options.StagingCustomEnvironments[idx].Name)
					annotations[annotationKey] = apicast.Options.StagingCustomEnvironments[idx].ResourceVersion
				}
			}
		}
	case *System:
		system := c
		systemDatabase := &corev1.Secret{}
		systemDatabaseSecretKey := k8sclient.ObjectKey{
			Name:      SystemSecretSystemDatabaseSecretName,
			Namespace: system.Options.Namespace,
		}
		err := client.Get(ctx, systemDatabaseSecretKey, systemDatabase)
		if err != nil {
			return nil, err
		}
		if helper.IsSecretWatchedBy3scale(systemDatabase) {
			annotationKey := fmt.Sprintf("%s%s", SystemDatabaseSecretResverAnnotationPrefix, systemDatabase.Name)
			annotations[annotationKey] = systemDatabase.ResourceVersion
		}

	case *SystemSearchd:
		systemDatabase := &corev1.Secret{}
		systemDatabaseSecretKey := k8sclient.ObjectKey{
			Name:      SystemSecretSystemDatabaseSecretName,
			Namespace: namespace,
		}
		err := client.Get(ctx, systemDatabaseSecretKey, systemDatabase)
		if err != nil {
			return nil, err
		}
		if helper.IsSecretWatchedBy3scale(systemDatabase) {
			annotationKey := fmt.Sprintf("%s%s", SystemDatabaseSecretResverAnnotationPrefix, systemDatabase.Name)
			annotations[annotationKey] = systemDatabase.ResourceVersion
		}

	case *Zync:
		zync := c
		zyncSecret := &corev1.Secret{}
		zyncSecretKey := k8sclient.ObjectKey{
			Name:      ZyncSecretName,
			Namespace: zync.Options.Namespace,
		}
		err := client.Get(ctx, zyncSecretKey, zyncSecret)
		if err != nil {
			fmt.Printf("failed to find zync secret, yet to be create %s", err)
			return nil, nil
		}
		if helper.IsSecretWatchedBy3scale(zyncSecret) {
			annotationKey := fmt.Sprintf("%s%s", ZyncSecretResverAnnotationPrefix, zyncSecret.Name)
			annotations[annotationKey] = zyncSecret.ResourceVersion
		}

	default:
		return nil, fmt.Errorf("unrecognized component %s is not supported", deploymentName)
	}

	return annotations, nil
}

func HasSecretHashChanged(ctx context.Context, client k8sclient.Client, deploymentAnnotation string, hashedSecret *corev1.Secret, watchNS string, component interface{}) bool {
	logger, _ := logr.FromContext(ctx)

	secretToCheck := &corev1.Secret{}
	secretToCheckKey := k8sclient.ObjectKey{
		Namespace: watchNS,
	}

	// Assign the name of the secret to check based on the component and secret type
	switch c := component.(type) {
	case *Apicast:
		switch {
		case strings.HasPrefix(deploymentAnnotation, HttpsCertSecretResverAnnotationPrefix):
			secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, HttpsCertSecretResverAnnotationPrefix)
		case strings.HasPrefix(deploymentAnnotation, OpenTelemetrySecretResverAnnotationPrefix):
			secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, OpenTelemetrySecretResverAnnotationPrefix)
		case strings.HasPrefix(deploymentAnnotation, CustomEnvSecretResverAnnotationPrefix):
			secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, CustomEnvSecretResverAnnotationPrefix)
		case strings.HasPrefix(deploymentAnnotation, CustomPoliciesSecretResverAnnotationPrefix):
			secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, CustomPoliciesSecretResverAnnotationPrefix)
		default:
			return false
		}
	case *System:
		switch {
		case strings.HasPrefix(deploymentAnnotation, SystemDatabaseSecretResverAnnotationPrefix):
			secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, SystemDatabaseSecretResverAnnotationPrefix)
		default:
			return false
		}
	case *SystemSearchd:
		switch {
		case strings.HasPrefix(deploymentAnnotation, SystemDatabaseSecretResverAnnotationPrefix):
			secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, SystemDatabaseSecretResverAnnotationPrefix)
		default:
			return false
		}
	case *Zync:
		switch {
		case strings.HasPrefix(deploymentAnnotation, ZyncSecretResverAnnotationPrefix):
			secretToCheckKey.Name = strings.TrimPrefix(deploymentAnnotation, ZyncSecretResverAnnotationPrefix)
		default:
			return false
		}
	default:
		logger.Info(fmt.Sprintf("unrecognized component %s is not supported", c))
		return false
	}

	// Get latest version of the secret to check
	err := client.Get(ctx, secretToCheckKey, secretToCheck)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to get secret %s", secretToCheckKey.Name))
		return false
	}

	// Compare the hash of the latest version of the secret's data to the reference in the hashed secret
	if HashSecret(secretToCheck.Data) != helper.GetSecretStringDataFromData(hashedSecret.Data)[secretToCheckKey.Name] {
		logger.V(1).Info(fmt.Sprintf("%s secret .data has changed - updating the resourceVersion in deployment's annotation", secretToCheckKey.Name))
		return true
	}

	logger.V(1).Info(fmt.Sprintf("%s secret .data has not changed since last checked", secretToCheckKey.Name))
	return false
}
