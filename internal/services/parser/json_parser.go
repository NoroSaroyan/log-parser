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
	// stationTypesSeen tracks every station type we saw for a given PCBA, so the
	// step-array check can ask "was there a station of the *inferred* type for
	// this PCBA" (the real Bug #1 question), not just "what was the last record
	// in the map" (which is noisy when a device legitimately has both PCBA and
	// Final stations).
	stationTypesSeen := make(map[string]map[string]bool)
	var testStepsToProcess []struct {
		steps []dto.TestStepDTO
		index int
	}

	// Per-file counters for end-of-parse diagnostic summary.
	var (
		cntDownload              int
		cntStationPCBA           int
		cntStationFinal          int
		cntStationEmptyPCBA      int // station parsed but PCBANumber empty (fell back to ProductSN downstream)
		cntStepArrays            int
		cntStepsMatchedSameType  int // steps matched a station of the same inferred type  — good
		cntStepsMatchedDiffType  int // steps matched a station of a DIFFERENT type  — Bug #1 signature
		cntStepsOrphan           int // scan PCBA present but no station in allStations
		cntStepsNoScan           int // no PCBA Scan / Compare / Valid step found
		cntStepsUnknownInfer     int // scan step found but no recognized type — shouldn't happen
	)

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
					if _, dup := allStations[pcbaNum]; dup {
						logger.Debug("Duplicate station record for PCBA (overwriting in lookup map)",
							logger.WithFields(map[string]interface{}{
								"object_index":      i,
								"pcba":              pcbaNum,
								"new_station_type":  record.TestStation,
								"reason":            "allStations map keyed only by PCBA; second record replaces the first. Both still flow into results.",
							}),
						)
					}
					allStations[pcbaNum] = record
					if stationTypesSeen[pcbaNum] == nil {
						stationTypesSeen[pcbaNum] = map[string]bool{}
					}
					stationTypesSeen[pcbaNum][strings.TrimSpace(record.TestStation)] = true
				} else {
					cntStationEmptyPCBA++
					logger.Debug("Station record has empty PCBANumber",
						logger.WithFields(map[string]interface{}{
							"object_index":  i,
							"station_type":  record.TestStation,
							"product_sn":    strings.TrimSpace(record.LogisticData.ProductSN),
							"error_codes":   record.ErrorCodes,
							"reason":        "PCBANumber empty (likely early-failure device). Not indexed for step matching. Downstream grouper falls back to ProductSN.",
						}),
					)
				}
				if ts == "PCBA" {
					cntStationPCBA++
				} else {
					cntStationFinal++
				}
				results = append(results, record)

			case "Download":
				var download dto.DownloadInfoDTO
				if err := json.Unmarshal(raw, &download); err != nil {
					//fmt.Printf("Failed to parse DownloadInfoDTO at index %d: %v\n", i, err)
					continue
				}
				cntDownload++
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
		cntStepArrays++

		inferredType, pcbaFromSteps := InferStationTypeFromSteps(steps)

		if pcbaFromSteps != "" && inferredType != "" {
			typesForPCBA := stationTypesSeen[pcbaFromSteps]
			if typesForPCBA[inferredType] {
				// A station of the correct type exists — normal case.
				cntStepsMatchedSameType++
				logger.Debug("Test steps matched station of same type",
					logger.WithFields(map[string]interface{}{
						"array_index":  i,
						"pcba":         pcbaFromSteps,
						"station_type": inferredType,
						"step_count":   len(steps),
					}),
				)
				results = append(results, steps)
			} else if len(typesForPCBA) > 0 {
				// A station exists for this PCBA, but NOT of the inferred type.
				// This is the Bug #1 scenario in its actual form: PCBA-stage
				// steps were uploaded but only a Final-stage station record
				// exists (or vice versa). The parser will still pass them
				// through (keying lookup by PCBA only), and the dispatcher will
				// later abort the whole file on the resulting size mismatch.
				cntStepsMatchedDiffType++
				seenTypes := make([]string, 0, len(typesForPCBA))
				for t := range typesForPCBA {
					seenTypes = append(seenTypes, t)
				}
				logger.Warn("Missing station record of inferred type (Bug #1 signature)",
					logger.WithFields(map[string]interface{}{
						"array_index":         i,
						"pcba":                pcbaFromSteps,
						"inferred_type":       inferredType,
						"station_types_seen":  seenTypes,
						"step_count":          len(steps),
						"reason":              "test steps of type '" + inferredType + "' arrived for this PCBA, but the station record of that type was not in the log. Only the listed station_types_seen were present.",
					}),
				)
				results = append(results, steps)
			} else {
				cntStepsOrphan++
				logger.Warn("Test steps cannot be matched to any test station",
					logger.WithFields(map[string]interface{}{
						"array_index":          i,
						"pcba_from_test_steps": pcbaFromSteps,
						"inferred_type":        inferredType,
						"step_count":           len(steps),
						"reason":               "PCBA number was extracted from test steps, but no matching TestStationRecord was found with this PCBA (record missing from log)",
						"debug_checklist": []string{
							"1. Verify test station record was parsed before test steps in JSON",
							"2. Check if test station record has empty PCBANumber (use ProductSN fallback)",
							"3. Verify PCBA numbers match exactly (no whitespace differences)",
							"4. Factory-side: station may have failed to POST /v1/stationinformation for this device",
						},
					}),
				)
			}
		} else if pcbaFromSteps == "" {
			cntStepsNoScan++
			logger.Debug("Test step array lacks PCBA identifier",
				logger.WithFields(map[string]interface{}{
					"array_index":          i,
					"step_count":           len(steps),
					"supported_pcba_steps": []string{"PCBA Scan", "Compare PCBA Serial Number", "Valid PCBA Serial Number"},
					"reason":               "No PCBA number found in test steps. Device likely failed before PCBA identification step was reached",
				}),
			)
		} else {
			cntStepsUnknownInfer++
			logger.Debug("Test step array has PCBA but classifier returned empty station type",
				logger.WithFields(map[string]interface{}{
					"array_index": i,
					"pcba":        pcbaFromSteps,
					"step_count":  len(steps),
				}),
			)
		}
	}

	logger.Info("Parser classification summary",
		logger.WithFields(map[string]interface{}{
			"download_payloads":               cntDownload,
			"station_pcba_payloads":           cntStationPCBA,
			"station_final_payloads":          cntStationFinal,
			"station_records_with_empty_pcba": cntStationEmptyPCBA,
			"step_arrays_total":               cntStepArrays,
			"steps_matched_same_type":         cntStepsMatchedSameType,
			"steps_matched_different_type":    cntStepsMatchedDiffType,
			"steps_orphan_no_station":         cntStepsOrphan,
			"steps_missing_pcba_scan":         cntStepsNoScan,
			"steps_unknown_infer":             cntStepsUnknownInfer,
			"note":                            "steps_matched_different_type > 0 means the log had steps of a type without a matching StationInformation record (Bug #1: parser still passes them through, dispatcher will later trip). steps_orphan_no_station > 0 means steps came without any station record at all.",
		}),
	)

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
