package repositories

import (
	"log-parser/internal/domain/models"
)

type StationRepository interface {
	Save(station *models.StationInformation) error
	FindByID(id string) (*models.StationInformation, error)
}

type TestRepository interface {
	Save(test *models.TestData) error
	FindByID(id string) (*models.TestData, error)
}
