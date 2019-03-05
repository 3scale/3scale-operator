package component

import "fmt"

type BuildConfigsOptionsBuilder struct {
	options BuildConfigsOptions
}

func (bcs *BuildConfigsOptionsBuilder) AppLabel(appLabel string) {
	bcs.options.appLabel = appLabel
}

func (bcs *BuildConfigsOptionsBuilder) GitRef(gitRef string) {
	bcs.options.gitRef = gitRef
}

func (bcs *BuildConfigsOptionsBuilder) Build() (*BuildConfigsOptions, error) {
	if bcs.options.appLabel == "" {
		return nil, fmt.Errorf("no AppLabel has been provided")
	}
	if bcs.options.gitRef == "" {
		return nil, fmt.Errorf("no Git Ref. has been provided")
	}
	return &bcs.options, nil
}
