package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryStoreConfiguration_String(t *testing.T) {
	t.Parallel()

	// Arrange.
	configuration := MemoryStoreConfiguration{
		KeyStoreConfiguration: KeyStoreConfiguration{
			KeyLength:        1,
			KeyMaxLength:     2,
			KeyMaxIterations: 3,
		},
	}

	// Act.
	str := configuration.String()

	// Assert.
	require.NotEmpty(t, str)
}

func TestMemoryStoreConfiguration_defaultMemoryStoreConfiguration(t *testing.T) {
	t.Parallel()

	// Act
	configuration := defaultMemoryStoreConfiguration()

	// Assert.
	require.Greater(t, configuration.KeyLength, 0)
	require.Greater(t, configuration.KeyMaxLength, 0)
	require.Greater(t, configuration.KeyMaxIterations, 0)
}
