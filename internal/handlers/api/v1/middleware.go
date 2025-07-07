package v1

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// JSON — наш набор middleware для API
var JSON = []func(http.Handler) http.Handler{
	// 1) Логируем запросы
	middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: log.New(log.Writer(), "", log.LstdFlags),
	}),
	// 2) Recover от panic
	middleware.Recoverer,
	// 3) Таймаут 15 секунд
	middleware.Timeout(15 * time.Second),
}

// respondJSON — helper для JSON‑ответов
func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// respondError — helper для ошибок
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
