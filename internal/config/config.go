package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port     string
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN monta a string de conexão no formato key=value aceito pelo driver.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// Load lê a configuração do ambiente. Os defaults casam com o docker-compose,
// então `make up` funciona sem nenhum arquivo .env.
func Load() Config {
	return Config{
		Port: env("APP_PORT", "8080"),
		Database: DatabaseConfig{
			Host:     env("DB_HOST", "localhost"),
			Port:     env("DB_PORT", "5432"),
			User:     env("DB_USER", "restock"),
			Password: env("DB_PASSWORD", "restock"),
			Name:     env("DB_NAME", "restock"),
			SSLMode:  env("DB_SSLMODE", "disable"),
		},
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
