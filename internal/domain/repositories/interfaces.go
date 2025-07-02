package repositories

import (
	"log-parser/internal/domain/models"
)

type DownloadInfoRepository interface {
	Insert(info *models.DownloadInfoDTO) error
	GetByPCBANumber(pcba string) (*models.DownloadInfoDTO, error)
}

type LogisticDataRepository interface {
	Insert(data *models.LogisticDataDTO) (int, error)
	GetByPCBANumber(pcba string) (*models.LogisticDataDTO, error)
}

type TestStationRecordRepository interface {
	Insert(record *models.TestStationRecordDTO) (int, error)
	GetByPCBANumber(pcba string) ([]*models.TestStationRecordDTO, error)
}

type TestStepRepository interface {
	InsertBatch(steps []*models.TestStepDTO, stationID int) error
	GetByPCBANumber(pcba string) ([]*models.TestStepDTO, error)
}
