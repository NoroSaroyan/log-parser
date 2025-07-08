package dto

import "fmt"

type TestStepDTO struct {
	TestStepName        string      `json:"TestStepName"`
	TestThresholdValue  string      `json:"TestThresholdValue"`
	TestMeasuredValue   interface{} `json:"TestMeasuredValue"`
	TestStepElapsedTime int         `json:"TestStepElapsedTime"`
	TestStepResult      string      `json:"TestStepResult"`
	TestStepErrorCode   string      `json:"TestStepErrorCode"`
}

func (t *TestStepDTO) GetMeasuredValueString() string {
	return fmt.Sprintf("%v", t.TestMeasuredValue)
}
