package downloadinfo

import (
	"context"
	"fmt"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/download"
)

type DownloadInfoService interface {
	InsertDownloadInfo(ctx context.Context, data dto.DownloadInfoDTO) error
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
