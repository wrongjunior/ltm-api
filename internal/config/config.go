package config

import (
	"os"
)

// Config содержит конфигурации для приложения
type Config struct {
	Port string
}

// LoadConfig загружает конфигурации приложения
func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Порт по умолчанию
	}

	return Config{
		Port: port,
	}
}
