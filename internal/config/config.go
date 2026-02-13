package config

import "os"

type Config struct {
	DBPath   string
	HTTPPort string
}

func New() Config {
	dbPath := getEnv("DB_PATH", "./data/certwatch.db")

	return Config{
		DBPath:   dbPath,
		HTTPPort: "8080",
	}
}

func getEnv(key string, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
