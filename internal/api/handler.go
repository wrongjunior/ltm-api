package api

import (
	"encoding/json"
	"ltm-api/estimator"
	"net/http"
	"os"
	_ "strconv"
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

	// Сохраняем файл временно на диск
	tempFile, err := os.CreateTemp("", "upload-*.txt")
	if err != nil {
		http.Error(w, "Unable to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	// Записываем файл на диск
	_, err = tempFile.ReadFrom(file)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	tempFile.Close()

	// Используем новый метод потоковой обработки текста
	result, err := estimator.StreamProcessFile(tempFile.Name(), 200, false, 4)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
