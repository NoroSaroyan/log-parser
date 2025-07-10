package dto

// Represents a group of data to be converted and inserted into database
type GroupedDataDTO struct {
	DownloadInfo       DownloadInfoDTO
	TestStationRecords []TestStationRecordDTO
	TestSteps          [][]TestStepDTO
}
