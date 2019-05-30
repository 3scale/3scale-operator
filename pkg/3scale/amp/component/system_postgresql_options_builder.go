package component

import "fmt"

type SystemPostgreSQLOptionsBuilder struct {
	options SystemPostgreSQLOptions
}

func (b *SystemPostgreSQLOptionsBuilder) AppLabel(appLabel string) {
	b.options.appLabel = appLabel
}

func (b *SystemPostgreSQLOptionsBuilder) DatabaseName(databaseName string) {
	b.options.databaseName = databaseName
}

func (b *SystemPostgreSQLOptionsBuilder) DatabaseURL(url string) {
	b.options.databaseURL = url
}

func (b *SystemPostgreSQLOptionsBuilder) User(user string) {
	b.options.user = user
}

func (b *SystemPostgreSQLOptionsBuilder) Password(password string) {
	b.options.password = password
}

func (b *SystemPostgreSQLOptionsBuilder) Build() (*SystemPostgreSQLOptions, error) {
	err := b.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	b.setNonRequiredOptions()

	return &b.options, nil
}

func (b *SystemPostgreSQLOptionsBuilder) setRequiredOptions() error {
	if b.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if b.options.databaseName == "" {
		return fmt.Errorf("no Database Name has been provided")
	}
	if b.options.user == "" {
		return fmt.Errorf("no User has been provided")
	}
	if b.options.password == "" {
		return fmt.Errorf("no Password has been provided")
	}
	if b.options.databaseURL == "" {
		return fmt.Errorf("no Database URL has been provided")
	}

	return nil
}

func (b *SystemPostgreSQLOptionsBuilder) setNonRequiredOptions() {
}
