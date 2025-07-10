/*
Package teststep provides a service layer responsible for managing TestStep data,
which represents individual test steps performed within a test station record context.

The TestStepService interface defines key business operations:
- Inserting multiple test steps linked to a specific TestStationRecord,
- Retrieving all test steps associated with a given TestStationRecordID.

Implementation notes:
  - The service sanitizes string fields of each test step DTO before processing,
    ensuring consistent and clean data input.
  - Conversion between DTO and DB models is handled by the dedicated converter package.
  - Batch insertion is performed via the repository to optimize database operations.
  - Retrieval operations convert DB models back into DTOs for external use.

Methods:

InsertTestSteps:
- Accepts a slice of TestStepDTOs and the parent TestStationRecordID.
- Trims whitespace from all relevant string fields including nested measured values.
- Converts each DTO to its DB representation and aggregates them.
- Delegates batch insertion to the repository.
- Returns a wrapped error if insertion fails.

GetByTestStationRecordID:
- Fetches all TestStep records linked to a specific TestStationRecordID.
- Converts retrieved DB models to DTOs before returning.
- Returns an error if retrieval fails.

This package encapsulates the business logic for handling detailed test step data,
with a clear separation between domain models, service logic, and persistence layers.
*/
package teststep

import (
	"context"
	"fmt"
	"github.com/NoroSaroyan/log-parser/internal/domain/models/db"
	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
	"github.com/NoroSaroyan/log-parser/internal/domain/repositories"
	"github.com/NoroSaroyan/log-parser/internal/services/converter/teststep"
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
