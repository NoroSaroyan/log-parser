/*
Package downloadinfo provides a service layer for managing DownloadInfo records,
which contain information related to firmware or software downloads tied to PCBA numbers.

The DownloadInfoService interface offers methods to:
- Insert new DownloadInfo records after sanitizing input,
- Retrieve DownloadInfo records by PCBA number,
with error handling that distinguishes between missing records and operational failures.

Conversion between DTO and DB models is handled by the dedicated converter package.
*/
package downloadinfo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/download"
	"strings"
)

// DownloadInfoService defines operations for managing DownloadInfo records.
type DownloadInfoService interface {
	// InsertDownloadInfo inserts a new DownloadInfo record after trimming input fields.
	//
	// Parameters:
	// - ctx: context for request lifecycle management.
	// - data: DownloadInfoDTO with the data to insert.
	//
	// Returns:
	// - error if insertion fails.
	InsertDownloadInfo(ctx context.Context, data dto.DownloadInfoDTO) error

	// GetByPCBANumber retrieves a DownloadInfo record by PCBA number.
	//
	// Parameters:
	// - ctx: context for request lifecycle management.
	// - pcbaNumber: PCBA number string to query.
	//
	// Returns:
	// - DownloadInfoDTO if found; zero-value DTO otherwise.
	// - error if the query fails for reasons other than no rows found.
	GetByPCBANumber(ctx context.Context, pcbaNumber string) (dto.DownloadInfoDTO, error)
}

type downloadInfoService struct {
	repo repositories.DownloadInfoRepository
}

// NewDownloadInfoService creates a new DownloadInfoService with the given repository.
// Panics if the repository is nil.
func NewDownloadInfoService(repo repositories.DownloadInfoRepository) DownloadInfoService {
	if repo == nil {
		log.Fatal("DownloadInfoRepository is nil in NewDownloadInfoService")
	}
	return &downloadInfoService{repo: repo}
}

// InsertDownloadInfo trims whitespace from input string fields and inserts the record.
func (s *downloadInfoService) InsertDownloadInfo(ctx context.Context, data dto.DownloadInfoDTO) error {
	data.TestStation = strings.TrimSpace(data.TestStation)
	data.FlashEntityType = strings.TrimSpace(data.FlashEntityType)
	data.TcuPCBANumber = strings.TrimSpace(data.TcuPCBANumber)
	data.TcuEntityFlashState = strings.TrimSpace(data.TcuEntityFlashState)
	data.PartNumber = strings.TrimSpace(data.PartNumber)
	data.ProductLine = strings.TrimSpace(data.ProductLine)
	data.DownloadToolVersion = strings.TrimSpace(data.DownloadToolVersion)
	data.DownloadFinishedTime = strings.TrimSpace(data.DownloadFinishedTime)

	dbModel := download.ConvertToDB(data)

	if err := s.repo.Insert(ctx, &dbModel); err != nil {
		return fmt.Errorf("failed to insert DownloadInfo: %w", err)
	}
	return nil
}

// GetByPCBANumber retrieves a DownloadInfo record by PCBA number.
// Returns zero-value DTO if not found, error on operational failure.
func (s *downloadInfoService) GetByPCBANumber(ctx context.Context, pcbaNumber string) (dto.DownloadInfoDTO, error) {
	pcbaNumber = strings.TrimSpace(pcbaNumber)

	if s == nil || s.repo == nil {
		return dto.DownloadInfoDTO{}, fmt.Errorf("downloadInfoService or repository is nil")
	}

	dbModel, err := s.repo.GetByPCBANumber(ctx, pcbaNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			// Not found: return zero-value DTO without error.
			return dto.DownloadInfoDTO{}, nil
		}
		return dto.DownloadInfoDTO{}, fmt.Errorf("failed to get DownloadInfo by PCBA number: %w", err)
	}

	if dbModel == nil {
		// No record found.
		return dto.DownloadInfoDTO{}, nil
	}

	dtoModel := download.ConvertToDTO(*dbModel)
	return dtoModel, nil
}
