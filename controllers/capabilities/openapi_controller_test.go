package controllers

import (
	"context"
	"github.com/3scale/3scale-operator/pkg/helper"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	capabilitiesv1beta1 "github.com/3scale/3scale-operator/apis/capabilities/v1beta1"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestOpenAPIReconciler_validateOASExtensions(t *testing.T) {
	productExtensionPath := field.NewPath("x-3scale-product")
	metricsExtensionPath := productExtensionPath.Child("metrics")
	policiesExtensionPath := productExtensionPath.Child("policies")

	type fields struct {
		BaseReconciler *reconcilers.BaseReconciler
	}
	type args struct {
		openapiObj *openapi3.T
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{
			name: "valid OAS",
			fields: fields{
				BaseReconciler: getBaseReconciler(getOpenAPICR(), getValidOpenAPISecret()),
			},
			args: args{
				openapiObj: getOpenAPIObj(getValidOpenAPISecret()),
			},
			want: nil,
		},
		{
			name: "unextended OAS",
			fields: fields{
				BaseReconciler: getBaseReconciler(getOpenAPICR(), getUnextendedOpenAPISecret()),
			},
			args: args{
				openapiObj: getOpenAPIObj(getUnextendedOpenAPISecret()),
			},
			want: nil,
		},
		{
			name: "bad OAS",
			fields: fields{
				BaseReconciler: getBaseReconciler(getOpenAPICR(), getBadExtendedOpenAPISecret()),
			},
			args: args{
				openapiObj: getOpenAPIObj(getBadExtendedOpenAPISecret()),
			},
			want: &helper.SpecFieldError{
				ErrorType: helper.InvalidError,
				FieldErrorList: field.ErrorList{
					{
						Type:     field.ErrorTypeRequired,
						Field:    metricsExtensionPath.String(),
						BadValue: "",
						Detail:   "metric metric01 is missing a friendlyName",
					},
					{
						Type:     field.ErrorTypeRequired,
						Field:    metricsExtensionPath.String(),
						BadValue: "",
						Detail:   "metric metric01 is missing a unit",
					},
					{
						Type:     field.ErrorTypeRequired,
						Field:    policiesExtensionPath.String(),
						BadValue: "",
						Detail:   "one or more policies are missing a name",
					},
					{
						Type:     field.ErrorTypeRequired,
						Field:    policiesExtensionPath.String(),
						BadValue: "",
						Detail:   "policy  is missing a version",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &OpenAPIReconciler{
				BaseReconciler: tt.fields.BaseReconciler,
			}
			got := r.validateOASExtensions(tt.args.openapiObj)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateOASExtensions() got = %v, want %v", got, tt.want)
			}
		})
	}
}
func getTestBaseReconciler(objects ...runtime.Object) (baseReconciler *reconcilers.BaseReconciler) {
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	err := capabilitiesv1beta1.AddToScheme(s)
	if err != nil {
		panic(err)
	}
	err = appsv1alpha1.AddToScheme(s)
	if err != nil {
		panic(err)
	}
	err = corev1.AddToScheme(s)
	if err != nil {
		panic(err)
	}

	var clientObjects []client.Object
	if objects != nil {
		for _, o := range objects {
			co, ok := o.(client.Object)
			if ok {
				clientObjects = append(clientObjects, co)
			}
		}
	}

	// Create a fake client to mock API calls.
	cl := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(objects...).WithStatusSubresource(clientObjects...).Build()
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)
	baseReconciler = reconcilers.NewBaseReconciler(context.TODO(), cl, s, cl, getOpenAPITestLogger(), clientset.Discovery(), recorder)
	return baseReconciler
}

func getOpenAPITestLogger() logr.Logger {
	return logf.Log.WithName("openapi_product_reconciler_test")
}

func getOpenAPICR() *capabilitiesv1beta1.OpenAPI {
	return &capabilitiesv1beta1.OpenAPI{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testCR",
			Namespace: "testNamespace",
		},
		Spec: capabilitiesv1beta1.OpenAPISpec{
			OpenAPIRef: capabilitiesv1beta1.OpenAPIRefSpec{
				SecretRef: &corev1.ObjectReference{
					Name:      "testOpenAPISecret",
					Namespace: "testNamespace",
				},
			},
		},
	}
}

func getOpenAPIObj(openAPISecret *corev1.Secret) *openapi3.T {
	openAPIReconciler := OpenAPIReconciler{
		BaseReconciler: getTestBaseReconciler(openAPISecret),
	}

	openapiObj, err := openAPIReconciler.readOpenAPI(getOpenAPICR())
	if err != nil {
		panic(err)
	}
	return openapiObj
}

func getValidOpenAPISecret() *corev1.Secret {
	secretData := `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      x-3scale-operation:
        mappingRule:
          metricMethodRef: "metric01"
          increment: 2
          last: true
      responses:
        '200':
          description: A paged array of pets
x-3scale-product:
  metrics:
    metric01:
      friendlyName: "My Metric 01"
      unit: "hits"
      description: "This is a custom metric"
  policies:
    - name: "myPolicy1"
      version: "0.1"
      enabled: true
      configuration:
        http_proxy: http://example.com
    - name: "myPolicy2"
      version: "2.0"
      enabled: true
      configurationRef:
        name: "my-config-policy-secret"
        namespace: "testNamespace"
  applicationPlans:
    plan01:
      name: "My Plan 01"
      appsRequireApproval: false
      trialPeriod: 1
      setupFee: "1.00"
      costMonth: "1.00"
      pricingRules:
        - from: 1
          to: 100
          pricePerUnit: "1.00"
          metricMethodRef:
            systemName: "metric01"
      limits:
        - period: "week"
          value: 100
          metricMethodRef:
            systemName: "hits"
            backend: "Swagger_Petstore"
`
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testOpenAPISecret",
			Namespace: "testNamespace",
		},
		Data: map[string][]byte{
			"oas": []byte(secretData),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

func getUnextendedOpenAPISecret() *corev1.Secret {
	secretData := `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      responses:
        '200':
          description: A paged array of pets
`
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testOpenAPISecret",
			Namespace: "testNamespace",
		},
		Data: map[string][]byte{
			"oas": []byte(secretData),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

func getBadExtendedOpenAPISecret() *corev1.Secret {
	secretData := `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      responses:
        '200':
          description: A paged array of pets
x-3scale-product:
  metrics:
    metric01:
      description: "This is a custom metric"
  policies:
    - enabled: true
      configuration:
        http_proxy: http://example.com
`
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testOpenAPISecret",
			Namespace: "testNamespace",
		},
		Data: map[string][]byte{
			"oas": []byte(secretData),
		},
		Type: corev1.SecretTypeOpaque,
	}
}
