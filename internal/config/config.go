package config

import (
	"flag"
	"net/url"

	"github.com/caarlos0/env/v6"
	"github.com/samber/lo"
)

type Configuration struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

var (
	Current *Configuration
)

func ParseConfiguration() error {
	envConfig, err := parseEnvironment()
	if err != nil {
		return err
	}

	flagConfig, err := parseFlags()
	if err != nil {
		return err
	}

	Current = &Configuration{
		ServerAddress: lo.Ternary(len(envConfig.ServerAddress) > 0, envConfig.ServerAddress, flagConfig.ServerAddress),
		BaseURL:       lo.Ternary(len(envConfig.BaseURL) > 0, envConfig.BaseURL, flagConfig.BaseURL),
	}

	_, err = url.ParseRequestURI(Current.BaseURL)
	return err
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
