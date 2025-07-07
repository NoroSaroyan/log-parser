package main

import (
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"log-parser/internal/infrastructure/database/repositories"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"log-parser/internal/config"
	"log-parser/internal/handlers/api/v1"
	"log-parser/internal/handlers/cli"
	"log-parser/internal/infrastructure/database"
	"log-parser/internal/services/downloadinfo"
	"log-parser/internal/services/logistic"
	"log-parser/internal/services/teststation"
	"log-parser/internal/services/teststep"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к базе данных
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация репозиториев
	downloadRepo := repositories.NewDownloadInfoRepository(db)
	logisticRepo := repositories.NewLogisticDataRepository(db)
	tsRepo := repositories.NewTestStationRecordRepository(db)
	stepRepo := repositories.NewTestStepRepository(db)

	// Инициализация сервисов
	downloadSvc := downloadinfo.NewDownloadInfoService(downloadRepo)
	logisticSvc := logistic.NewLogisticDataService(logisticRepo)
	testStationSvc := teststation.NewTestStationService(tsRepo)
	testStepSvc := teststep.NewTestStepService(stepRepo)

	// Инициализация роутера и middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Регистрация маршрутов API v1
	v1.RegisterAPIV1(r, downloadSvc, logisticSvc, testStationSvc, testStepSvc)

	// Конфигурация HTTP-сервера
	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Канал ошибок (для двух горутин и shutdown)
	errs := make(chan error, 2)

	// Запуск REST API
	go func() {
		log.Printf("Starting REST API server on %s", cfg.Server.Address)
		errs <- server.ListenAndServe()
	}()

	// Запуск CLI лог-парсера
	go func() {
		log.Println("Starting CLI log parser")
		if err := cli.Run(); err != nil {
			errs <- err
		}
	}()

	// Graceful shutdown по SIGINT/SIGTERM
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		log.Println("Received shutdown signal, shutting down server gracefully...")
		server.Close()
		errs <- nil
	}()

	// Ожидание ошибок или сигнала
	if err := <-errs; err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}
