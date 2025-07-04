package teststation

import (
	"context"
	"fmt"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/teststation"
)

type TestStationService interface {
	InsertTestStationRecord(ctx context.Context, data dto.TestStationRecordDTO, logisticDataID int) (int, error)
}

type testStationService struct {
	repo repositories.TestStationRecordRepository
}

func NewTestStationService(repo repositories.TestStationRecordRepository) TestStationService {
	return &testStationService{repo: repo}
}

func (s *testStationService) InsertTestStationRecord(ctx context.Context, data dto.TestStationRecordDTO, logisticDataID int) (int, error) {
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
