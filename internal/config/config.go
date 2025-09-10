package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

type Config struct {
	PostgresURL string
	ApiAddress  string
	Env         string
	LogLevel    string
}

func Load() *Config {
	_ = godotenv.Load("config.env")

	var AppConfig Config

	AppConfig = Config{
		PostgresURL: getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/projectdb?sslmode=disable"),
		ApiAddress:  getEnv("API_ADDRESS", ":8080"),
		Env:         getEnv("ENV", "prod"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	log.Println("Config loaded")
	return &AppConfig
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}
