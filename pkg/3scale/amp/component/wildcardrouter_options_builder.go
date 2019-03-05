package component

import "fmt"

type WildcardRouterOptionsBuilder struct {
	options WildcardRouterOptions
}

func (wr *WildcardRouterOptionsBuilder) AppLabel(appLabel string) {
	wr.options.appLabel = appLabel
}

func (wr *WildcardRouterOptionsBuilder) WildcardDomain(wildcardDomain string) {
	wr.options.wildcardDomain = wildcardDomain
}

func (wr *WildcardRouterOptionsBuilder) WildcardPolicy(wildcardPolicy string) {
	wr.options.wildcardPolicy = wildcardPolicy
}

func (wr *WildcardRouterOptionsBuilder) Build() (*WildcardRouterOptions, error) {
	err := wr.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	wr.setNonRequiredOptions()

	return &wr.options, nil
}

func (wr *WildcardRouterOptionsBuilder) setRequiredOptions() error {
	if wr.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if wr.options.wildcardDomain == "" {
		return fmt.Errorf("no Wildcard Domain has been provided")
	}
	if wr.options.wildcardPolicy == "" {
		return fmt.Errorf("no Wildcard Policy has been provided")
	}

	return nil
}

func (wr *WildcardRouterOptionsBuilder) setNonRequiredOptions() {

}
