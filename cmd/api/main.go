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

	"log-parser/internal/app"
	v1 "log-parser/internal/handlers/api/v1"
)

func main() {
	application, err := app.InitializeApp("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer application.CloseDB()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	v1.RegisterAPIV1(
		r,
		application.DownloadInfoService,
		application.LogisticService,
		application.TestStationService,
		application.TestStepService,
	)

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
