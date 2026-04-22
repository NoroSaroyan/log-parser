/*
Package processor provides services for processing and organizing parsed domain data.

Function:

  - GroupByPCBANumber: Takes a heterogeneous slice of parsed domain objects (resulting from
    JSON parsing of logs) and groups them by their PCBANumber, aggregating related DTOs into
    a unified structure for downstream processing or database insertion.

GroupByPCBANumber organizes parsed domain entities into logical groups keyed by the PCBANumber,
which serves as the primary identifier linking DownloadInfoDTO, TestStationRecordDTO, and
TestStepDTO data that belong together.

Parsing outputs typically consist of a slice of interface{} containing:
- DownloadInfoDTO objects,
- TestStationRecordDTO objects,
- Arrays of TestStepDTO objects.

The function expects these types and will return an error if unexpected types are encountered.

Grouping logic:
  - For DownloadInfoDTO, uses the TcuPCBANumber field as the group key.
  - For TestStationRecordDTO, uses LogisticData.PCBANumber as the key. If PCBANumber is empty,
    falls back to LogisticData.ProductSN as the grouping key.
  - For arrays of TestStepDTO, attempts to find a PCBA identifier by inspecting each step's
    TestStepName for "PCBA Scan", "Compare PCBA Serial Number", or "Valid PCBA Serial Number",
    using the corresponding measured value as the key.

Each group (dto.GroupedDataDTO) contains exactly one DownloadInfoDTO, one or more TestStationRecordDTOs,
and one or more slices of TestStepDTO arrays that share the same PCBANumber (or ProductSN if PCBA is empty).

If any required key is missing (both PCBANumber and ProductSN empty for TestStationRecordDTO, or no PCBA
identifier found in TestStepDTO array), the function returns an error describing the issue.

The resulting slice of GroupedDataDTO structs is unordered but contains all groups with
their aggregated related data, ready for further processing such as validation, database
insertion, or business logic execution.

This function is critical for assembling the parsed log data into coherent domain aggregates,
ensuring data integrity by grouping related pieces before downstream workflows.
*/
package processor

import (
	"fmt"
	"strings"

	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
	"github.com/NoroSaroyan/log-parser/internal/infrastructure/logger"
)

func GroupByPCBANumber(parsed []interface{}) ([]dto.GroupedDataDTO, error) {
	groups := map[string]*dto.GroupedDataDTO{}
	for _, item := range parsed {
		switch v := item.(type) {
		case dto.DownloadInfoDTO:
			key := strings.TrimSpace(v.TcuPCBANumber)
			if key == "" {
				// Skip DownloadInfo records with empty TcuPCBANumber - they are optional
				logger.Debug("Skipping DownloadInfo with empty TcuPCBANumber",
					logger.WithFields(map[string]interface{}{
						"reason": "DownloadInfo requires a TCU PCBA number for proper grouping and database insertion",
					}),
				)
				continue
			}
			group, ok := groups[key]
			if !ok {
				group = &dto.GroupedDataDTO{}
				groups[key] = group
			}
			group.DownloadInfo = v

		case dto.TestStationRecordDTO:
			key := strings.TrimSpace(v.LogisticData.PCBANumber)
			// Fallback to ProductSN if PCBANumber is empty
			if key == "" {
				key = strings.TrimSpace(v.LogisticData.ProductSN)
				if key == "" {
					return nil, fmt.Errorf("TestStationRecordDTO missing both LogisticData.PCBANumber and ProductSN")
				}
			}
			group, ok := groups[key]
			if !ok {
				group = &dto.GroupedDataDTO{}
				groups[key] = group
			}
			group.TestStationRecords = append(group.TestStationRecords, v)

		case []dto.TestStepDTO:
			var key string
			for _, step := range v {
				if step.TestStepName == "PCBA Scan" || step.TestStepName == "Compare PCBA Serial Number" || step.TestStepName == "Valid PCBA Serial Number" {
					key = strings.TrimSpace(step.GetMeasuredValueString())
					break
				}
			}

			if key == "" {
				return nil, fmt.Errorf("TestStepDTO array missing PCBA Scan step with measured value")
			}
			group, ok := groups[key]
			if !ok {
				group = &dto.GroupedDataDTO{}
				groups[key] = group
			}
			group.TestSteps = append(group.TestSteps, v)

		default:
			return nil, fmt.Errorf("unexpected type in parsed data: %T", v)
		}
	}

	result := make([]dto.GroupedDataDTO, 0, len(groups))
	// Diagnostic counters for the grouping phase.
	var (
		groupsWithDownload int
		groupsWithAnyStation int
		groupsWithAnySteps int
		stationsByType = map[string]int{}
		stepsByType    = map[string]int{}
		mismatchGroups int // groups where step count > station record count (pre-dispatch)
	)
	for key, g := range groups {
		if (g.DownloadInfo != dto.DownloadInfoDTO{}) {
			groupsWithDownload++
		}
		if len(g.TestStationRecords) > 0 {
			groupsWithAnyStation++
			for _, tsr := range g.TestStationRecords {
				stationsByType[strings.TrimSpace(tsr.TestStation)]++
			}
		}
		if len(g.TestSteps) > 0 {
			groupsWithAnySteps++
			for _, steps := range g.TestSteps {
				// Inline classifier (can't import parser here — cycle).
				var t string
				for _, s := range steps {
					switch s.TestStepName {
					case "PCBA Scan":
						t = "PCBA"
					case "Compare PCBA Serial Number", "Valid PCBA Serial Number":
						t = "Final"
					}
					if t != "" {
						break
					}
				}
				if t == "" {
					t = "unknown"
				}
				stepsByType[t]++
			}
		}
		if len(g.TestSteps) > len(g.TestStationRecords) {
			mismatchGroups++
			// Classify the mismatch cause: is it a missing station (Bug #1) or
			// a retry-asymmetry (same type, unequal counts)?
			stationsCountByType := map[string]int{}
			for _, tsr := range g.TestStationRecords {
				stationsCountByType[strings.TrimSpace(tsr.TestStation)]++
			}
			stepsCountByType := map[string]int{}
			for _, steps := range g.TestSteps {
				var t string
				for _, s := range steps {
					switch s.TestStepName {
					case "PCBA Scan":
						t = "PCBA"
					case "Compare PCBA Serial Number", "Valid PCBA Serial Number":
						t = "Final"
					}
					if t != "" {
						break
					}
				}
				if t == "" {
					t = "unknown"
				}
				stepsCountByType[t]++
			}
			// Determine the category
			var cause string
			missing := []string{}
			asymmetric := []string{}
			for t, stepsN := range stepsCountByType {
				stationsN := stationsCountByType[t]
				if stationsN == 0 {
					missing = append(missing, t)
				} else if stepsN > stationsN {
					asymmetric = append(asymmetric, t)
				}
			}
			if len(missing) > 0 && len(asymmetric) == 0 {
				cause = "bug1_missing_station_record_of_type"
			} else if len(asymmetric) > 0 && len(missing) == 0 {
				cause = "retry_asymmetry_same_type"
			} else {
				cause = "mixed"
			}
			logger.Warn("Group has more step arrays than station records (will fail in dispatcher)",
				logger.WithFields(map[string]interface{}{
					"group_key":              key,
					"station_record_count":   len(g.TestStationRecords),
					"step_array_count":       len(g.TestSteps),
					"has_download":           g.DownloadInfo != (dto.DownloadInfoDTO{}),
					"stations_by_type":       stationsCountByType,
					"steps_by_type":          stepsCountByType,
					"cause":                  cause,
					"missing_station_types":  missing,
					"asymmetric_types":       asymmetric,
				}),
			)
		}
		result = append(result, *g)
	}

	logger.Info("Grouping summary",
		logger.WithFields(map[string]interface{}{
			"groups_total":          len(groups),
			"groups_with_download":  groupsWithDownload,
			"groups_with_station":   groupsWithAnyStation,
			"groups_with_steps":     groupsWithAnySteps,
			"stations_by_type":      stationsByType,
			"step_arrays_by_type":   stepsByType,
			"groups_with_mismatch":  mismatchGroups,
		}),
	)

	return result, nil
}
