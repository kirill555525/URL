package config

import (
	"flag"
	"os"
)

type Config struct {
	Addr            string // адрес запуска HTTP-сервера
	BaseURL         string // базовый адрес результирующего URL
	FlagLogLevel    string
	FileStoragePath string
}

var cfg *Config

func Init() *Config {
	addr := flag.String("a", "localhost:8080", "HTTP server address (e.g., localhost:8080)")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL for short links (e.g., http://localhost:8080/)")
	flagLogLevel := flag.String("l", "info", "Log level (debug, info, warn, error, fatal)")
	flagFileStoragePath := flag.String("f", "/tmp/short-url-db.json", "File storage path")

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

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		*flagFileStoragePath = envFileStoragePath
	}

	cfg = &Config{
		Addr:            *addr,
		BaseURL:         *baseURL,
		FlagLogLevel:    *flagLogLevel,
		FileStoragePath: *flagFileStoragePath,
	}

	return cfg
}

func GetConfig() *Config {
	return cfg
}
