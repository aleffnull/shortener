package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
)

type Configuration struct {
	ServerAddress string                      `env:"SERVER_ADDRESS" validate:"required,hostname_port"`
	BaseURL       string                      `env:"BASE_URL" validate:"required,url"`
	AuditFile     string                      `env:"AUDIT_FILE" validate:"omitempty,filepath"`
	AuditURL      string                      `env:"AUDIT_URL" validate:"omitempty,url"`
	MemoryStore   *MemoryStoreConfiguration   `validate:"required"`
	FileStore     *FileStoreConfiguration     `validate:"required"`
	DatabaseStore *DatabaseStoreConfiguration `validate:"required"`
	CPUProfile    string                      `env:"CPU_PROFILE" validate:"omitempty,filepath"`
	MemoryProfile string                      `env:"MEMORY_PROFILE" validate:"omitempty,filepath"`
}

func (c *Configuration) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(
		sb,
		"&Configuration{ServerAddress:%v BaseURL:%v",
		c.ServerAddress,
		c.BaseURL)

	if len(c.AuditFile) > 0 {
		fmt.Fprintf(sb, " AuditFile:%v", c.AuditFile)
	}

	if len(c.AuditURL) > 0 {
		fmt.Fprintf(sb, " AuditURL:%v", c.AuditURL)
	}

	if len(c.CPUProfile) > 0 {
		fmt.Fprintf(sb, " CPUProfile:%v", c.CPUProfile)
	}

	if len(c.MemoryProfile) > 0 {
		fmt.Fprintf(sb, " MemoryProfile:%v", c.MemoryProfile)
	}

	if c.MemoryStore == nil {
		fmt.Fprintf(sb, " MemoryStore:<nil>")
	} else {
		fmt.Fprintf(sb, " MemoryStore:%v", c.MemoryStore)
	}

	if c.FileStore == nil {
		fmt.Fprintf(sb, " FileStore:<nil>")
	} else {
		fmt.Fprintf(sb, " FileStore:%v", c.FileStore)
	}

	if c.DatabaseStore == nil {
		fmt.Fprintf(sb, " DatabaseStore:<nil>")
	} else {
		fmt.Fprintf(sb, " DatabaseStore:%v", c.DatabaseStore)
	}

	fmt.Fprintf(sb, "}")
	return sb.String()
}

func GetConfiguration() (*Configuration, error) {
	envConfig, err := parseEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	flagConfig := parseFlags()

	configuration := &Configuration{
		ServerAddress: lo.Ternary(len(envConfig.ServerAddress) > 0, envConfig.ServerAddress, flagConfig.ServerAddress),
		BaseURL:       lo.Ternary(len(envConfig.BaseURL) > 0, envConfig.BaseURL, flagConfig.BaseURL),
		AuditFile:     lo.Ternary(len(envConfig.AuditFile) > 0, envConfig.AuditFile, flagConfig.AuditFile),
		AuditURL:      lo.Ternary(len(envConfig.AuditURL) > 0, envConfig.AuditURL, flagConfig.AuditURL),
		MemoryStore:   defaultMemoryStoreConfiguration(),
		FileStore: &FileStoreConfiguration{
			FilePath: lo.Ternary(
				len(envConfig.FileStore.FilePath) > 0,
				envConfig.FileStore.FilePath,
				flagConfig.FileStore.FilePath),
		},
		DatabaseStore: NewDatabaseStoreConfiguration(lo.Ternary(
			len(envConfig.DatabaseStore.DataSourceName) > 0,
			envConfig.DatabaseStore.DataSourceName,
			flagConfig.DatabaseStore.DataSourceName,
		)),
		CPUProfile:    lo.Ternary(len(envConfig.CPUProfile) > 0, envConfig.CPUProfile, flagConfig.CPUProfile),
		MemoryProfile: lo.Ternary(len(envConfig.MemoryProfile) > 0, envConfig.MemoryProfile, flagConfig.MemoryProfile),
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(configuration)
	if err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return configuration, nil
}

func parseFlags() *Configuration {
	configuration := &Configuration{
		FileStore:     &FileStoreConfiguration{},
		DatabaseStore: &DatabaseStoreConfiguration{},
	}

	flag.StringVar(&configuration.ServerAddress, "a", "localhost:8080", "address and port of running server")
	flag.StringVar(&configuration.BaseURL, "b", "http://localhost:8080", "short link base URL")
	flag.StringVar(&configuration.AuditFile, "audit-file", "", "audit file path")
	flag.StringVar(&configuration.AuditURL, "audit-url", "", "audit endpoint URL")
	flag.StringVar(&configuration.FileStore.FilePath, "f", "shortener.jsonl", "path to storage file")
	flag.StringVar(&configuration.DatabaseStore.DataSourceName, "d", "", "data source name")
	flag.StringVar(&configuration.CPUProfile, "cpu-profile", "", "path to CPU profile file")
	flag.StringVar(&configuration.MemoryProfile, "memory-profile", "", "path to memory profile file")
	flag.Parse()

	return configuration
}

func parseEnvironment() (*Configuration, error) {
	configuration := &Configuration{
		FileStore:     &FileStoreConfiguration{},
		DatabaseStore: &DatabaseStoreConfiguration{},
	}
	err := env.Parse(configuration)

	return configuration, err
}
