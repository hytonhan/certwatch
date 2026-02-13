package config

import (
	"os"
	"time"
)

type Config struct {
	DBPath              string
	HTTPPort            string
	ExpiryCheckInterval time.Duration
	ExpiryWindow        time.Duration
}

func New() Config {
	dbPath := getEnv("DB_PATH", "./data/certwatch.db")

	return Config{
		DBPath:              dbPath,
		HTTPPort:            "8080",
		ExpiryCheckInterval: time.Minute,
		ExpiryWindow:        time.Hour * 72,
	}
}

func getEnv(key string, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
