package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"ltm-api/internal/api"
)

// NewRouter создает новый chi роутер и определяет маршруты
func NewRouter() http.Handler {
	r := chi.NewRouter()

	// Определение маршрутов
	r.Post("/estimate/reading-time", api.EstimateReadingTime)
	r.Post("/estimate/upload", api.EstimateFromFile)

	return r
}
