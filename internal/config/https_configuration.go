package config

import "fmt"

type HTTPSConfiguration struct {
	Enabled         bool   `env:"ENABLE_HTTPS"`
	CertificateFile string `env:"CERTIFICATE_FILE" validate:"required_if=Enabled true,omitempty,filepath"`
	KeyFile         string `env:"KEY_FILE" validate:"required_if=Enabled true,omitempty,filepath"`
}

func (c *HTTPSConfiguration) String() string {
	return fmt.Sprintf(
		"&HTTPSConfiguration{Enabled:%v CertificateFile:'%v' KeyFile:'%v'}",
		c.Enabled,
		c.CertificateFile,
		c.KeyFile,
	)
}
