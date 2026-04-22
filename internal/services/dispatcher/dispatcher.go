package dispatcher

import (
	"context"
	"fmt"
	"strings"

	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
	"github.com/NoroSaroyan/log-parser/internal/infrastructure/logger"
	"github.com/NoroSaroyan/log-parser/internal/services/downloadinfo"
	"github.com/NoroSaroyan/log-parser/internal/services/logistic"
	"github.com/NoroSaroyan/log-parser/internal/services/parser"
	"github.com/NoroSaroyan/log-parser/internal/services/teststation"
	"github.com/NoroSaroyan/log-parser/internal/services/teststep"
)

// groupKey returns the best-available PCBA-like identifier for a group,
// preferring station record PCBA, then Download TCU PCBA, then ProductSN.
// Used for logging so error lines never show an empty identifier.
func groupKey(group dto.GroupedDataDTO) string {
	for _, tsr := range group.TestStationRecords {
		if p := strings.TrimSpace(tsr.LogisticData.PCBANumber); p != "" {
			return p
		}
	}
	if p := strings.TrimSpace(group.DownloadInfo.TcuPCBANumber); p != "" {
		return p
	}
	for _, tsr := range group.TestStationRecords {
		if p := strings.TrimSpace(tsr.LogisticData.ProductSN); p != "" {
			return "productsn:" + p
		}
	}
	return "<unknown>"
}

// describeGroup builds a compact structured summary of a group for logging.
// Intentionally verbose — meant for diagnostic WARN/ERROR paths.
func describeGroup(group dto.GroupedDataDTO) map[string]interface{} {
	stations := make([]map[string]interface{}, 0, len(group.TestStationRecords))
	for idx, tsr := range group.TestStationRecords {
		stations = append(stations, map[string]interface{}{
			"index":        idx,
			"station_type": strings.TrimSpace(tsr.TestStation),
			"pcba":         strings.TrimSpace(tsr.LogisticData.PCBANumber),
			"product_sn":   strings.TrimSpace(tsr.LogisticData.ProductSN),
			"finished_at":  strings.TrimSpace(tsr.TestFinishedTime),
			"all_passed":   tsr.IsAllPassed,
			"error_codes":  tsr.ErrorCodes,
		})
	}
	stepArrays := make([]map[string]interface{}, 0, len(group.TestSteps))
	for idx, steps := range group.TestSteps {
		inferredType, stepPCBA := parser.InferStationTypeFromSteps(steps)
		stepArrays = append(stepArrays, map[string]interface{}{
			"index":         idx,
			"inferred_type": inferredType,
			"scan_pcba":     stepPCBA,
			"step_count":    len(steps),
		})
	}
	return map[string]interface{}{
		"pcba":                 groupKey(group),
		"has_download":         group.DownloadInfo != (dto.DownloadInfoDTO{}),
		"download_tcu_pcba":    strings.TrimSpace(group.DownloadInfo.TcuPCBANumber),
		"station_record_count": len(group.TestStationRecords),
		"step_array_count":     len(group.TestSteps),
		"stations":             stations,
		"step_arrays":          stepArrays,
	}
}

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
// Phase 0 hotfix semantics: a failure inside a single group is logged and
// skipped — the rest of the file continues processing. A structural
// "more step arrays than station records" situation is no longer fatal:
// we insert what we can pair and log WARN about the excess. The method
// only returns a non-nil error if *every* group failed (catastrophic
// scenario — e.g. DB unreachable).
//
// See interface documentation for full details.
func (s *dispatcherService) DispatchGroups(ctx context.Context, groups []dto.GroupedDataDTO) error {
	logger.Info("Dispatch starting",
		logger.WithFields(map[string]interface{}{
			"group_count": len(groups),
		}),
	)

	var (
		groupsOK           int
		groupsFailed       int
		groupsWithExcess   int // processed OK but had unmatched step arrays
		groupsMismatchType int // processed OK but had step/station type mismatches
	)

	for _, group := range groups {
		result := s.dispatchSingleGroup(ctx, group)
		if result.err != nil {
			groupsFailed++
			logger.Warn("Group dispatch failed — skipping group, continuing with next",
				logger.WithFields(map[string]interface{}{
					"pcba":           groupKey(group),
					"error":          result.err.Error(),
					"stage":          result.failedStage,
					"group_snapshot": describeGroup(group),
					"note":           "Phase 0 hotfix: single-group errors are non-fatal. File processing continues.",
				}),
			)
			continue
		}
		groupsOK++
		if result.unmatchedStepArrays > 0 {
			groupsWithExcess++
		}
		if result.typeMismatches > 0 {
			groupsMismatchType++
		}
	}

	logger.Info("Dispatch finished",
		logger.WithFields(map[string]interface{}{
			"group_count":          len(groups),
			"groups_ok":            groupsOK,
			"groups_failed":        groupsFailed,
			"groups_with_excess":   groupsWithExcess,
			"groups_mismatch_type": groupsMismatchType,
		}),
	)

	if len(groups) > 0 && groupsOK == 0 {
		return fmt.Errorf("all %d groups failed to dispatch; see WARN events for details", len(groups))
	}
	return nil
}

// groupDispatchResult summarises the outcome of dispatching one group.
// err == nil means the group was processed (possibly partially — see
// unmatchedStepArrays). A non-nil err means the group was abandoned mid-way
// and the caller should log+continue.
type groupDispatchResult struct {
	err                 error
	failedStage         string // "download" | "logistic" | "station" | "steps"
	unmatchedStepArrays int    // step arrays we couldn't pair with a station (excess)
	typeMismatches      int    // paired but station_type != inferred step type
}

// dispatchSingleGroup inserts all rows for one grouped PCBA. Returns the
// outcome via groupDispatchResult. Errors from DB operations are returned;
// structural "more step arrays than station records" is logged as WARN and
// reported via unmatchedStepArrays (not an error).
func (s *dispatcherService) dispatchSingleGroup(ctx context.Context, group dto.GroupedDataDTO) groupDispatchResult {
	key := groupKey(group)

	logger.Debug("Dispatching group",
		logger.WithFields(map[string]interface{}{
			"pcba":                 key,
			"has_download":         group.DownloadInfo != (dto.DownloadInfoDTO{}),
			"station_record_count": len(group.TestStationRecords),
			"step_array_count":     len(group.TestSteps),
		}),
	)

	if (group.DownloadInfo != dto.DownloadInfoDTO{}) {
		if err := s.downloadInfoService.InsertDownloadInfo(ctx, group.DownloadInfo); err != nil {
			logger.Error("Failed to insert DownloadInfo",
				err,
				logger.WithFields(map[string]interface{}{
					"pcba":     key,
					"tcu_pcba": group.DownloadInfo.TcuPCBANumber,
				}),
			)
			return groupDispatchResult{
				err:         fmt.Errorf("failed to insert DownloadInfo for PCBA %s: %w", key, err),
				failedStage: "download",
			}
		}
	}

	testStationIDs := make([]int, 0, len(group.TestStationRecords))
	for _, tsr := range group.TestStationRecords {
		if (tsr.LogisticData == dto.LogisticDataDTO{}) {
			logger.Error("Station record has no LogisticData",
				logger.WithFields(map[string]interface{}{
					"station_type": strings.TrimSpace(tsr.TestStation),
					"group_key":    key,
				}),
			)
			return groupDispatchResult{
				err:         fmt.Errorf("missing LogisticData for PCBA %s: cannot insert TestStationRecord", key),
				failedStage: "logistic",
			}
		}

		logisticDataID, err := s.logisticDataService.GetOrInsertLogisticData(ctx, tsr.LogisticData)
		if err != nil {
			logger.Error("Failed to insert LogisticData",
				err,
				logger.WithFields(map[string]interface{}{
					"pcba":             strings.TrimSpace(tsr.LogisticData.PCBANumber),
					"product_sn":       strings.TrimSpace(tsr.LogisticData.ProductSN),
					"station_type":     strings.TrimSpace(tsr.TestStation),
					"logistic_data_id": logisticDataID,
					"group_key":        key,
				}),
			)
			return groupDispatchResult{
				err:         fmt.Errorf("failed to insert LogisticData for PCBA %s: %w", key, err),
				failedStage: "logistic",
			}
		}
		if logisticDataID == 0 {
			logger.Error("Resolved LogisticDataID is 0",
				logger.WithFields(map[string]interface{}{
					"pcba":         strings.TrimSpace(tsr.LogisticData.PCBANumber),
					"product_sn":   strings.TrimSpace(tsr.LogisticData.ProductSN),
					"station_type": strings.TrimSpace(tsr.TestStation),
					"group_key":    key,
					"dto_snapshot": tsr.LogisticData,
				}),
			)
			return groupDispatchResult{
				err:         fmt.Errorf("resolved LogisticDataID is 0 for PCBA %s", key),
				failedStage: "logistic",
			}
		}

		testStationID, err := s.testStationService.InsertTestStationRecord(ctx, tsr, logisticDataID)
		if err != nil {
			logger.Error("Failed to insert TestStationRecord",
				err,
				logger.WithFields(map[string]interface{}{
					"pcba":             strings.TrimSpace(tsr.LogisticData.PCBANumber),
					"station_type":     strings.TrimSpace(tsr.TestStation),
					"logistic_data_id": logisticDataID,
					"group_key":        key,
				}),
			)
			return groupDispatchResult{
				err:         fmt.Errorf("failed to insert TestStationRecord for PCBA %s: %w", key, err),
				failedStage: "station",
			}
		}
		testStationIDs = append(testStationIDs, testStationID)
	}

	result := groupDispatchResult{}

	for i, stepsSlice := range group.TestSteps {
		if i >= len(testStationIDs) {
			// Phase 0: no longer fatal. Log once with full context and skip
			// the unmatched step arrays. Downstream fix will pair them
			// properly by (PCBA, StationType).
			if result.unmatchedStepArrays == 0 {
				inferredType, stepPCBA := parser.InferStationTypeFromSteps(stepsSlice)
				logger.Warn("Skipping unmatched step array(s) — more step arrays than station records",
					logger.WithFields(map[string]interface{}{
						"pcba":                      key,
						"station_record_count":      len(testStationIDs),
						"step_array_count":          len(group.TestSteps),
						"first_unmatched_index":     i,
						"first_unmatched_inferred":  inferredType,
						"first_unmatched_scan_pcba": stepPCBA,
						"first_unmatched_steps":     len(stepsSlice),
						"group_snapshot":            describeGroup(group),
						"note":                      "Phase 0 hotfix: excess step arrays are skipped, not causing a fatal error. Root cause is either Bug #1 (missing StationInformation of inferred type) or retry asymmetry — see docs/problem.md.",
					}),
				)
			}
			result.unmatchedStepArrays = len(group.TestSteps) - len(testStationIDs)
			break
		}

		inferredType, _ := parser.InferStationTypeFromSteps(stepsSlice)
		pairedType := strings.TrimSpace(group.TestStationRecords[i].TestStation)
		if inferredType != "" && pairedType != "" && inferredType != pairedType {
			result.typeMismatches++
			logger.Warn("Dispatcher paired step array to station of different type",
				logger.WithFields(map[string]interface{}{
					"pcba":          key,
					"array_index":   i,
					"inferred_type": inferredType,
					"paired_type":   pairedType,
					"step_count":    len(stepsSlice),
					"note":          "steps will be inserted under a station of mismatched type. Data is still stored but semantically incorrect until the structural fix (Phase 1) lands.",
				}),
			)
		}
		if err := s.testStepService.InsertTestSteps(ctx, stepsSlice, testStationIDs[i]); err != nil {
			logger.Error("Failed to insert TestSteps",
				err,
				logger.WithFields(map[string]interface{}{
					"pcba":                   key,
					"test_station_record_id": testStationIDs[i],
					"step_count":             len(stepsSlice),
				}),
			)
			result.err = fmt.Errorf("failed to insert TestSteps for TestStationRecordID %d: %w", testStationIDs[i], err)
			result.failedStage = "steps"
			return result
		}
	}

	return result
}
