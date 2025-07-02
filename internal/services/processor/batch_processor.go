package processor

import (
	"log-parser/internal/domain/models"
)

type GroupedData struct {
	DownloadInfo       *models.DownloadInfoDTO
	TestStations       []*models.TestStationRecordDTO
	TestStepsByStation map[string][]*models.TestStepDTO
}
