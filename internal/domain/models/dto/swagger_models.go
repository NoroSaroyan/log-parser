package dto

// swagger models for the generation and correct documentation
type TestStationWithSteps struct {
	TestStationRecordDTO
	TestSteps []TestStepDTO `json:"TestSteps"`
}
type PCBANumbersResponse struct {
	PCBANumbers []string `json:"PCBANumbers"`
}
