package teststation

import (
	db "github.com/NoroSaroyan/log-parser/internal/domain/models/db"
	dto "github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
)

func ConvertToDB(dto dto.TestStationRecordDTO) db.TestStationRecordDB {
	return db.TestStationRecordDB{
		PartNumber:       dto.PartNumber,
		TestStation:      dto.TestStation,
		EntityType:       dto.EntityType,
		ProductLine:      dto.ProductLine,
		TestToolVersion:  dto.TestToolVersion,
		TestFinishedTime: dto.TestFinishedTime,
		IsAllPassed:      dto.IsAllPassed,
		ErrorCodes:       dto.ErrorCodes,
		LogisticDataID:   dto.LogisticDataID,
	}
}

func ConvertToDTO(db db.TestStationRecordDB) dto.TestStationRecordDTO {
	return dto.TestStationRecordDTO{
		PartNumber:       db.PartNumber,
		TestStation:      db.TestStation,
		EntityType:       db.EntityType,
		ProductLine:      db.ProductLine,
		TestToolVersion:  db.TestToolVersion,
		TestFinishedTime: db.TestFinishedTime,
		IsAllPassed:      db.IsAllPassed,
		ErrorCodes:       db.ErrorCodes,
		LogisticDataID:   db.LogisticDataID,
	}
}
