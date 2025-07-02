package models

type DownloadInfoDB struct {
	ID                   int    `db:"id"`
	TestStation          string `db:"test_station"`
	FlashEntityType      string `db:"flash_entity_type"`
	TcuPCBANumber        string `db:"tcu_pcba_number"`
	FlashElapsedTime     int    `db:"flash_elapsed_time"`
	TcuEntityFlashState  string `db:"tcu_entity_flash_state"`
	PartNumber           string `db:"part_number"`
	ProductLine          string `db:"product_line"`
	DownloadToolVersion  string `db:"download_tool_version"`
	DownloadFinishedTime string `db:"download_finished_time"`
}
