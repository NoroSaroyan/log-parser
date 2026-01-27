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
	for _, g := range groups {
		result = append(result, *g)
	}

	return result, nil
}
