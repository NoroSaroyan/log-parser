/*
Package parser contains utilities to parse and extract structured domain data
from mixed JSON arrays typically found in raw log files.

The core challenge addressed here is that JSON arrays may contain heterogeneous
elements of varying types, representing different domain concepts (e.g., test steps,
test station records, download info). This package helps parse such mixed content
and segregate it into meaningful domain DTOs.

Function:

  - ParseMixedJSONArray: Accepts raw JSON byte data representing an array of mixed
    elements. Each element may be an object or an array, corresponding to one of the
    domain DTO types. The function unmarshals each element based on its structure and
    domain-specific logic, returning a slice of parsed domain objects.
*/
package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
	"github.com/NoroSaroyan/log-parser/internal/infrastructure/logger"
)

// ParseMixedJSONArray parses a raw JSON byte array representing a mixed-type JSON array.
//
// The input JSON is expected to be a top-level array containing elements of differing types:
// - Arrays of TestStepDTO objects (representing test steps)
// - Objects representing TestStationRecordDTO or DownloadInfoDTO
//
// Parsing logic:
//   - Iterates each element, inspects the first non-whitespace byte to distinguish arrays vs objects.
//   - For arrays ('['), attempts to parse as []TestStepDTO, then checks for a PCBA identifier
//     in the steps by looking for "PCBA Scan", "Compare PCBA Serial Number", or "Valid PCBA Serial Number"
//     steps to filter only test steps corresponding to a known TestStation record.
//   - For objects ('{'), parses into a generic map first to check the 'TestStation' field:
//   - If 'TestStation' is "PCBA" or "Final", attempts to parse as TestStationRecordDTO.
//   - All TestStationRecords are indexed by their PCBANumber to correlate with test steps.
//   - If 'TestStation' is "Download", parses as DownloadInfoDTO.
//   - Ignores or logs any elements with unexpected structures or missing fields.
//
// Returns a slice of interface{} containing parsed domain DTOs (TestStationRecordDTO,
// DownloadInfoDTO, []TestStepDTO) filtered and grouped according to domain rules.
//
// Errors during unmarshaling individual elements are logged and do not halt parsing of
// the entire array.
//
// This function enables the application to reliably parse complex mixed JSON arrays from logs,
// correctly associating test steps with their corresponding test stations.
func ParseMixedJSONArray(data []byte) ([]interface{}, error) {
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return nil, fmt.Errorf("unmarshal top-level array: %w", err)
	}

	var results []interface{}
	allStations := make(map[string]dto.TestStationRecordDTO)
	var testStepsToProcess []struct {
		steps []dto.TestStepDTO
		index int
	}

	// First pass: collect all TestStationRecords and DownloadInfo
	for i, raw := range rawItems {
		rawTrim := trimSpaces(raw)
		if len(rawTrim) == 0 {
			continue
		}

		switch rawTrim[0] {
		case '[':
			var steps []dto.TestStepDTO
			if err := json.Unmarshal(raw, &steps); err != nil {
				logger.Debug("Failed to unmarshal test step array",
				logger.WithFields(map[string]interface{}{
					"array_index": i,
					"error":       err.Error(),
					"reason":      "This array will not be processed. The JSON structure may be malformed or not a valid test step array",
				}),
			)
				continue
			}

			// Store TestSteps for second pass processing
			testStepsToProcess = append(testStepsToProcess, struct {
				steps []dto.TestStepDTO
				index int
			}{steps, i})

		case '{':
			var probe map[string]interface{}
			if err := json.Unmarshal(raw, &probe); err != nil {
				logger.Debug("Failed to probe JSON object structure",
				logger.WithFields(map[string]interface{}{
					"object_index": i,
					"error":        err.Error(),
					"reason":       "Could not parse this JSON object to determine its type (TestStation, Download, etc.)",
				}),
			)
				continue
			}

			tsRaw, hasTS := probe["TestStation"]
			if !hasTS {
				logger.Debug("JSON object missing TestStation field",
				logger.WithFields(map[string]interface{}{
					"object_index": i,
					"reason":       "Expected JSON object to have 'TestStation' field identifying it as PCBA, Final, or Download test data",
				}),
			)
				continue
			}

			ts, ok := tsRaw.(string)
			if !ok {
				logger.Debug("TestStation field has invalid type",
				logger.WithFields(map[string]interface{}{
					"object_index": i,
					"reason":       "TestStation field must be a string, but got a different type. Cannot determine record type",
				}),
			)
				continue
			}

			switch ts {
			case "PCBA", "Final":
				var record dto.TestStationRecordDTO
				if err := json.Unmarshal(raw, &record); err != nil {
					logger.Debug("Failed to parse TestStationRecord",
					logger.WithFields(map[string]interface{}{
						"object_index": i,
						"test_station": ts,
						"error":        err.Error(),
						"reason":       "TestStationRecord object structure does not match expected DTO. See JSON structure at this index in the log file",
					}),
				)
					continue
				}
				//fmt.Printf("Parsed %s TestStationRecord at index %d: PCBANumber=%s, PartNumber=%s\n",
				//	record.TestStation, i, record.LogisticData.PCBANumber, record.PartNumber)
				pcbaNum := strings.TrimSpace(record.LogisticData.PCBANumber)
				if pcbaNum != "" {
					allStations[pcbaNum] = record
				}
				results = append(results, record)

			case "Download":
				var download dto.DownloadInfoDTO
				if err := json.Unmarshal(raw, &download); err != nil {
					//fmt.Printf("Failed to parse DownloadInfoDTO at index %d: %v\n", i, err)
					continue
				}
				results = append(results, download)

			default:
				logger.Debug("Unknown TestStation type encountered",
				logger.WithFields(map[string]interface{}{
					"object_index":   i,
					"test_station":   ts,
					"supported_types": []string{"PCBA", "Final", "Download"},
					"reason":         "TestStation has an unexpected value. Supported types are: PCBA, Final, Download",
				}),
			)
				continue
			}

		default:
			logger.Debug("Unexpected JSON element type",
			logger.WithFields(map[string]interface{}{
				"element_index": i,
				"first_token":   string(rawTrim[0]),
				"reason":        "Expected JSON array '[' or object '{', but got different token. This element will be skipped",
			}),
		)
			continue
		}
	}

	for _, testStepData := range testStepsToProcess {
		steps := testStepData.steps
		i := testStepData.index

		var pcbaFromSteps string
		for _, s := range steps {
			if s.TestStepName == "Compare PCBA Serial Number" || s.TestStepName == "PCBA Scan" || s.TestStepName == "Valid PCBA Serial Number" {
				pcbaFromSteps = strings.TrimSpace(s.GetMeasuredValueString())
				break
			}
		}

		if pcbaFromSteps != "" {
			if _, ok := allStations[pcbaFromSteps]; ok {
				//fmt.Printf("Matched test steps for PCBA: %s (%s station) at index %d (%d steps)\n",
				//	pcbaFromSteps, station.TestStation, i, len(steps))
				results = append(results, steps)
			} else {
				logger.Warn("Test steps cannot be matched to any test station",
				logger.WithFields(map[string]interface{}{
					"array_index":             i,
					"pcba_from_test_steps":    pcbaFromSteps,
					"step_count":              len(steps),
					"reason":                  "PCBA number was extracted from test steps, but no matching TestStationRecord was found with this PCBA",
					"debug_checklist":         []string{"1. Verify test station record was parsed before test steps in JSON", "2. Check if test station record has empty PCBANumber (use ProductSN fallback)", "3. Verify PCBA numbers match exactly (no whitespace differences)"},
				}),
			)
			}
		} else {
			logger.Debug("Test step array lacks PCBA identifier",
			logger.WithFields(map[string]interface{}{
				"array_index":            i,
				"step_count":             len(steps),
				"supported_pcba_steps":   []string{"PCBA Scan", "Compare PCBA Serial Number", "Valid PCBA Serial Number", "Write PCBA Serial Number"},
				"reason":                 "No PCBA number found in test steps. Device likely failed before PCBA identification step was reached",
				"debug_checklist":         []string{"1. Check if array contains: 'PCBA Scan', 'Compare PCBA Serial Number', or 'Valid PCBA Serial Number' steps", "2. If none found, device probably failed early in test sequence", "3. These steps should have TestMeasuredValue containing the PCBA number"},
			}),
		)
		}
	}

	return results, nil
}

// trimSpaces returns a subslice of raw JSON bytes with leading whitespace characters removed.
//
// Used internally to identify the first meaningful byte token in a JSON element,
// helping determine its type (array or object).
func trimSpaces(raw json.RawMessage) []byte {
	for i, b := range raw {
		if b != ' ' && b != '\n' && b != '\t' && b != '\r' {
			return raw[i:]
		}
	}
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
