package dto

type GroupedDataDTO struct {
	DownloadInfo       DownloadInfoDTO
	TestStationRecords []TestStationRecordDTO
	TestSteps          [][]TestStepDTO // массив массивов тестов, каждый для TestStationRecord
}
