// Package v1 Provides HTTP middleware and helper functions for building JSON APIs.
// It defines common middleware for request logging, recovery, and timeouts, as well
// as utility functions for sending JSON responses in a consistent format.
package v1

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// JSON is a collection of HTTP middleware handlers for JSON APIs.
//
// It includes:
//
//   - A request logger with standard log formatting.
//   - A panic recovery middleware that converts panics into HTTP 500 errors.
//   - A timeout middleware that enforces a maximum request duration of 15 seconds.
//
// Example usage:
//
//	r := chi.NewRouter()
//	r.Use(v1.JSON...)
//
// These middleware are designed to be composable and work well in Go HTTP servers.
var JSON = []func(http.Handler) http.Handler{
	middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: log.New(log.Writer(), "", log.LstdFlags),
	}),
	middleware.Recoverer,
	middleware.Timeout(15 * time.Second),
}

// respondJSON writes a JSON-encoded response with the given HTTP status code.
//
// It sets the Content-Type header to "application/json" and serializes the provided
// payload to JSON. If encoding fails, it writes an empty body with the given status code.
//
// Example usage:
//
//	data := map[string]string{"message": "ok"}
//	v1.respondJSON(w, http.StatusOK, data)
func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// respondError writes a JSON-encoded error response with the given HTTP status code.
//
// The response body contains a single "error" field with the provided message.
// It is intended for sending standardized error messages to API clients.
//
// Example usage:
//
//	v1.respondError(w, http.StatusBadRequest, "invalid request payload")
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
