package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/NoroSaroyan/log-parser/internal/app"
	v1 "github.com/NoroSaroyan/log-parser/internal/handlers/api/v1"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/NoroSaroyan/log-parser/cmd/api/docs"
)

// @title Log Parser API
// @version 1.0
// @description API for log parsing service
// @host localhost:8080
// @BasePath /api/v1
func main() {
	file := os.Getenv("CONFIG_FILE")
	if len(file) == 0 {
		file = "configs/config.yaml"
	}
	application, err := app.InitializeApp(file)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer application.CloseDB()

	r := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	})
	r.Use(corsMiddleware.Handler)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		v1.RegisterAPIV1(r,
			application.DownloadInfoService,
			application.LogisticService,
			application.TestStationService,
			application.TestStepService,
		)
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	errs := make(chan error, 1)

	go func() {
		log.Printf("Starting REST API server on %s", server.Addr)
		errs <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Received shutdown signal, shutting down server gracefully...")
	if err := server.Close(); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	if err := <-errs; err != nil && err != http.ErrServerClosed {
		log.Fatalf("Fatal error: %v", err)
	}

	log.Println("Server stopped")
}
