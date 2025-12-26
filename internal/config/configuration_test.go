package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfiguration_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		configuration *Configuration
	}{
		{
			name:          "WHEN no fields THEN ok",
			configuration: &Configuration{},
		},
		{
			name: "WHEN all fields THEN ok",
			configuration: &Configuration{
				ServerAddress: "localhost:8080",
				BaseURL:       "http://localhost:8080",
				AuditFile:     "audit.jsonl",
				AuditURL:      "http://auditor/audit",
				MemoryStore: &MemoryStoreConfiguration{
					KeyStoreConfiguration: KeyStoreConfiguration{
						KeyLength:        1,
						KeyMaxLength:     2,
						KeyMaxIterations: 3,
					},
				},
				FileStore: &FileStoreConfiguration{
					FilePath: "foo.jsonl",
				},
				DatabaseStore: &DatabaseStoreConfiguration{
					KeyStoreConfiguration: KeyStoreConfiguration{
						KeyLength:        4,
						KeyMaxLength:     5,
						KeyMaxIterations: 6,
					},
					DataSourceName: "somedb://somewhere",
				},
				CPUProfile:    "profiles/cpu.pprof",
				MemoryProfile: "profiles/memory.pprof",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			str := tt.configuration.String()

			// Assert.
			require.NotEmpty(t, str)
		})
	}
}

func TestConfiguration_GetConfiguration(t *testing.T) {
	t.Parallel()

	// Act.
	configuration, err := GetConfiguration()

	// Assert.
	require.NotNil(t, configuration)
	require.NoError(t, err)
}
