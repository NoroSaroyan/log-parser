package db

type TestStepDB struct {
	ID                  int    `db:"id"`
	TestStepName        string `db:"test_step_name"`
	TestThresholdValue  string `db:"test_threshold_value"`
	TestMeasuredValue   string `db:"test_measured_value"`
	TestStepElapsedTime int    `db:"test_step_elapsed_time"`
	TestStepResult      string `db:"test_step_result"`
	TestStepErrorCode   string `db:"test_step_error_code"`
	TestStationRecordID int    `db:"test_station_record_id"`
}
