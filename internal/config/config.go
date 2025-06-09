package config

import (
	"flag"
	"net/url"
)

type Configuration struct {
	ServerAddress string
	BaseURL       string
}

var (
	Current Configuration
)

func ParseFlags() error {
	flag.StringVar(&Current.ServerAddress, "a", "localhost:8080", "address and port of running server")
	flag.StringVar(&Current.BaseURL, "b", "http://localhost:8080", "short link base URL")

	flag.Parse()

	_, err := url.ParseRequestURI(Current.BaseURL)
	return err
}
