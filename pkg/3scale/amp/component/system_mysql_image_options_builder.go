package component

import "fmt"

type SystemMySQLImageOptionsBuilder struct {
	options SystemMySQLImageOptions
}

func (b *SystemMySQLImageOptionsBuilder) AmpRelease(release string) {
	b.options.ampRelease = release
}

func (b *SystemMySQLImageOptionsBuilder) AppLabel(appLabel string) {
	b.options.appLabel = appLabel
}

func (b *SystemMySQLImageOptionsBuilder) Image(image string) {
	b.options.image = image
}

func (b *SystemMySQLImageOptionsBuilder) InsecureImportPolicy(insecure bool) {
	b.options.insecureImportPolicy = insecure
}

func (b *SystemMySQLImageOptionsBuilder) Build() (*SystemMySQLImageOptions, error) {
	err := b.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	b.setNonRequiredOptions()

	return &b.options, nil
}

func (b *SystemMySQLImageOptionsBuilder) setRequiredOptions() error {
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

func (b *SystemMySQLImageOptionsBuilder) setNonRequiredOptions() {
}
