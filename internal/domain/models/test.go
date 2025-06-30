package models

import (
	"encoding/json"
)

type TestStep struct {
	TestStepName        string `json:"TestStepName"`
	TestThresholdValue  string `json:"TestThresholdValue"`
	TestMeasuredValue   string `json:"TestMeasuredValue"`
	TestStepElapsedTime int    `json:"TestStepElapsedTime"`
	TestStepResult      string `json:"TestStepResult"`
	TestStepErrorCode   string `json:"TestStepErrorCode"`
}

type TestData []TestStep

func UnmarshalTestData(data []byte) (TestData, error) {
	var t TestData
	err := json.Unmarshal(data, &t)
	return t, err
}
