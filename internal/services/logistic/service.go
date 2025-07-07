package logistic

import (
	"context"
	"database/sql"
	"fmt"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/logistic"
)

type LogisticDataService interface {
	InsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error)
	GetOrInsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error)
	GetById(ctx context.Context, id int) (dto.LogisticDataDTO, error)
	GetByPCBANumber(ctx context.Context, PCBANumber string) (dto.LogisticDataDTO, error)
}

type logisticDataService struct {
	repo repositories.LogisticDataRepository
}

func (s *logisticDataService) GetOrInsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error) {
	//TODO implement me
	panic("implement me")
}

func NewLogisticDataService(repo repositories.LogisticDataRepository) LogisticDataService {
	return &logisticDataService{repo: repo}
}

func (s *logisticDataService) InsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error) {
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

func (s *logisticDataService) GetByPCBANumber(ctx context.Context, PCBANumber string) (dto.LogisticDataDTO, error) {
	dbModel, err := s.repo.GetByPCBANumber(ctx, PCBANumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.LogisticDataDTO{}, nil
		}
		return dto.LogisticDataDTO{}, fmt.Errorf("failed to get LogisticData by ID: %w", err)
	}

	dtoModel := logistic.ConvertToDTO(*dbModel)
	return dtoModel, nil
}
