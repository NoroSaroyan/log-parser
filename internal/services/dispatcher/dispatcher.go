package dispatcher

import (
	"context"
	"fmt"
	"log-parser/internal/domain/models/dto"
	"log-parser/internal/services/downloadinfo"
	"log-parser/internal/services/logistic"
	"log-parser/internal/services/teststation"
	"log-parser/internal/services/teststep"
)

type DispatcherService interface {
	DispatchGroups(ctx context.Context, groups []dto.GroupedDataDTO) error
}

type dispatcherService struct {
	downloadInfoService downloadinfo.DownloadInfoService
	logisticDataService logistic.LogisticDataService
	testStationService  teststation.TestStationService
	testStepService     teststep.TestStepService
}

func NewDispatcherService(
	downloadInfoSvc downloadinfo.DownloadInfoService,
	logisticSvc logistic.LogisticDataService,
	testStationSvc teststation.TestStationService,
	testStepSvc teststep.TestStepService,
) DispatcherService {
	return &dispatcherService{
		downloadInfoService: downloadInfoSvc,
		logisticDataService: logisticSvc,
		testStationService:  testStationSvc,
		testStepService:     testStepSvc,
	}
}

func (s *dispatcherService) DispatchGroups(ctx context.Context, groups []dto.GroupedDataDTO) error {
	for _, group := range groups {
		// Insert DownloadInfo
		if (group.DownloadInfo != dto.DownloadInfoDTO{}) {
			if err := s.downloadInfoService.InsertDownloadInfo(ctx, group.DownloadInfo); err != nil {
				return fmt.Errorf("failed to insert DownloadInfo for PCBA %s: %w", group.DownloadInfo.TcuPCBANumber, err)
			}
		}

		testStationIDs := make([]int, 0, len(group.TestStationRecords))
		for _, tsr := range group.TestStationRecords {
			var logisticDataID int
			var err error

			// Insert LogisticData
			if (tsr.LogisticData != dto.LogisticDataDTO{}) {
				logisticDataID, err = s.logisticDataService.GetOrInsertLogisticData(ctx, tsr.LogisticData)
				if err != nil {
					fmt.Printf("DEBUG: LogisticDataID for PCBA %s = %d\n", tsr.LogisticData.PCBANumber, logisticDataID)
					return fmt.Errorf("failed to insert LogisticData for PCBA %s: %w", tsr.LogisticData.PCBANumber, err)
				}
				if logisticDataID == 0 {
					return fmt.Errorf("resolved LogisticDataID is 0 for PCBA %s", tsr.LogisticData.PCBANumber)
				}
			} else {
				return fmt.Errorf("missing LogisticData for PCBA %s: cannot insert TestStationRecord", tsr.LogisticData.PCBANumber)
			}

			// Insert TestStationRecord
			testStationID, err := s.testStationService.InsertTestStationRecord(ctx, tsr, logisticDataID)
			if err != nil {
				return fmt.Errorf("failed to insert TestStationRecord for PCBA %s: %w", tsr.LogisticData.PCBANumber, err)
			}
			testStationIDs = append(testStationIDs, testStationID)
		}

		// Insert TestSteps
		for i, stepsSlice := range group.TestSteps {
			if i >= len(testStationIDs) {
				return fmt.Errorf("mismatch in TestSteps count and TestStationRecords count for PCBA %s", group.DownloadInfo.TcuPCBANumber)
			}
			if err := s.testStepService.InsertTestSteps(ctx, stepsSlice, testStationIDs[i]); err != nil {
				return fmt.Errorf("failed to insert TestSteps for TestStationRecordID %d: %w", testStationIDs[i], err)
			}
		}
	}

	return nil
}
