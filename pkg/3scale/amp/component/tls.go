package component

type TLSConfig struct {
	Enabled       bool
	Mode          string
	CACertificate string
	Certificate   string
	Key           string
}

// HasCA returns whether the configuration has a certificate authority or not.
func (c *TLSConfig) HasCA() bool {
	return c.CACertificate != ""
}

// HasCertAuth returns whether the configuration has certificate authentication or not.
func (c *TLSConfig) HasCertAuth() bool {
	return (c.Certificate != "" && c.Key != "")
}
