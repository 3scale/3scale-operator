// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"./pkg/apis/apps/v1alpha1.APIManager":       schema_pkg_apis_apps_v1alpha1_APIManager(ref),
		"./pkg/apis/apps/v1alpha1.APIManagerSpec":   schema_pkg_apis_apps_v1alpha1_APIManagerSpec(ref),
		"./pkg/apis/apps/v1alpha1.APIManagerStatus": schema_pkg_apis_apps_v1alpha1_APIManagerStatus(ref),
	}
}

func schema_pkg_apis_apps_v1alpha1_APIManager(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "APIManager is the Schema for the apimanagers API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.APIManagerSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.APIManagerStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/apps/v1alpha1.APIManagerSpec", "./pkg/apis/apps/v1alpha1.APIManagerStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_apps_v1alpha1_APIManagerSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "APIManagerSpec defines the desired state of APIManager",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"wildcardDomain": {
						SchemaProps: spec.SchemaProps{
							Description: "Wildcard domain as configured in the API Manager object",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"appLabel": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"tenantName": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"imageStreamTagImportInsecure": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"resourceRequirementsEnabled": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"apicast": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.ApicastSpec"),
						},
					},
					"backend": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.BackendSpec"),
						},
					},
					"system": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.SystemSpec"),
						},
					},
					"zync": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.ZyncSpec"),
						},
					},
					"highAvailability": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.HighAvailabilitySpec"),
						},
					},
					"podDisruptionBudget": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/apps/v1alpha1.PodDisruptionBudgetSpec"),
						},
					},
				},
				Required: []string{"wildcardDomain"},
			},
		},
		Dependencies: []string{
			"./pkg/apis/apps/v1alpha1.ApicastSpec", "./pkg/apis/apps/v1alpha1.BackendSpec", "./pkg/apis/apps/v1alpha1.HighAvailabilitySpec", "./pkg/apis/apps/v1alpha1.PodDisruptionBudgetSpec", "./pkg/apis/apps/v1alpha1.SystemSpec", "./pkg/apis/apps/v1alpha1.ZyncSpec"},
	}
}

func schema_pkg_apis_apps_v1alpha1_APIManagerStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "APIManagerStatus defines the observed state of APIManager",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"conditions": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("./pkg/apis/apps/v1alpha1.APIManagerCondition"),
									},
								},
							},
						},
					},
					"deployments": {
						SchemaProps: spec.SchemaProps{
							Description: "APIManager Deployment Configs",
							Ref:         ref("github.com/RHsyseng/operator-utils/pkg/olm.DeploymentStatus"),
						},
					},
				},
				Required: []string{"deployments"},
			},
		},
		Dependencies: []string{
			"./pkg/apis/apps/v1alpha1.APIManagerCondition", "github.com/RHsyseng/operator-utils/pkg/olm.DeploymentStatus"},
	}
}
