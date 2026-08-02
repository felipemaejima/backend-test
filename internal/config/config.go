package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port     string
	Database DatabaseConfig
	Log      LogConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type LogConfig struct {
	Level  string
	Format string
	File   string
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

func Load() (Config, error) {
	cfg := Config{
		Port: env("APP_PORT", "8080"),
		Database: DatabaseConfig{
			Host:     env("DB_HOST", "localhost"),
			Port:     env("DB_PORT", "5432"),
			User:     env("DB_USER", "restock"),
			Password: env("DB_PASSWORD", "restock"),
			Name:     env("DB_NAME", "restock"),
			SSLMode:  env("DB_SSLMODE", "disable"),
		},
		Log: LogConfig{
			Level:  env("LOG_LEVEL", "info"),
			Format: env("LOG_FORMAT", "json"),
			File:   os.Getenv("LOG_FILE"),
		},
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// validate falha no boot em vez de deixar um valor inválido explodir na
// primeira requisição, com mensagem obscura.
func (c Config) validate() error {
	if err := validatePort("APP_PORT", c.Port); err != nil {
		return err
	}
	if err := validatePort("DB_PORT", c.Database.Port); err != nil {
		return err
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST não pode ser vazio")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME não pode ser vazio")
	}
	if err := validateOneOf("LOG_LEVEL", c.Log.Level, "debug", "info", "warn", "error"); err != nil {
		return err
	}
	return validateOneOf("LOG_FORMAT", c.Log.Format, "json", "text")
}

func validatePort(name, value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%s=%q não é um número", name, value)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s=%d fora da faixa 1-65535", name, port)
	}
	return nil
}

func validateOneOf(name, value string, allowed ...string) error {
	for _, candidate := range allowed {
		if strings.EqualFold(value, candidate) {
			return nil
		}
	}
	return fmt.Errorf("%s=%q inválido; use um de: %s", name, value, strings.Join(allowed, ", "))
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
