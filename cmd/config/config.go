package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr    string // адрес запуска HTTP-сервера
	BaseURL string // базовый адрес результирующего URL
}

var cfg *Config

func Init() *Config {
	addr := flag.String("a", "localhost:8080", "HTTP server address (e.g., localhost:8080)")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL for short links (e.g., http://localhost:8080/)")

	flag.Parse()

	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		*addr = envAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		*baseURL = envBaseURL
	}

	cfg = &Config{
		Addr:    *addr,
		BaseURL: *baseURL,
	}

	return cfg
}

func GetConfig() *Config {
	return cfg
}
