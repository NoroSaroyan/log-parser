package logistic

import (
	"context"
	"database/sql"
	"fmt"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/domain/repositories"
	"log-parser/internal/services/converter/logistic"
	"strings"
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
	PCBANumber = strings.TrimSpace(PCBANumber)

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
