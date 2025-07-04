package dto

type GroupedDataDTO struct {
	DownloadInfo       DownloadInfoDTO
	TestStationRecords []TestStationRecordDTO
	TestSteps          [][]TestStepDTO
}
