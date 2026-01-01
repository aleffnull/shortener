package config

import (
	"os"
	"path"
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
				HTTPS: &HTTPSConfiguration{
					Enabled:         true,
					CertificateFile: "server.crt",
					KeyFile:         "server.key",
				},
				CPUProfile:    "profiles/cpu.pprof",
				MemoryProfile: "profiles/memory.pprof",
				ConfigFile:    "config.json",
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

func TestConfiguration_parseFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		want       *Configuration
		wantError  bool
		hookBefore func() (*Configuration, *Configuration)
	}{
		{
			name: "WHEN no file THEN defaut",
			want: &Configuration{
				FileStore:     &FileStoreConfiguration{},
				DatabaseStore: &DatabaseStoreConfiguration{},
				HTTPS:         &HTTPSConfiguration{},
			},
			hookBefore: func() (*Configuration, *Configuration) {
				return &Configuration{}, &Configuration{}
			},
		},
		{
			name:      "WHEN not existing file THEN error",
			wantError: true,
			hookBefore: func() (*Configuration, *Configuration) {
				return &Configuration{
					ConfigFile: "not_existing_file.json",
				}, &Configuration{}
			},
		},
		{
			name:      "WHEN unmarshal error THEN error",
			wantError: true,
			hookBefore: func() (*Configuration, *Configuration) {
				filePath := path.Join(t.TempDir(), "config.json")
				file, err := os.Create(filePath)
				require.NoError(t, err)
				defer file.Close()

				return &Configuration{
					ConfigFile: filePath,
				}, &Configuration{}
			},
		},
		{
			name: "WHEN no errors THEN ok",
			want: &Configuration{
				ServerAddress: "http://localhost",
				FileStore:     &FileStoreConfiguration{},
				DatabaseStore: &DatabaseStoreConfiguration{},
				HTTPS:         &HTTPSConfiguration{},
			},
			hookBefore: func() (*Configuration, *Configuration) {
				filePath := path.Join(t.TempDir(), "config.json")
				json := []byte(`{"server_address": "http://localhost"}`)
				require.NoError(t, os.WriteFile(filePath, json, 0644))

				return &Configuration{
					ConfigFile: filePath,
				}, &Configuration{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			envConfig, flagConfig := tt.hookBefore()

			// Act.
			configuration, err := parseFile(envConfig, flagConfig)

			// Assert.
			require.Equal(t, tt.want, configuration)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfiguration_getStringValue(t *testing.T) {
	t.Parallel()

	type args struct {
		envValue  string
		flagValue string
		fileValue string
	}

	tests := []struct {
		name string
		args *args
		want string
	}{
		{
			name: "WHEN all THEN env value",
			args: &args{
				envValue:  "env",
				flagValue: "flag",
				fileValue: "file",
			},
			want: "env",
		},
		{
			name: "WHEN flag and file THEN flag value",
			args: &args{
				flagValue: "flag",
				fileValue: "file",
			},
			want: "flag",
		},
		{
			name: "WHEN file THEN file value",
			args: &args{
				fileValue: "file",
			},
			want: "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act-assert.
			got := getStringValue(tt.args.envValue, tt.args.flagValue, tt.args.fileValue)
			require.Equal(t, tt.want, got)
		})
	}
}
