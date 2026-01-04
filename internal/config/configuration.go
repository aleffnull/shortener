package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
	HTTPS         *HTTPSConfiguration         `validate:"required"`
	CPUProfile    string                      `env:"CPU_PROFILE" validate:"omitempty,filepath"`
	MemoryProfile string                      `env:"MEMORY_PROFILE" validate:"omitempty,filepath"`
	TrustedSubnet string                      `env:"TRUSTED_SUBNET" validate:"omitempty,cidr"`
	ConfigFile    string                      `env:"CONFIG"`
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

	if c.HTTPS == nil {
		fmt.Fprintf(sb, " HTTPS:<nil>")
	} else {
		fmt.Fprintf(sb, " HTTPS:%v", c.HTTPS)
	}

	if len(c.TrustedSubnet) > 0 {
		fmt.Fprintf(sb, " TrustedSubnet:%v", c.TrustedSubnet)
	}

	if len(c.ConfigFile) > 0 {
		fmt.Fprintf(sb, " ConfigFile:%v", c.ConfigFile)
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
	fileConfig, err := parseFile(envConfig, flagConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	configuration := &Configuration{
		ServerAddress: getStringValue(envConfig.ServerAddress, flagConfig.ServerAddress, fileConfig.ServerAddress),
		BaseURL:       getStringValue(envConfig.BaseURL, flagConfig.BaseURL, fileConfig.BaseURL),
		AuditFile:     getStringValue(envConfig.AuditFile, flagConfig.AuditFile, fileConfig.AuditFile),
		AuditURL:      getStringValue(envConfig.AuditURL, flagConfig.AuditURL, fileConfig.AuditURL),
		MemoryStore:   defaultMemoryStoreConfiguration(),
		FileStore: &FileStoreConfiguration{
			FilePath: getStringValue(
				envConfig.FileStore.FilePath,
				flagConfig.FileStore.FilePath,
				fileConfig.FileStore.FilePath,
			),
		},
		DatabaseStore: NewDatabaseStoreConfiguration(
			getStringValue(
				envConfig.DatabaseStore.DataSourceName,
				flagConfig.DatabaseStore.DataSourceName,
				fileConfig.DatabaseStore.DataSourceName,
			)),
		HTTPS: &HTTPSConfiguration{
			Enabled: envConfig.HTTPS.Enabled || flagConfig.HTTPS.Enabled || fileConfig.HTTPS.Enabled,
			CertificateFile: getStringValue(
				envConfig.HTTPS.CertificateFile,
				flagConfig.HTTPS.CertificateFile,
				fileConfig.HTTPS.CertificateFile,
			),
			KeyFile: getStringValue(
				envConfig.HTTPS.KeyFile,
				flagConfig.HTTPS.KeyFile,
				fileConfig.HTTPS.KeyFile,
			),
		},
		CPUProfile:    getStringValue(envConfig.CPUProfile, flagConfig.CPUProfile, fileConfig.CPUProfile),
		MemoryProfile: getStringValue(envConfig.MemoryProfile, flagConfig.MemoryProfile, fileConfig.MemoryProfile),
		TrustedSubnet: getStringValue(envConfig.TrustedSubnet, flagConfig.TrustedSubnet, fileConfig.TrustedSubnet),
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
		HTTPS:         &HTTPSConfiguration{},
	}

	flag.StringVar(&configuration.ServerAddress, "a", "localhost:8080", "address and port of running server")
	flag.StringVar(&configuration.BaseURL, "b", "http://localhost:8080", "short link base URL")
	flag.StringVar(&configuration.AuditFile, "audit-file", "", "audit file path")
	flag.StringVar(&configuration.AuditURL, "audit-url", "", "audit endpoint URL")
	flag.StringVar(&configuration.FileStore.FilePath, "f", "shortener.jsonl", "path to storage file")
	flag.StringVar(&configuration.DatabaseStore.DataSourceName, "d", "", "data source name")
	flag.BoolVar(&configuration.HTTPS.Enabled, "s", false, "use HTTPS")
	flag.StringVar(&configuration.HTTPS.CertificateFile, "cert-file", "", "HTTP certificate file")
	flag.StringVar(&configuration.HTTPS.KeyFile, "key-file", "", "HTTP key file")
	flag.StringVar(&configuration.CPUProfile, "cpu-profile", "", "path to CPU profile file")
	flag.StringVar(&configuration.MemoryProfile, "memory-profile", "", "path to memory profile file")
	flag.StringVar(&configuration.TrustedSubnet, "t", "", "trusted subnet CIDR")
	flag.StringVar(&configuration.ConfigFile, "config", "", "path to configuration file")
	flag.Parse()

	return configuration
}

func parseEnvironment() (*Configuration, error) {
	configuration := &Configuration{
		FileStore:     &FileStoreConfiguration{},
		DatabaseStore: &DatabaseStoreConfiguration{},
		HTTPS:         &HTTPSConfiguration{},
	}
	err := env.Parse(configuration)

	return configuration, err
}

func parseFile(envConfig *Configuration, flagConfig *Configuration) (*Configuration, error) {
	configFile := lo.Ternary(len(envConfig.ConfigFile) > 0, envConfig.ConfigFile, flagConfig.ConfigFile)
	if len(configFile) == 0 {
		return &Configuration{
			FileStore:     &FileStoreConfiguration{},
			DatabaseStore: &DatabaseStoreConfiguration{},
			HTTPS:         &HTTPSConfiguration{},
		}, nil
	}

	jsonBytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%v': %w", configFile, err)
	}

	configurationFile := &ConfigurationFile{}
	if err := json.Unmarshal(jsonBytes, configurationFile); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from config file '%v': %w", configFile, err)
	}

	configuration := &Configuration{
		ServerAddress: configurationFile.ServerAddress,
		BaseURL:       configurationFile.BaseURL,
		AuditFile:     configurationFile.AuditFile,
		AuditURL:      configurationFile.AuditURL,
		FileStore: &FileStoreConfiguration{
			FilePath: configurationFile.FileStoreFilePath,
		},
		DatabaseStore: &DatabaseStoreConfiguration{
			DataSourceName: configurationFile.DatabaseStoreDataSourceName,
		},
		HTTPS: &HTTPSConfiguration{
			Enabled:         configurationFile.HTTPSEnabled,
			CertificateFile: configurationFile.HTTPSCertificateFile,
			KeyFile:         configurationFile.HTTPSKeyFile,
		},
		CPUProfile:    configurationFile.CPUProfile,
		MemoryProfile: configurationFile.MemoryProfile,
		TrustedSubnet: configurationFile.TrustedSubnet,
	}

	return configuration, nil
}

func getStringValue(envValue string, flagValue string, fileValue string) string {
	if len(envValue) > 0 {
		return envValue
	}

	if len(flagValue) > 0 {
		return flagValue
	}

	return fileValue
}
