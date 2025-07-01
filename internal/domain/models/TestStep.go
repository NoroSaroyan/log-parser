package models

type TestStep struct {
	TestStepName        string `json:"TestStepName"`
	TestThresholdValue  string `json:"TestThresholdValue"`
	TestMeasuredValue   string `json:"TestMeasuredValue"`
	TestStepElapsedTime int    `json:"TestStepElapsedTime"`
	TestStepResult      string `json:"TestStepResult"`
	TestStepErrorCode   string `json:"TestStepErrorCode"`
}
