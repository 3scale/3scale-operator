package component

import "fmt"

type ZyncOptions struct {
	// zyncNonRequiredOptions
	databaseURL *string
	// zyncRequiredOptions
	appLabel            string
	authenticationToken string
	databasePassword    string
	secretKeyBase       string
}

type ZyncOptionsBuilder struct {
	options ZyncOptions
}

func (z *ZyncOptionsBuilder) AppLabel(appLabel string) {
	z.options.appLabel = appLabel
}

func (z *ZyncOptionsBuilder) AuthenticationToken(authToken string) {
	z.options.authenticationToken = authToken
}

func (z *ZyncOptionsBuilder) DatabasePassword(dbPass string) {
	z.options.databasePassword = dbPass
}

func (z *ZyncOptionsBuilder) SecretKeyBase(secretKeyBase string) {
	z.options.secretKeyBase = secretKeyBase
}

func (z *ZyncOptionsBuilder) DatabaseURL(dbURL string) {
	z.options.databaseURL = &dbURL
}

func (z *ZyncOptionsBuilder) Build() (*ZyncOptions, error) {
	err := z.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	z.setNonRequiredOptions()

	return &z.options, nil
}

func (z *ZyncOptionsBuilder) setRequiredOptions() error {
	if z.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if z.options.authenticationToken == "" {
		return fmt.Errorf("no Authentication Token has been provided")
	}
	if z.options.databasePassword == "" {
		return fmt.Errorf("no Database Password has been provided")
	}
	if z.options.secretKeyBase == "" {
		return fmt.Errorf("no Secret Key Base has been provided")
	}

	return nil
}

func (z *ZyncOptionsBuilder) setNonRequiredOptions() {
	defaultDatabaseURL := "postgresql://zync:" + z.options.databasePassword + "@zync-database:5432/zync_production"
	if z.options.databaseURL == nil {
		z.options.databaseURL = &defaultDatabaseURL
	}
}
