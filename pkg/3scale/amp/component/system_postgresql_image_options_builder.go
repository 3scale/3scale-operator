package component

import "fmt"

type SystemPostgreSQLImageOptionsBuilder struct {
	options SystemPostgreSQLImageOptions
}

func (b *SystemPostgreSQLImageOptionsBuilder) AmpRelease(release string) {
	b.options.ampRelease = release
}

func (b *SystemPostgreSQLImageOptionsBuilder) AppLabel(appLabel string) {
	b.options.appLabel = appLabel
}

func (b *SystemPostgreSQLImageOptionsBuilder) Image(image string) {
	b.options.image = image
}

func (b *SystemPostgreSQLImageOptionsBuilder) InsecureImportPolicy(insecure bool) {
	b.options.insecureImportPolicy = insecure
}

func (b *SystemPostgreSQLImageOptionsBuilder) Build() (*SystemPostgreSQLImageOptions, error) {
	err := b.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	b.setNonRequiredOptions()

	return &b.options, nil
}

func (b *SystemPostgreSQLImageOptionsBuilder) setRequiredOptions() error {
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

func (b *SystemPostgreSQLImageOptionsBuilder) setNonRequiredOptions() {
}
