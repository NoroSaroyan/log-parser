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

type DownloadInfoService interface {
	InsertDownloadInfo(ctx context.Context, data dto.DownloadInfoDTO) error
	GetByPCBANumber(ctx context.Context, pcbaNumber string) (dto.DownloadInfoDTO, error)
}

type downloadInfoService struct {
	repo repositories.DownloadInfoRepository
}

func NewDownloadInfoService(repo repositories.DownloadInfoRepository) DownloadInfoService {
	if repo == nil {
		log.Fatal("Repo is nil in NewDownloadInfoService")
	}
	return &downloadInfoService{repo: repo}
}

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
func (s *downloadInfoService) GetByPCBANumber(ctx context.Context, pcbaNumber string) (dto.DownloadInfoDTO, error) {
	pcbaNumber = strings.TrimSpace(pcbaNumber)
	if s == nil || s.repo == nil {
		return dto.DownloadInfoDTO{}, fmt.Errorf("downloadInfoService or repo is nil")
	}

	dbModel, err := s.repo.GetByPCBANumber(ctx, pcbaNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.DownloadInfoDTO{}, nil
		}
		return dto.DownloadInfoDTO{}, fmt.Errorf("failed to get DownloadInfo by PCBA number: %w", err)
	}

	if dbModel == nil {
		return dto.DownloadInfoDTO{}, nil
	}

	dtoModel := download.ConvertToDTO(*dbModel)
	return dtoModel, nil
}
