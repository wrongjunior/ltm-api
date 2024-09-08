package api

import (
	"encoding/json"
	"io"
	"ltm-api/estimator"
	"net/http"
)

// EstimateReadingTime обрабатывает запрос на оценку времени чтения текста
func EstimateReadingTime(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text         string  `json:"text"`
		ReadingSpeed float64 `json:"readingSpeed"`
		HasVisuals   bool    `json:"hasVisuals"`
		WorkerCount  int     `json:"workerCount"`
	}

	// Декодируем тело запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Вызываем пакет estimator для анализа текста
	result, err := estimator.EstimateReadingTimeParallel(req.Text, req.ReadingSpeed, req.HasVisuals, req.WorkerCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// EstimateFromFile обрабатывает запрос на анализ текста из файла
func EstimateFromFile(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Читаем содержимое файла
	fileContent, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Используем стандартную скорость чтения и параметры
	result, err := estimator.EstimateReadingTimeParallel(string(fileContent), 200, false, 4)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
