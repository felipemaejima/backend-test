package config_test

import (
	"strings"
	"testing"

	"github.com/felipemaejima/backend-test/internal/config"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"APP_PORT", "DB_HOST", "DB_PORT", "DB_USER",
		"DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"LOG_LEVEL", "LOG_FORMAT", "LOG_FILE",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, expected 8080", cfg.Port)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("Host = %q, expected localhost", cfg.Database.Host)
	}
	if cfg.Log.Level != "info" || cfg.Log.Format != "json" {
		t.Errorf("Log = %+v, expected info/json", cfg.Log)
	}
	if cfg.Log.File != "" {
		t.Errorf("File = %q, expected empty so logs go only to stdout", cfg.Log.File)
	}
}

func TestLoadFromEnv(t *testing.T) {
	clearEnv(t)
	t.Setenv("APP_PORT", "9090")
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, expected 9090", cfg.Port)
	}
	if cfg.Database.Host != "postgres" {
		t.Errorf("Host = %q, expected postgres", cfg.Database.Host)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Level = %q, expected debug", cfg.Log.Level)
	}
	if cfg.Database.User != "restock" {
		t.Errorf("User = %q, expected the default", cfg.Database.User)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"non numeric app port", "APP_PORT", "abc"},
		{"app port out of range", "APP_PORT", "70000"},
		{"app port zero", "APP_PORT", "0"},
		{"negative database port", "DB_PORT", "-1"},
		{"non numeric database port", "DB_PORT", "porta"},
		{"unknown log level", "LOG_LEVEL", "verbose"},
		{"misspelled log level", "LOG_LEVEL", "warining"},
		{"unknown log format", "LOG_FORMAT", "xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv(tt.key, tt.value)

			cfg, err := config.Load()
			if err == nil {
				t.Fatalf("expected an error for %s=%q, got %+v", tt.key, tt.value, cfg)
			}
			if !strings.Contains(err.Error(), tt.key) {
				t.Errorf("error %q should name the offending variable %s", err, tt.key)
			}
		})
	}
}

func TestLogLevelIsCaseInsensitive(t *testing.T) {
	clearEnv(t)
	t.Setenv("LOG_LEVEL", "WARN")
	t.Setenv("LOG_FORMAT", "TEXT")

	if _, err := config.Load(); err != nil {
		t.Fatalf("uppercase values should be accepted: %v", err)
	}
}

func TestDSN(t *testing.T) {
	dsn := config.DatabaseConfig{
		Host:     "postgres",
		Port:     "5432",
		User:     "restock",
		Password: "secret",
		Name:     "restock",
		SSLMode:  "disable",
	}.DSN()

	for _, fragment := range []string{
		"host=postgres", "port=5432", "user=restock",
		"password=secret", "dbname=restock", "sslmode=disable",
	} {
		if !strings.Contains(dsn, fragment) {
			t.Errorf("DSN missing %q: %s", fragment, dsn)
		}
	}
}
