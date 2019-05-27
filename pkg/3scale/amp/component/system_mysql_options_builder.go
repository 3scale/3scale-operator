package component

import "fmt"

type SystemMysqlOptions struct {
	// mysqlRequiredOptions
	appLabel     string
	databaseName string
	user         string
	password     string
	rootPassword string
	databaseURL  string
}

type SystemMysqlOptionsBuilder struct {
	options SystemMysqlOptions
}

func (m *SystemMysqlOptionsBuilder) AppLabel(appLabel string) {
	m.options.appLabel = appLabel
}

func (m *SystemMysqlOptionsBuilder) DatabaseName(databaseName string) {
	m.options.databaseName = databaseName
}

func (m *SystemMysqlOptionsBuilder) User(user string) {
	m.options.user = user
}

func (m *SystemMysqlOptionsBuilder) Password(password string) {
	m.options.password = password
}

func (m *SystemMysqlOptionsBuilder) RootPassword(rootPassword string) {
	m.options.rootPassword = rootPassword
}

func (m *SystemMysqlOptionsBuilder) DatabaseURL(url string) {
	m.options.databaseURL = url
}

func (m *SystemMysqlOptionsBuilder) Build() (*SystemMysqlOptions, error) {
	err := m.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	m.setNonRequiredOptions()

	return &m.options, nil
}

func (m *SystemMysqlOptionsBuilder) setRequiredOptions() error {
	if m.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if m.options.databaseName == "" {
		return fmt.Errorf("no Database Name has been provided")
	}
	if m.options.user == "" {
		return fmt.Errorf("no User has been provided")
	}
	if m.options.password == "" {
		return fmt.Errorf("no Password has been provided")
	}
	if m.options.rootPassword == "" {
		return fmt.Errorf("no Root Password has been provided")
	}
	if m.options.databaseURL == "" {
		return fmt.Errorf("no Database URL has been provided")
	}

	return nil
}

func (m *SystemMysqlOptionsBuilder) setNonRequiredOptions() {
}
