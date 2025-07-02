package dto

type GroupedDataDTO struct {
	Download  *DownloadInfoDTO
	Stations  []*TestStationRecordDTO
	TestSteps []*TestStepDTO
}

var dataMap map[string]*GroupedDataDTO // key = PCBANumber
