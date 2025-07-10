/*
Package logistic provides a service layer responsible for managing LogisticData records,
which contain metadata about hardware, software, and device identification associated with PCBA numbers.

The LogisticDataService interface defines key business operations:
- Inserting LogisticData records with uniqueness ensured by PCBA number,
- Retrieving LogisticData by ID or PCBA number,
- Combined insert-or-retrieve logic to simplify upsert workflows.

Implementation details:
  - Input string fields are sanitized via trimming whitespace for consistency.
  - Conversion between DTO and DB models is handled by the dedicated converter package.
  - Checks for existing records by PCBA number prevent duplicate insertions.
  - Error handling distinguishes between missing records and operational errors,
    returning zero-value DTOs when no records exist.

Methods:

InsertLogisticData:
- Trims all relevant string fields in the provided DTO.
- Checks if a LogisticData record with the same PCBA number already exists.
- Inserts a new record only if no existing one is found.
- Returns the ID of the existing or newly inserted record.
- Returns an error on any database operation failure.

GetOrInsertLogisticData:
- Attempts to insert a LogisticData record unconditionally.
- Returns the assigned record ID or 0 if insertion failed or no ID was assigned.

GetById:
- Retrieves a LogisticData record by its integer ID.
- Returns a zero-value DTO if no record is found.
- Returns errors only for operational failures.

GetByPCBANumber:
- Retrieves a LogisticData record by PCBA number (trimmed).
- Returns a zero-value DTO if no record is found.
- Returns errors only for operational failures.

This package cleanly separates business logic from persistence,
enabling robust and maintainable handling of logistic metadata.
*/
package logistic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
	"github.com/NoroSaroyan/log-parser/internal/domain/repositories"
	"github.com/NoroSaroyan/log-parser/internal/services/converter/logistic"
	"strings"
)

// LogisticDataService defines operations for managing LogisticData records.
type LogisticDataService interface {
	// InsertLogisticData inserts a new LogisticData record if none exists for the PCBA number.
	// Returns the inserted or existing record ID.
	// Trims input string fields for consistency.
	InsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error)

	// GetOrInsertLogisticData attempts to insert LogisticData and returns its ID.
	// Returns 0 if insertion failed or no ID was assigned.
	GetOrInsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error)

	// GetById retrieves a LogisticData record by its ID.
	// Returns a zero-value DTO if no record is found.
	GetById(ctx context.Context, id int) (dto.LogisticDataDTO, error)

	// GetByPCBANumber retrieves a LogisticData record by PCBA number.
	// Returns a zero-value DTO if no record is found.
	GetByPCBANumber(ctx context.Context, PCBANumber string) (dto.LogisticDataDTO, error)
}

type logisticDataService struct {
	repo repositories.LogisticDataRepository
}

// NewLogisticDataService creates a new LogisticDataService using the provided repository.
func NewLogisticDataService(repo repositories.LogisticDataRepository) LogisticDataService {
	return &logisticDataService{repo: repo}
}

// InsertLogisticData trims input fields, checks if a record exists by PCBA number,
// inserts the data if new, and returns the record ID or existing record ID.
// Returns an error on database failure.
func (s *logisticDataService) InsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error) {
	// Trim all relevant string fields for clean input.
	data.PCBANumber = strings.TrimSpace(data.PCBANumber)
	data.ProductSN = strings.TrimSpace(data.ProductSN)
	data.PartNumber = strings.TrimSpace(data.PartNumber)
	data.VPAppVersion = strings.TrimSpace(data.VPAppVersion)
	data.VPBootLoaderVersion = strings.TrimSpace(data.VPBootLoaderVersion)
	data.VPCoreVersion = strings.TrimSpace(data.VPCoreVersion)
	data.SupplierHardwareVersion = strings.TrimSpace(data.SupplierHardwareVersion)
	data.ManufacturerHardwareVersion = strings.TrimSpace(data.ManufacturerHardwareVersion)
	data.ManufacturerSoftwareVersion = strings.TrimSpace(data.ManufacturerSoftwareVersion)
	data.BleMac = strings.TrimSpace(data.BleMac)
	data.BleSN = strings.TrimSpace(data.BleSN)
	data.BleVersion = strings.TrimSpace(data.BleVersion)
	data.BlePassworkKey = strings.TrimSpace(data.BlePassworkKey)
	data.APAppVersion = strings.TrimSpace(data.APAppVersion)
	data.APKernelVersion = strings.TrimSpace(data.APKernelVersion)
	data.TcuICCID = strings.TrimSpace(data.TcuICCID)
	data.PhoneNumber = strings.TrimSpace(data.PhoneNumber)
	data.IMEI = strings.TrimSpace(data.IMEI)
	data.IMSI = strings.TrimSpace(data.IMSI)
	data.ProductionDate = strings.TrimSpace(data.ProductionDate)

	// Check for existing record by PCBA number.
	existingID, err := s.repo.GetIDByPCBANumber(ctx, data.PCBANumber)
	if err == nil {
		return existingID, nil
	}

	dbModel := logistic.ConvertToDB(data)
	if err := s.repo.Insert(ctx, &dbModel); err != nil {
		return 0, fmt.Errorf("failed to insert LogisticData: %w", err)
	}
	return dbModel.ID, nil
}

// GetOrInsertLogisticData attempts to insert LogisticData and returns the assigned ID.
// Returns 0 if insertion failed or ID is zero.
func (s *logisticDataService) GetOrInsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error) {
	dbModel := logistic.ConvertToDB(data)
	err := s.repo.Insert(ctx, &dbModel)
	if err != nil {
		return 0, err
	}
	if dbModel.ID == 0 {
		return 0, nil
	}
	return dbModel.ID, nil
}

// GetById retrieves a LogisticData record by its ID.
// Returns zero-value DTO if not found, error on failure.
func (s *logisticDataService) GetById(ctx context.Context, id int) (dto.LogisticDataDTO, error) {
	dbModel, err := s.repo.GetById(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.LogisticDataDTO{}, nil
		}
		return dto.LogisticDataDTO{}, fmt.Errorf("failed to get LogisticData by ID: %w", err)
	}

	dtoModel := logistic.ConvertToDTO(*dbModel)
	return dtoModel, nil
}

// GetByPCBANumber retrieves a LogisticData record by PCBA number (trimmed).
// Returns zero-value DTO if not found, error on failure.
func (s *logisticDataService) GetByPCBANumber(ctx context.Context, PCBANumber string) (dto.LogisticDataDTO, error) {
	PCBANumber = strings.TrimSpace(PCBANumber)

	dbModel, err := s.repo.GetByPCBANumber(ctx, PCBANumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.LogisticDataDTO{}, nil
		}
		return dto.LogisticDataDTO{}, fmt.Errorf("failed to get LogisticData by PCBA number: %w", err)
	}

	dtoModel := logistic.ConvertToDTO(*dbModel)
	return dtoModel, nil
}
