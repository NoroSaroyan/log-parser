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
	GetOrInsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error) // 👈 add

}

type logisticDataService struct {
	repo repositories.LogisticDataRepository
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

func (s *logisticDataService) GetOrInsertLogisticData(ctx context.Context, data dto.LogisticDataDTO) (int, error) {
	id, err := s.repo.GetIDByPCBANumber(ctx, data.PCBANumber)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query LogisticData: %w", err)
	}

	return s.InsertLogisticData(ctx, data)
}
