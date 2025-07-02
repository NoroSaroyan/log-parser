package processor

import (
	"log-parser/internal/domain/models/dto"
)

type GroupedData struct {
	DownloadInfo       *dto.DownloadInfoDTO
	TestStations       []*dto.TestStationRecordDTO
	TestStepsByStation map[string][]*dto.TestStepDTO
}
