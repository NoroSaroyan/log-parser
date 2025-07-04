package teststep

import (
	"log-parser/internal/domain/models/db"
	"log-parser/internal/domain/models/dto"
)

func ConvertToDB(dto dto.TestStepDTO, testStationRecordID int) db.TestStepDB {
	return db.TestStepDB{
		TestStepName:        dto.TestStepName,
		TestThresholdValue:  dto.TestThresholdValue,
		TestMeasuredValue:   dto.TestMeasuredValue,
		TestStepElapsedTime: dto.TestStepElapsedTime,
		TestStepResult:      dto.TestStepResult,
		TestStepErrorCode:   dto.TestStepErrorCode,
		TestStationRecordID: testStationRecordID,
	}
}

func ConvertToDTO(db db.TestStepDB) dto.TestStepDTO {
	return dto.TestStepDTO{
		TestStepName:        db.TestStepName,
		TestThresholdValue:  db.TestThresholdValue,
		TestMeasuredValue:   db.TestMeasuredValue,
		TestStepElapsedTime: db.TestStepElapsedTime,
		TestStepResult:      db.TestStepResult,
		TestStepErrorCode:   db.TestStepErrorCode,
	}
}
