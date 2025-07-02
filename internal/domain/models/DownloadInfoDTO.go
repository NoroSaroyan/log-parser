package models

type DownloadInfoDTO struct {
	TestStation          string `json:"TestStation"`
	FlashEntityType      string `json:"FlashEntityType"`
	TcuPCBANumber        string `json:"TcuPCBANumber"`
	FlashElapsedTime     int    `json:"FlashElapsedTime"`
	TcuEntityFlashState  string `json:"TcuEntityFlashState"`
	PartNumber           string `json:"PartNumber"`
	ProductLine          string `json:"ProductLine"`
	DownloadToolVersion  string `json:"DownloadToolVersion"`
	DownloadFinishedTime string `json:"DownloadFinishedTime"`
}
