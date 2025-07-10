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

// DispatcherService defines the interface for dispatching grouped data DTOs
// to the respective services that insert or update data in the database.
type DispatcherService interface {
	// DispatchGroups processes multiple GroupedDataDTO objects sequentially.
	//
	// For each group:
	//   1. Inserts DownloadInfo if present.
	//   2. Inserts LogisticData for each TestStationRecord, ensuring a valid LogisticDataID.
	//   3. Inserts each TestStationRecord linked to the corresponding LogisticDataID.
	//   4. Inserts TestSteps linked to their corresponding TestStationRecord.
	//
	// If any insertion fails, the error is returned immediately and processing stops.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeout propagation.
	//   - groups: slice of grouped data DTOs, each containing DownloadInfo, TestStationRecords, and TestSteps.
	//
	// Returns:
	//   - error if any insertion operation fails or data mismatches occur.
	DispatchGroups(ctx context.Context, groups []dto.GroupedDataDTO) error
}

type dispatcherService struct {
	downloadInfoService downloadinfo.DownloadInfoService
	logisticDataService logistic.LogisticDataService
	testStationService  teststation.TestStationService
	testStepService     teststep.TestStepService
}

// NewDispatcherService creates a new DispatcherService implementation with the required dependencies.
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

// DispatchGroups implements DispatcherService.DispatchGroups.
//
// See interface documentation for full details.
func (s *dispatcherService) DispatchGroups(ctx context.Context, groups []dto.GroupedDataDTO) error {
	for _, group := range groups {
		if (group.DownloadInfo != dto.DownloadInfoDTO{}) {
			if err := s.downloadInfoService.InsertDownloadInfo(ctx, group.DownloadInfo); err != nil {
				return fmt.Errorf("failed to insert DownloadInfo for PCBA %s: %w", group.DownloadInfo.TcuPCBANumber, err)
			}
		}

		testStationIDs := make([]int, 0, len(group.TestStationRecords))
		for _, tsr := range group.TestStationRecords {
			var logisticDataID int
			var err error

			if (tsr.LogisticData != dto.LogisticDataDTO{}) {
				logisticDataID, err = s.logisticDataService.GetOrInsertLogisticData(ctx, tsr.LogisticData)
				if err != nil {
					fmt.Printf("DEBUG: LogisticDataID for PCBA %s = %d\n", tsr.LogisticData.PCBANumber, logisticDataID)
					return fmt.Errorf("failed to insert LogisticData for PCBA %s: %w", tsr.LogisticData.PCBANumber, err)
				}
				if logisticDataID == 0 {
					fmt.Printf("DEBUG: LogisticDataDTO: %+v\n", tsr.LogisticData)
					return fmt.Errorf("resolved LogisticDataID is 0 for PCBA %s", tsr.LogisticData.PCBANumber)
				}
			} else {
				return fmt.Errorf("missing LogisticData for PCBA %s: cannot insert TestStationRecord", tsr.LogisticData.PCBANumber)
			}

			testStationID, err := s.testStationService.InsertTestStationRecord(ctx, tsr, logisticDataID)
			if err != nil {
				return fmt.Errorf("failed to insert TestStationRecord for PCBA %s: %w", tsr.LogisticData.PCBANumber, err)
			}
			testStationIDs = append(testStationIDs, testStationID)
		}

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
