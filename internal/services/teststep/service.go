package teststep

import (
	"context"
	"fmt"
	"log-parser/internal/domain/models/db"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/teststep"
)

type TestStepService interface {
	InsertTestSteps(ctx context.Context, steps []dto.TestStepDTO, testStationRecordID int) error
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
		converted := teststep.ConvertToDB(step, testStationRecordID)
		dbModels = append(dbModels, &converted)
	}

	if err := s.repo.InsertBatch(ctx, dbModels, testStationRecordID); err != nil {
		return fmt.Errorf("failed to insert TestSteps: %w", err)
	}

	return nil
}
