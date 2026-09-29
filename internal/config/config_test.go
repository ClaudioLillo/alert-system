package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.Host == "" {
		t.Fatal("Host must not be empty")
	}
	if cfg.Port == "" {
		t.Fatal("Port must not be empty")
	}
}

func TestDSNUsesEnvironmentVariables(t *testing.T) {
	oldHost := os.Getenv("DB_HOST")
	oldPort := os.Getenv("DB_PORT")
	oldUser := os.Getenv("DB_USER")
	oldPassword := os.Getenv("DB_PASSWORD")
	oldName := os.Getenv("DB_NAME")
	oldSSLMode := os.Getenv("DB_SSLMODE")
	defer func() {
		_ = os.Setenv("DB_HOST", oldHost)
		_ = os.Setenv("DB_PORT", oldPort)
		_ = os.Setenv("DB_USER", oldUser)
		_ = os.Setenv("DB_PASSWORD", oldPassword)
		_ = os.Setenv("DB_NAME", oldName)
		_ = os.Setenv("DB_SSLMODE", oldSSLMode)
	}()

	_ = os.Setenv("DB_HOST", "db.example.com")
	_ = os.Setenv("DB_PORT", "5432")
	_ = os.Setenv("DB_USER", "app_user")
	_ = os.Setenv("DB_PASSWORD", "secret")
	_ = os.Setenv("DB_NAME", "alert_system")
	_ = os.Setenv("DB_SSLMODE", "require")

	cfg := Load()
	dsn := cfg.DSN()
	if dsn == "" {
		t.Fatal("DSN must not be empty")
	}
	if dsn != "host=db.example.com port=5432 user=app_user password=secret dbname=alert_system sslmode=require" {
		t.Fatalf("unexpected DSN: %s", dsn)
	}
}
