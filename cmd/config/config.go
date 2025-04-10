package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr         string // адрес запуска HTTP-сервера
	BaseURL      string // базовый адрес результирующего URL
	FlagLogLevel string
}

var cfg *Config

func Init() *Config {
	addr := flag.String("a", "localhost:8080", "HTTP server address (e.g., localhost:8080)")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL for short links (e.g., http://localhost:8080/)")
	flagLogLevel := flag.String("l", "info", "Log level (debug, info, warn, error, fatal)")

	flag.Parse()

	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		*addr = envAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		*baseURL = envBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		*flagLogLevel = envLogLevel
	}

	cfg = &Config{
		Addr:         *addr,
		BaseURL:      *baseURL,
		FlagLogLevel: *flagLogLevel,
	}

	return cfg
}

func GetConfig() *Config {
	return cfg
}
