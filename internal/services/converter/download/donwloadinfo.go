package download

import (
	db "github.com/NoroSaroyan/log-parser/internal/domain/models/db"
	dto "github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
)

func ConvertToDB(dto dto.DownloadInfoDTO) db.DownloadInfoDB {
	return db.DownloadInfoDB{
		TestStation:          dto.TestStation,
		FlashEntityType:      dto.FlashEntityType,
		TcuPCBANumber:        dto.TcuPCBANumber,
		FlashElapsedTime:     dto.FlashElapsedTime,
		TcuEntityFlashState:  dto.TcuEntityFlashState,
		PartNumber:           dto.PartNumber,
		ProductLine:          dto.ProductLine,
		DownloadToolVersion:  dto.DownloadToolVersion,
		DownloadFinishedTime: dto.DownloadFinishedTime,
	}
}

func ConvertToDTO(db db.DownloadInfoDB) dto.DownloadInfoDTO {
	return dto.DownloadInfoDTO{
		TestStation:          db.TestStation,
		FlashEntityType:      db.FlashEntityType,
		TcuPCBANumber:        db.TcuPCBANumber,
		FlashElapsedTime:     db.FlashElapsedTime,
		TcuEntityFlashState:  db.TcuEntityFlashState,
		PartNumber:           db.PartNumber,
		ProductLine:          db.ProductLine,
		DownloadToolVersion:  db.DownloadToolVersion,
		DownloadFinishedTime: db.DownloadFinishedTime,
	}
}
