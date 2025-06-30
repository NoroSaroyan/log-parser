package models

import (
	"encoding/json"
)

type StationInformation struct {
	PartNumber       string       `json:"PartNumber"`
	TestStation      string       `json:"TestStation"`
	EntityType       string       `json:"EntityType"`
	ProductLine      string       `json:"ProductLine"`
	TestToolVersion  string       `json:"TestToolVersion"`
	TestFinishedTime string       `json:"TestFinishedTime"`
	IsAllPassed      bool         `json:"IsAllPassed"`
	ErrorCodes       string       `json:"ErrorCodes"`
	LogisticData     LogisticData `json:"LogisticData"`
}

func UnmarshalStationInformation(data []byte) (StationInformation, error) {
	var s StationInformation
	err := json.Unmarshal(data, &s)
	return s, err
}
