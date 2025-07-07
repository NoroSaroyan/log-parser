package teststation

import (
	"context"
	"fmt"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/teststation"
	"strings"
)

type TestStationService interface {
	InsertTestStationRecord(ctx context.Context, data dto.TestStationRecordDTO, logisticDataID int) (int, error)
	GetByPCBANumber(ctx context.Context, pcbaNumber string) ([]dto.TestStationRecordDTO, error)
}

type testStationService struct {
	repo repositories.TestStationRecordRepository
}

func NewTestStationService(repo repositories.TestStationRecordRepository) TestStationService {
	return &testStationService{repo: repo}
}

func (s *testStationService) InsertTestStationRecord(ctx context.Context, data dto.TestStationRecordDTO, logisticDataID int) (int, error) {
	data.PartNumber = strings.TrimSpace(data.PartNumber)
	data.TestStation = strings.TrimSpace(data.TestStation)
	data.EntityType = strings.TrimSpace(data.EntityType)
	data.ProductLine = strings.TrimSpace(data.ProductLine)
	data.TestToolVersion = strings.TrimSpace(data.TestToolVersion)
	data.TestFinishedTime = strings.TrimSpace(data.TestFinishedTime)
	data.ErrorCodes = strings.TrimSpace(data.ErrorCodes)

	dbModel := teststation.ConvertToDB(data)
	dbModel.LogisticDataID = logisticDataID

	if err := s.repo.Insert(ctx, &dbModel); err != nil {
		return 0, fmt.Errorf("failed to insert TestStationRecord: %w", err)
	}
	if dbModel.ID == 0 {
		return 0, fmt.Errorf("unexpected: inserted TestStationRecord returned ID=0 for PCBA=%s", dbModel.PartNumber)
	}

	return dbModel.ID, nil
}

func (s *testStationService) GetByPCBANumber(ctx context.Context, pcbaNumber string) ([]dto.TestStationRecordDTO, error) {
	pcbaNumber = strings.TrimSpace(pcbaNumber)

	dbRecords, err := s.repo.GetByPCBANumber(ctx, pcbaNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get TestStationRecords by PCBA number: %w", err)
	}

	var dtos []dto.TestStationRecordDTO
	for _, dbRec := range dbRecords {
		dtos = append(dtos, teststation.ConvertToDTO(*dbRec))
	}

	return dtos, nil
}
