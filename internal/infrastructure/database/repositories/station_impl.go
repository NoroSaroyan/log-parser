package database

import (
	"database/sql"
	"log-parser/internal/domain/models"
	"log-parser/internal/domain/repositories"
)

type stationRepo struct {
	db *sql.DB
}

func NewStationRepository(db *sql.DB) repositories.StationRepository {
	return &stationRepo{db: db}
}

func (r *stationRepo) Save(station *models.StationInformation) error {
	// SQL insert/update logic here
	return nil
}

func (r *stationRepo) FindByID(id string) (*models.StationInformation, error) {
	// SQL select logic here
	return nil, nil
}
