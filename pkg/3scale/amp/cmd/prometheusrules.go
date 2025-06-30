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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/prometheusrules"
)

var prometheusRulesNamespace string

// Compatibility with Openshift <4.9
var compatPre49 bool

// templateCmd represents the template command
var prometheusRulesCmd = &cobra.Command{
	Use:   getPrometheusRulesUsage(),
	Short: getPrometheusRulesShortDescription(),
	Long:  getPrometheusRulesLongDescription(),
	Args:  cobra.ExactArgs(1),
	RunE:  runPrometheusRulesCommand,
}

func getPrometheusRulesUsage() string {
	return "prometheusrules <rules-name>"
}

func getPrometheusRulesShortDescription() string {
	return "generate prometheus rules serialized resource"
}

func getPrometheusRulesLongDescription() string {
	return "generate prometheus rules serialized resource"
}

func runPrometheusRulesCommand(cmd *cobra.Command, args []string) error {
	// Check factory items names do not conflict
	factoryMap := map[string]prometheusrules.PrometheusRuleFactory{}
	for _, factoryBuilder := range prometheusrules.PrometheusRuleFactories {
		factory := factoryBuilder()
		if _, ok := factoryMap[factory.Type()]; ok {
			return fmt.Errorf("PrometheusRule factory %s already exists", factory.Type())
		}
		factoryMap[factory.Type()] = factory
	}

	prName := args[0]
	factory, ok := factoryMap[prName]
	if !ok {
		return fmt.Errorf("factory %s not found", prName)
	}

	prometheusRulesObj := factory.PrometheusRule(compatPre49, prometheusRulesNamespace)

	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil,
		json.SerializerOptions{Yaml: true, Pretty: true, Strict: true})
	return serializer.Encode(prometheusRulesObj, os.Stdout)
}

func init() {
	prometheusRulesCmd.Flags().StringVar(&prometheusRulesNamespace, "namespace", "", "Namespace to be used when generating the prometheus rules")
	prometheusRulesCmd.Flags().BoolVar(&compatPre49, "compat", false, "Generate rules compatible with Openshift releases prior to 4.9")
	err := prometheusRulesCmd.MarkFlagRequired("namespace")
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(prometheusRulesCmd)
}
