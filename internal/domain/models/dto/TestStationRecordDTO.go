package dto

type TestStationRecordDTO struct {
	PartNumber       string          `json:"PartNumber"`
	TestStation      string          `json:"TestStation"` // "PCBA", "Final", etc.
	EntityType       string          `json:"EntityType"`
	ProductLine      string          `json:"ProductLine"`
	TestToolVersion  string          `json:"TestToolVersion"`
	TestFinishedTime string          `json:"TestFinishedTime"`
	IsAllPassed      bool            `json:"IsAllPassed"`
	ErrorCodes       string          `json:"ErrorCodes"`
	LogisticDataID   int             `json:"LogisticDataID"`
	LogisticData     LogisticDataDTO `json:"LogisticData"`
}
