package app

import (
	"fmt"
	"github.com/NoroSaroyan/log-parser/internal/config"
	"github.com/NoroSaroyan/log-parser/internal/infrastructure/database"
	"github.com/NoroSaroyan/log-parser/internal/infrastructure/database/repositories"
	"github.com/NoroSaroyan/log-parser/internal/services/downloadinfo"
	"github.com/NoroSaroyan/log-parser/internal/services/logistic"
	"github.com/NoroSaroyan/log-parser/internal/services/teststation"
	"github.com/NoroSaroyan/log-parser/internal/services/teststep"
	"log"
)

type App struct {
	DownloadInfoService downloadinfo.DownloadInfoService
	LogisticService     logistic.LogisticDataService
	TestStationService  teststation.TestStationService
	TestStepService     teststep.TestStepService
	CloseDB             func() error
}

func InitializeApp(configPath string) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	db, err := database.NewPostgresDB(&cfg.Database)

	if err != nil {
		log.Printf("Failed to connect to database: %w", err)
		fmt.Printf("Config: %v", cfg.Database)
		return nil, err
	}
	log.Println("Connected to Postgres database")

	downloadRepo := repositories.NewDownloadInfoRepository(db)
	logisticRepo := repositories.NewLogisticDataRepository(db)
	testStationRepo := repositories.NewTestStationRecordRepository(db)
	testStepRepo := repositories.NewTestStepRepository(db)

	downloadService := downloadinfo.NewDownloadInfoService(downloadRepo)
	logisticService := logistic.NewLogisticDataService(logisticRepo)
	testStationService := teststation.NewTestStationService(testStationRepo)
	testStepService := teststep.NewTestStepService(testStepRepo)

	app := &App{
		DownloadInfoService: downloadService,
		LogisticService:     logisticService,
		TestStationService:  testStationService,
		TestStepService:     testStepService,
		CloseDB:             db.Close,
	}

	return app, nil
}
