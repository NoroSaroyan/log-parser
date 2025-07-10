/*
Package teststation provides a service layer responsible for managing TestStationRecord data,
which represents domain information about individual test stations linked to PCBA units.

The TestStationService interface defines the core business operations:
- Inserting new test station records linked to logistic data,
- Retrieving test station records by PCBA number in both DTO and DB model forms,
- Fetching all distinct PCBA numbers present in the system (optionally filtered by station type).

Implementation details:
  - The service relies on a TestStationRecordRepository to interact with the underlying database,
    abstracting persistence logic from the service layer.
  - Input strings are sanitized by trimming whitespace to ensure data consistency.
  - Conversion between DTO and DB models is handled via a dedicated converter package.

Methods:

InsertTestStationRecord:
- Accepts a TestStationRecordDTO and a logisticDataID foreign key.
- Cleans string fields for uniformity.
- Converts DTO to DB model, sets the logistic data ID, and inserts into the database.
- Returns the newly inserted record's ID or an error if insertion fails or returns invalid ID.

GetByPCBANumber:
- Retrieves all TestStationRecordDTO entries matching the provided PCBA number.
- Trims input string and queries repository.
- Converts DB records to DTOs before returning.
- Returns error if the query fails.

GetDbObjectsByPCBANumber:
- Similar to GetByPCBANumber but returns raw DB models.
- Validates non-empty PCBA number input.
- Returns error if no records are found or query fails.

GetAllPCBANumbers:
- Fetches all distinct PCBA numbers from the repository.
- Intended for listing or validation use cases.
- The `stationType` parameter is currently unused but reserved for future filtering support.

Overall, this service encapsulates business logic around test station records with a focus on data integrity,
cleanliness, and clear separation of concerns between domain, service, and persistence layers.
*/
package teststation

import (
	"context"
	"fmt"
	"github.com/NoroSaroyan/log-parser/internal/domain/models/db"
	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
	"github.com/NoroSaroyan/log-parser/internal/domain/repositories"
	"github.com/NoroSaroyan/log-parser/internal/services/converter/teststation"
	"strings"
)

type TestStationService interface {
	InsertTestStationRecord(ctx context.Context, data dto.TestStationRecordDTO, logisticDataID int) (int, error)
	GetByPCBANumber(ctx context.Context, pcbaNumber string) ([]dto.TestStationRecordDTO, error)
	GetDbObjectsByPCBANumber(ctx context.Context, pcbaNumber string) ([]*db.TestStationRecordDB, error)
	GetAllPCBANumbers(ctx context.Context, stationType string) ([]string, error)
}

// testStationService is the concrete implementation of TestStationService.
type testStationService struct {
	repo repositories.TestStationRecordRepository
}

// NewTestStationService creates a new TestStationService with the given repository dependency.
func NewTestStationService(repo repositories.TestStationRecordRepository) TestStationService {
	return &testStationService{repo: repo}
}

// InsertTestStationRecord inserts a TestStationRecordDTO linked to logisticDataID into the database.
//
// It trims whitespace from relevant string fields, converts the DTO to a DB model,
// and uses the repository to persist it. Returns the new record's ID or an error.
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

// GetByPCBANumber fetches all TestStationRecordDTOs associated with the given PCBA number.
//
// Trims the input PCBA number string and queries the repository.
// Converts DB models to DTOs before returning.
// Returns an error if the repository query fails.
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

// GetDbObjectsByPCBANumber returns raw DB model TestStationRecordDB objects for the given PCBA number.
//
// Validates the PCBA number is not empty.
// Returns an error if no records are found or if the query fails.
func (s *testStationService) GetDbObjectsByPCBANumber(ctx context.Context, pcbaNumber string) ([]*db.TestStationRecordDB, error) {
	pcbaNumber = strings.TrimSpace(pcbaNumber)
	if pcbaNumber == "" {
		return nil, fmt.Errorf("pcbaNumber cannot be empty")
	}

	dbRecords, err := s.repo.GetByPCBANumber(ctx, pcbaNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get TestStationRecords by PCBA number: %w", err)
	}

	if len(dbRecords) == 0 {
		return nil, fmt.Errorf("no TestStationRecords found for PCBA number: %s", pcbaNumber)
	}

	return dbRecords, nil
}

// GetAllPCBANumbers returns all distinct PCBA numbers stored in the repository.
//
// The stationType parameter is currently unused but reserved for future enhancements such as filtering by test station type.
func (s *testStationService) GetAllPCBANumbers(ctx context.Context, stationType string) ([]string, error) {
	return s.repo.GetAllPCBANumbers(ctx)
}
