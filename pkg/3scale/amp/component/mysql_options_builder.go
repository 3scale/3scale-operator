package component

import "fmt"

type MysqlOptionsBuilder struct {
	options MysqlOptions
}

func (m *MysqlOptionsBuilder) AppLabel(appLabel string) {
	m.options.appLabel = appLabel
}

func (m *MysqlOptionsBuilder) DatabaseName(databaseName string) {
	m.options.databaseName = databaseName
}

func (m *MysqlOptionsBuilder) Image(image string) {
	m.options.image = image
}

func (m *MysqlOptionsBuilder) User(user string) {
	m.options.user = user
}

func (m *MysqlOptionsBuilder) Password(password string) {
	m.options.password = password
}

func (m *MysqlOptionsBuilder) RootPassword(rootPassword string) {
	m.options.rootPassword = rootPassword
}

func (m *MysqlOptionsBuilder) DatabaseURL(url string) {
	m.options.databaseURL = url
}

func (m *MysqlOptionsBuilder) Build() (*MysqlOptions, error) {
	err := m.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	m.setNonRequiredOptions()

	return &m.options, nil
}

func (m *MysqlOptionsBuilder) setRequiredOptions() error {
	if m.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}
	if m.options.databaseName == "" {
		return fmt.Errorf("no Database Name has been provided")
	}
	if m.options.image == "" {
		return fmt.Errorf("no Database Image has been provided")
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

func (m *MysqlOptionsBuilder) setNonRequiredOptions() {
}
