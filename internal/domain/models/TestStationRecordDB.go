package models

type TestStationRecordDB struct {
	ID               int    `db:"id"`
	PartNumber       string `db:"part_number"`
	TestStation      string `db:"test_station"`
	EntityType       string `db:"entity_type"`
	ProductLine      string `db:"product_line"`
	TestToolVersion  string `db:"test_tool_version"`
	TestFinishedTime string `db:"test_finished_time"`
	IsAllPassed      bool   `db:"is_all_passed"`
	ErrorCodes       string `db:"error_codes"`
	LogisticDataID   int    `db:"logistic_data_id"`
}
