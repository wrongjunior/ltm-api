package main

import (
	"log"
	"net/http"
	_ "os"

	"ltm-api/internal/config"
	"ltm-api/internal/router"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Создаем новый роутер с chi
	r := router.NewRouter()

	// Запуск сервера
	log.Printf("Starting server on port %s...", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
