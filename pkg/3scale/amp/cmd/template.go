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

	amptemplate "github.com/3scale/3scale-operator/pkg/3scale/amp/template"
	"github.com/spf13/cobra"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
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

// runCommand is the function that will be executed by Command
// because it is passed as an argument on the Command struct variable
// defined in this file.
// The signature of the runCommand function (excluding its name)
// is the one needed by the Cobra library
func runCommand(cmd *cobra.Command, args []string) {
	templateName := args[0]
	componentOptions := []string{}

	template := amptemplate.NewTemplate(templateName, componentOptions)

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
