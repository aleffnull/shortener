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
	ServerAddress string                    `env:"SERVER_ADDRESS" validate:"required,hostname_port"`
	BaseURL       string                    `env:"BASE_URL" validate:"required,url"`
	MemoryStore   *MemoryStoreConfiguration `validate:"required"`
}

func (c *Configuration) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(
		sb,
		"&Configuration{ServerAddress:%v BaseURL:%v",
		c.ServerAddress,
		c.BaseURL)

	if c.MemoryStore == nil {
		fmt.Fprintf(sb, " MemoryStore:<nil>")
	} else {
		fmt.Fprintf(sb, " MemoryStore:%v", c.MemoryStore)
	}

	fmt.Fprintf(sb, "}")
	return sb.String()
}

func GetConfiguration() (*Configuration, error) {
	envConfig, err := parseEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	flagConfig, err := parseFlags()
	if err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	configuration := &Configuration{
		ServerAddress: lo.Ternary(len(envConfig.ServerAddress) > 0, envConfig.ServerAddress, flagConfig.ServerAddress),
		BaseURL:       lo.Ternary(len(envConfig.BaseURL) > 0, envConfig.BaseURL, flagConfig.BaseURL),
		MemoryStore:   defaultMemoryStoreConfiguration(),
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(configuration)
	if err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return configuration, nil
}

func parseFlags() (*Configuration, error) {
	configuration := &Configuration{}

	flag.StringVar(&configuration.ServerAddress, "a", "localhost:8080", "address and port of running server")
	flag.StringVar(&configuration.BaseURL, "b", "http://localhost:8080", "short link base URL")
	flag.Parse()

	return configuration, nil
}

func parseEnvironment() (*Configuration, error) {
	configuration := &Configuration{}
	err := env.Parse(configuration)

	return configuration, err
}
