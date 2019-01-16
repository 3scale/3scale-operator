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
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/luci/go-render/render"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"

	appsv1 "github.com/openshift/api/apps/v1"
	authorizationv1 "github.com/openshift/api/authorization/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	networkv1 "github.com/openshift/api/network/v1"
	oauthv1 "github.com/openshift/api/oauth/v1"
	projectv1 "github.com/openshift/api/project/v1"
	quotav1 "github.com/openshift/api/quota/v1"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	templatev1 "github.com/openshift/api/template/v1"
	userv1 "github.com/openshift/api/user/v1"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse [file]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		appsv1.Install(scheme.Scheme)
		authorizationv1.Install(scheme.Scheme)
		buildv1.Install(scheme.Scheme)
		imagev1.Install(scheme.Scheme)
		networkv1.Install(scheme.Scheme)
		oauthv1.Install(scheme.Scheme)
		projectv1.Install(scheme.Scheme)
		quotav1.Install(scheme.Scheme)
		routev1.Install(scheme.Scheme)
		securityv1.Install(scheme.Scheme)
		templatev1.Install(scheme.Scheme)
		userv1.Install(scheme.Scheme)

		dat, err := ioutil.ReadFile(args[0])
		check(err)

		// Create a YAML serializer.  JSON is a subset of YAML, so is supported too.
		s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme,
			scheme.Scheme)

		var template templatev1.Template

		_, _, err = s.Decode([]byte(dat), nil, &template)
		check(err)

		fmt.Print("&templatev1.Template{\n")

		fmt.Print("TypeMeta:")
		typemeta := render.Render(template.TypeMeta)
		fmt.Print(cleanup(typemeta))
		fmt.Print(",\n")

		fmt.Print("ObjectMeta:")
		objectmeta := render.Render(template.ObjectMeta)
		fmt.Print(cleanup(objectmeta))
		fmt.Print(",\n")

		fmt.Print("Message:")
		message := render.Render(template.Message)
		fmt.Print(cleanup(message))
		fmt.Print(",\n")

		fmt.Print("Parameters: []templatev1.Parameter{")
		for _, p := range template.Parameters {
			param := render.Render(p)
			fmt.Print(cleanup(param) + ",\n")
		}
		fmt.Print("},\n")

		// Some types, e.g. Template, contain RawExtensions.  If the appropriate types
		// are registered, these can be decoded in a second pass.
		fmt.Print("Objects: []runtime.RawExtension{")
		for i, o := range template.Objects {
			o.Object, _, err = s.Decode(o.Raw, nil, nil)
			check(err)
			o.Raw = nil

			template.Objects[i] = o

			object := render.Render(o.Object)
			fmt.Print("{Object: " + cleanup(object) + "},\n")
		}
		fmt.Print("},\n")
	},
}

var transformations = []map[string]string{
	{
		`(, )?\w+:\(\*[\w\d\.]+\)\(nil\)`:      ``,
		`(, )?\w+:\[\][\w\d\.]+\(nil\)`:        ``,
		`(\w+):\(\*(\w+\.\w+)\)\{`:             `$1:&$2{`, // Foo: (*v1.DeploymentConfig){ => Foo: &v1.DeploymentConfig{
		`((, )?\w+:[\w\d\[\]\.]+\((nil|"")\))`: ``,
		`((, )?\w+:"")`:                        ``,
		`^\(\*(\w+\.\w+)\){`:                   `&$1{`, // (*v1.DeploymentConfig){ => &v1.DeploymentConfig{
		`(,\s)?\w+:\w+\.\w+{}`:                 ``,
	},
	{`(,\s)?\w+:\w+\.\w+{}`: ``},
	{`(,\s)?\w+:\w+\.\w+{}`: ``},
	{`(,\s)?\w+:\w+\.\w+{}`: ``},
	{`(,\s)?\w+:\w+\.\w+{}`: ``},
	{`{,`: `{`},
	{`resource.Quantity{.+?s:"([^"]+)",.+?BinarySI"\)}`: `resource.MustParse("$1")`},
	{`\b({)\b`: "$1\n"},
	{`([^\{])(},)`: "$1,\n$2"},
	{`(,)\s(\w+:)`: "$1\n$2"},
	{`^,\n`: ``},
	{`\(\*int64\)\((\d+)\)`: `&[]int64{$1}[0]`}, // (*int64)(600) => &[]int64{$1}[0]
	{
		`(v1\.(Image|Tag))`: `image$1`,
		`(v1\.(Deploy|\w+Deployment|Lifecycle|\w+Hook))`: `apps$1`,
		`(v1\.(Route|TLS|Wildcard|InsecureEdge))`:        `route$1`,
		`(v1\.(\w+Meta|Time))`:                           `meta$1`,
		`(v1\.(Parameter|Template))`:                     `template$1`,
	},
}

var replacements = map[string]string{
	`v1.Time{Time:time.Time{wall:0, ext:0, loc:(*time.Location)(nil)}}`: `v1.Time{}`,
}

func cleanup(code string) string {
	for pattern, replacement := range replacements {
		code = strings.Replace(code, pattern, replacement, -1)
	}

	for _, transforms := range transformations {
		for pattern, replacement := range transforms {
			code = string(regexp.MustCompile(pattern).ReplaceAllString(code, replacement))
		}
	}

	return code
}

func init() {
	rootCmd.AddCommand(parseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// parseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
