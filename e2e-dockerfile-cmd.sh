#!/bin/sh

go fmt ./...
go vet ./...
test testbin/setup-envtest.sh
source testbin/setup-envtest.sh
fetch_envtest_tools testbin
setup_envtest_env testbin
./3scale-operator-test-harness-controllers-capabilities.test -test.v
./3scale-operator-test-harness-controllers-apps.test -test.v -ginkgo.v -ginkgo.progress -ginkgo.debug


