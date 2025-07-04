package app

import (
	"log"
	"log-parser/internal/config"
	"log-parser/internal/infrastructure/database"
	"log-parser/internal/infrastructure/database/repositories"
	"log-parser/internal/services/downloadinfo"
	"log-parser/internal/services/logistic"
	"log-parser/internal/services/teststation"
	"log-parser/internal/services/teststep"
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
		log.Println("Failed to connect to database", err)
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
