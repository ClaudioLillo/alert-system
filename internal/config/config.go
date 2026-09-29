package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config stores all application settings.
type Config struct {
	HTTPAddr     string
	MigrationDir string
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
}

// Load reads environment variables and returns the runtime configuration.
func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		HTTPAddr:     getEnv("HTTP_ADDR", "0.0.0.0:8080"),
		MigrationDir: getEnv("MIGRATION_DIR", "./db/migrations"),
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         getEnv("DB_PORT", "5432"),
		User:         getEnv("DB_USER", "postgres"),
		Password:     getEnv("DB_PASSWORD", "postgres"),
		Name:         getEnv("DB_NAME", "alert_system"),
		SSLMode:      getEnv("DB_SSLMODE", "require"),
	}

	return cfg
}

// DSN builds a PostgreSQL connection string from the environment variables. It keeps
// support for a legacy DB_URL value if someone still sets it in the environment.
func (c Config) DSN() string {
	if c.Host != "" || c.User != "" || c.Name != "" {
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.Host,
			c.Port,
			c.User,
			c.Password,
			c.Name,
			c.SSLMode,
		)
	}

	if value := os.Getenv("DB_URL"); value != "" {
		return value
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
		c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
