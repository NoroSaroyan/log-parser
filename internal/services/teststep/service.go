package teststep

import (
	"context"
	"fmt"
	"log-parser/internal/domain/models/db"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/teststep"
	"strings"
)

type TestStepService interface {
	InsertTestSteps(ctx context.Context, steps []dto.TestStepDTO, testStationRecordID int) error
	GetByTestStationRecordID(ctx context.Context, testStationRecordID int) ([]dto.TestStepDTO, error)
}

type testStepService struct {
	repo repositories.TestStepRepository
}

func NewTestStepService(repo repositories.TestStepRepository) TestStepService {
	return &testStepService{repo: repo}
}

func (s *testStepService) InsertTestSteps(ctx context.Context, steps []dto.TestStepDTO, testStationRecordID int) error {
	var dbModels []*db.TestStepDB
	for _, step := range steps {
		step.TestStepName = strings.TrimSpace(step.TestStepName)
		step.TestStepResult = strings.TrimSpace(step.TestStepResult)
		step.TestStepErrorCode = strings.TrimSpace(step.TestStepErrorCode)
		step.TestThresholdValue = strings.TrimSpace(step.TestThresholdValue)
		step.TestMeasuredValue = strings.TrimSpace(step.GetMeasuredValueString())

		converted := teststep.ConvertToDB(step, testStationRecordID)
		dbModels = append(dbModels, &converted)
	}

	if err := s.repo.InsertBatch(ctx, dbModels, testStationRecordID); err != nil {
		return fmt.Errorf("failed to insert TestSteps: %w", err)
	}

	return nil
}

func (s *testStepService) GetByTestStationRecordID(ctx context.Context, testStationRecordID int) ([]dto.TestStepDTO, error) {
	dbSteps, err := s.repo.GetByTestStationRecordID(ctx, testStationRecordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get TestSteps by TestStationRecordID: %w", err)
	}

	var dtos []dto.TestStepDTO
	for _, dbStep := range dbSteps {
		dtos = append(dtos, teststep.ConvertToDTO(*dbStep))
	}

	return dtos, nil
}
