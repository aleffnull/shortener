package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPSConfiguration_String(t *testing.T) {
	t.Parallel()

	// Arrange.
	configuration := HTTPSConfiguration{
		Enabled:         true,
		CertificateFile: "server.crt",
		KeyFile:         "server.key",
	}

	// Act.
	str := configuration.String()

	// Assert.
	require.NotEmpty(t, str)
}
