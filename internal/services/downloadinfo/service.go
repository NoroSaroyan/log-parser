package downloadinfo

import (
	"context"
	"database/sql"
	"fmt"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/download"
)

type DownloadInfoService interface {
	InsertDownloadInfo(ctx context.Context, data dto.DownloadInfoDTO) error
	GetByPCBANumber(ctx context.Context, pcbaNumber string) (dto.DownloadInfoDTO, error)
}

type downloadInfoService struct {
	repo repositories.DownloadInfoRepository
}

func NewDownloadInfoService(repo repositories.DownloadInfoRepository) DownloadInfoService {
	return &downloadInfoService{repo: repo}
}

func (s *downloadInfoService) InsertDownloadInfo(ctx context.Context, data dto.DownloadInfoDTO) error {
	dbModel := download.ConvertToDB(data)

	if err := s.repo.Insert(ctx, &dbModel); err != nil {
		return fmt.Errorf("failed to insert DownloadInfo: %w", err)
	}
	return nil
}
func (s *downloadInfoService) GetByPCBANumber(ctx context.Context, pcbaNumber string) (dto.DownloadInfoDTO, error) {
	dbModel, err := s.repo.GetByPCBANumber(ctx, pcbaNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.DownloadInfoDTO{}, nil
		}
		return dto.DownloadInfoDTO{}, fmt.Errorf("failed to get LogisticData by ID: %w", err)
	}

	dtoModel := download.ConvertToDTO(*dbModel)
	return dtoModel, nil
}
