module github.com/3scale/3scale-operator

go 1.13

require (
	github.com/3scale/3scale-porta-go-client v0.6.0
	github.com/RHsyseng/operator-utils v0.0.0-20200506183821-e3b4a2ba9c30
	github.com/coreos/prometheus-operator v0.38.1-0.20200424145508-7e176fda06cc
	github.com/getkin/kin-openapi v0.22.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-playground/validator/v10 v10.2.0
	github.com/google/go-cmp v0.4.0
	github.com/google/uuid v1.1.1
	github.com/integr8ly/grafana-operator/v3 v3.6.0
	github.com/luci/go-render v0.0.0-20160219211803-9a04cc21af0f
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/prometheus/client_golang v1.5.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.5.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.3
)

replace github.com/openshift/api => github.com/openshift/api v0.0.0-20200527184302-a843dc3262a0 // Required until https://github.com/operator-framework/operator-lifecycle-manager/pull/1241 is resolved

// Required due to other libraries referencing v12.0.0+incompatible and without replace we can't have v0.18.6 specified
// in the require section
replace k8s.io/client-go => k8s.io/client-go v0.18.6

// security release to address CVE-2020-15257
replace github.com/containerd/containerd => github.com/containerd/containerd v1.3.9

// security release to address CVE-2020-14040
replace golang.org/x/text => golang.org/x/text v0.3.3

// security release to address CVE-2020-8912
replace github.com/aws/aws-sdk-go => github.com/aws/aws-sdk-go v1.34.0

// security release to address CVE-2020-9283. First version
// that addresses the CVE is golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
// but we replace to the most recent version that appeared on go.sum before
// this change
replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200414173820-0848c9571904
