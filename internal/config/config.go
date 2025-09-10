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
}

func Load() *Config {
	if err := godotenv.Load("config.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	var AppConfig Config

	AppConfig = Config{
		PostgresURL: getEnv("POSTGRES_URL", "postgres://user:password@localhost:5432/projectdb?sslmode=disable"),
		ApiAddress:  getEnv("API_ADDRESS", ":8080"),
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
