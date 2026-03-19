package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	SessionSecret      string
	DatabaseURL        string
	TCPAddr            string
}

// LoadConfig reads the environment variables from .env and returns a Config struct
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	return &Config{
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		SessionSecret:      getEnv("SESSION_SECRET", "super-secret-key"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://admin:admin@localhost:5432/hdp?sslmode=disable"),
		TCPAddr:            getEnv("TCP_ADDR", ":9090"),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
