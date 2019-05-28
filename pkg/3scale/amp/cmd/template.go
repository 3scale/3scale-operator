// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"
	"strings"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	templatev1 "github.com/openshift/api/template/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	componentsSeparator = "+"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   getUsage(),
	Short: getShortDescription(),
	Long:  getLongDescription(),
	Args:  cobra.ExactArgs(1),
	Run:   runCommand,
}

func getUsage() string {
	usageStr := "template <components>"
	return usageStr
}

func getShortDescription() string {
	shortDescription := "A brief description of your command"
	return shortDescription
}

func getLongDescription() string {
	longDescription := `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`
	return longDescription
}

func getAvailableOrderedComponents() []component.ComponentType {
	return []component.ComponentType{
		component.AmpTemplateType,
		component.AmpS3TemplateType,
		component.AmpEvalTemplateType,
		component.AmpHATemplateType,
		component.AmpEvalS3TemplateType,
		component.AmpPostgreSQLTemplateType,

		component.AmpImagesType,
		component.RedisType,
		component.BackendType,
		component.MySQLType,
		component.MemcachedType,
		component.SystemType,
		component.ZyncType,
		component.ApicastType,
		component.WildcardRouterType,
		component.S3Type,
		component.ProductizedType,
		component.EvaluationType,
		component.HighAvailabilityType,
	}
}

func isAnAvailableComponent(inputComponent string, availableComponents []component.ComponentType) bool {
	for _, component := range availableComponents {
		if inputComponent == string(component) {
			return true
		}
	}
	return false
}

// runCommand is the function that will be executed by Command
// because it is passed as an argument on the Command struct variable
// defined in this file.
// The signature of the runCommand function (excluding its name)
// is the one needed by the Cobra library
func runCommand(cmd *cobra.Command, args []string) {
	inputComponents := parseComponentNames(&args[0])
	availableOrderedComponents := getAvailableOrderedComponents()
	template := generateTemplate()
	componentOptions := []string{}
	componentObjects := []component.Component{}

	for component, _ := range inputComponents {
		if !isAnAvailableComponent(component, availableOrderedComponents) {
			panic("invalid component")
		}
	}
	for _, element := range availableOrderedComponents {
		if _, ok := inputComponents[string(element)]; ok {
			res := component.NewComponent(string(element), componentOptions)
			res.AssembleIntoTemplate(template, componentObjects)
		}
	}

	for _, element := range availableOrderedComponents {
		if _, ok := inputComponents[string(element)]; ok {
			res := component.NewComponent(string(element), componentOptions)
			res.PostProcess(template, componentObjects)
		}
	}

	// for _, cmpntn := range components {
	// 	componentObjects := []component.Component{}
	// 	res := component.NewComponent(cmpntn, componentOptions)
	// 	res.AssembleIntoTemplate(template, componentObjects)
	// 	yamlSerializer := kubejson.NewYAMLSerializer(kubejson.DefaultMetaFactory, nil, nil)
	// 	err := yamlSerializer.Encode(template, os.Stdout)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	serializedResult, err := runtime.DefaultUnstructuredConverter.ToUnstructured(template)
	if err != nil {
		panic(err)
	}

	// TODO the code in the function below modifies the results to be able to use
	// double brace expansion in some fields, sets them
	// and serializes the contents to YAML. This code
	// is not reliable enough yet but it works for the fields
	// being set
	addDoubleBraceExpansionFieldsToResult(serializedResult)

	// Print the results in YAML format. Cannot use the NewYAMLSerializer from the
	// kubernetes apimachinery library because the methods to serialize
	// require a kubernetes object, which is incompatible with having
	// double braces expansion
	ec := yaml.NewEncoder(os.Stdout)
	ec.Encode(serializedResult)
	if err != nil {
		panic(err)
	}
}

func addDoubleBraceExpansionFieldsToResult(serializedResult map[string]interface{}) {
	for _, intobj := range serializedResult["objects"].([]interface{}) {
		obj := intobj.(map[string]interface{})
		if obj["kind"] == "ImageStream" { // Case of setting the importPolicy.insecure field in the '${AMP_RELEASE}' tag for all ImageStreams
			objspec := obj["spec"].(map[string]interface{})
			for _, intobjtag := range objspec["tags"].([]interface{}) {
				objtag := intobjtag.(map[string]interface{})
				objtagname := objtag["name"].(string)
				if objtagname == "${AMP_RELEASE}" {
					if _, ok := objtag["importPolicy"]; ok {
						importPolicyFields := objtag["importPolicy"].(map[string]interface{})
						importPolicyFields["insecure"] = "${{IMAGESTREAM_TAG_IMPORT_INSECURE}}"
					}
				}
			}
		} else if obj["kind"] == "PersistentVolumeClaim" { // Case of setting the storageClass field in the 'system-storage' PersistentVolumeClaim
			objmetadata := obj["metadata"].(map[string]interface{})
			objname := objmetadata["name"].(string)
			if objname == "system-storage" {
				objspec := obj["spec"].(map[string]interface{})
				objspec["storageClassName"] = "${{RWX_STORAGE_CLASS}}"
			}
		}
	}
}

func parseFlags() {

}

func parseComponentNames(componentsStr *string) map[string]int {
	result := make(map[string]int)
	components := strings.Split(*componentsStr, componentsSeparator)
	for idx, component := range components {
		result[component] = idx
	}
	return result
}

func init() {
	rootCmd.AddCommand(templateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// templateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// templateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func generateTemplate() *templatev1.Template {
	template := templatev1.Template{
		TypeMeta:   buildTemplateType(),
		ObjectMeta: buildTemplateMeta(),
		Message:    buildTemplateMessage(),
		Parameters: buildTemplateParameters(),
		Objects:    buildObjects(),
	}
	return &template
}

// The 'metadata' part of a template
func buildTemplateMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:        "3scale-api-management",
		Annotations: buildTemplateMetaAnnotations(),
	}
	return meta
}

func buildTemplateMetaAnnotations() map[string]string {
	annotations := map[string]string{
		"openshift.io/display-name":          "3scale API Management",
		"openshift.io/provider-display-name": "Red Hat, Inc.",
		"iconClass":                          "icon-3scale",
		"description":                        "3scale API Management main system",
		"tags":                               "integration, api management, 3scale", //TODO maybe refactor the tag values
	}
	return annotations
}

func buildTemplateType() metav1.TypeMeta {
	typeMeta := metav1.TypeMeta{
		Kind:       "Template",
		APIVersion: "template.openshift.io/v1",
	}
	return typeMeta
}

func buildTemplateMessage() string {
	return "Login on https://${TENANT_NAME}-admin.${WILDCARD_DOMAIN} as ${ADMIN_USERNAME}/${ADMIN_PASSWORD}"
}

func buildTemplateParameters() []templatev1.Parameter {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "AMP_RELEASE",
			Description: "AMP release tag.",
			Required:    true,
			Value:       "2.5.0",
		},
		templatev1.Parameter{
			Name:        "APP_LABEL",
			Description: "Used for object app labels",
			Required:    true,
			Value:       "3scale-api-management",
		},
		templatev1.Parameter{
			Name:        "TENANT_NAME",
			Description: "Tenant name under the root that Admin UI will be available with -admin suffix.",
			Required:    true,
			Value:       "3scale",
		},
		templatev1.Parameter{
			Name:        "RWX_STORAGE_CLASS",
			Description: "The Storage Class to be used by ReadWriteMany PVCs",
			Required:    false,
			Value:       "null",
			//Value: "null", //TODO this is incorrect. The value that should be used at the end is null and not "null", however template parameters only accept strings.
		},
	}
	return parameters
}

func buildObjects() []runtime.RawExtension {
	objects := []runtime.RawExtension{}
	return objects
}
