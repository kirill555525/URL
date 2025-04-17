package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Addr            string // адрес запуска HTTP-сервера
	BaseURL         string // базовый адрес результирующего URL
	FlagLogLevel    string
	FileStoragePath string
	DatabaseDSN     string
}

var cfg *Config

func Init() *Config {

	ps := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, `5432`, `postgres`, `postgres`, `postgres`)

	addr := flag.String("a", "localhost:8080", "HTTP server address (e.g., localhost:8080)")
	baseURL := flag.String("b", "http://localhost:8080", "Base URL for short links (e.g., http://localhost:8080/)")
	flagLogLevel := flag.String("l", "info", "Log level (debug, info, warn, error, fatal)")
	flagFileStoragePath := flag.String("f", "/tmp/short-url-db.json", "File storage path")
	flagDatabaseDSN := flag.String("d", ps, "Database DSN")

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

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		*flagDatabaseDSN = envDatabaseDSN
	}

	cfg = &Config{
		Addr:            *addr,
		BaseURL:         *baseURL,
		FlagLogLevel:    *flagLogLevel,
		FileStoragePath: *flagFileStoragePath,
		DatabaseDSN:     *flagDatabaseDSN,
	}

	return cfg
}

func GetConfig() *Config {
	return cfg
}
