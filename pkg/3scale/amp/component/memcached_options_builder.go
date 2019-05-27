package component

import "fmt"

type MemcachedOptions struct {
	// memcachedRequiredOptions
	appLabel string
}

type MemcachedOptionsBuilder struct {
	options MemcachedOptions
}

func (m *MemcachedOptionsBuilder) AppLabel(appLabel string) {
	m.options.appLabel = appLabel
}

func (m *MemcachedOptionsBuilder) Build() (*MemcachedOptions, error) {
	err := m.setRequiredOptions()
	if err != nil {
		return nil, err
	}

	m.setNonRequiredOptions()

	return &m.options, nil
}

func (m *MemcachedOptionsBuilder) setRequiredOptions() error {
	if m.options.appLabel == "" {
		return fmt.Errorf("no AppLabel has been provided")
	}

	return nil
}

func (m *MemcachedOptionsBuilder) setNonRequiredOptions() {

}
