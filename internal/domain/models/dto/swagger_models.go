package dto

type TestStationWithSteps struct {
	TestStationRecordDTO
	TestSteps []TestStepDTO `json:"TestSteps"`
}
type PCBANumbersResponse struct {
	PCBANumbers []string `json:"PCBANumbers"`
}
