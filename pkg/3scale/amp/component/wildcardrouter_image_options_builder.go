package component

import "fmt"

type WildcardRouterImageOptionsBuilder struct {
	options WildcardRouterImageOptions
}

func (b *WildcardRouterImageOptionsBuilder) AmpRelease(release string) {
	b.options.ampRelease = release
}

func (b *WildcardRouterImageOptionsBuilder) AppLabel(appLabel string) {
	b.options.appLabel = appLabel
}

func (b *WildcardRouterImageOptionsBuilder) Image(image string) {
	b.options.image = image
}

func (b *WildcardRouterImageOptionsBuilder) InsecureImportPolicy(insecure bool) {
	b.options.insecureImportPolicy = insecure
}

func (b *WildcardRouterImageOptionsBuilder) Build() (*WildcardRouterImageOptions, error) {
	err := b.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	b.setNonRequiredOptions()

	return &b.options, nil
}

func (b *WildcardRouterImageOptionsBuilder) setRequiredOptions() error {
	if b.options.ampRelease == "" {
		return fmt.Errorf("no AmpRelease has been provided")
	}
	if b.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if b.options.image == "" {
		return fmt.Errorf("no Image has been provided")
	}

	return nil
}

func (b *WildcardRouterImageOptionsBuilder) setNonRequiredOptions() {
}
