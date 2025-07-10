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

	"log-parser/internal/domain/models/dto"
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
//     in the steps to filter only test steps corresponding to a known "Final" TestStation record.
//   - For objects ('{'), parses into a generic map first to check the 'TestStation' field:
//   - If 'TestStation' is "PCBA" or "Final", attempts to parse as TestStationRecordDTO.
//   - "Final" TestStationRecords are indexed by their PCBANumber to correlate with test steps.
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
	finalStations := make(map[string]dto.TestStationRecordDTO)

	for i, raw := range rawItems {
		rawTrim := trimSpaces(raw)
		if len(rawTrim) == 0 {
			continue
		}

		switch rawTrim[0] {
		case '[':
			var steps []dto.TestStepDTO
			if err := json.Unmarshal(raw, &steps); err != nil {
				fmt.Printf("Error unmarshaling TestStepDTO array at index %d: %v\n", i, err)
				continue
			}

			var pcbaFromSteps string
			for _, s := range steps {
				if s.TestStepName == "Compare PCBA Serial Number" || s.TestStepName == "PCBA Scan" {
					pcbaFromSteps = s.GetMeasuredValueString()
					break
				}
			}

			if pcbaFromSteps != "" {
				if _, ok := finalStations[pcbaFromSteps]; ok {
					fmt.Printf("Matched FINAL test steps for PCBA: %s at index %d (%d steps)\n", pcbaFromSteps, i, len(steps))
					results = append(results, steps)
				} else {
					fmt.Printf("Skipping test steps — no FINAL station match for PCBA: %s at index %d\n", pcbaFromSteps, i)
				}
			} else {
				fmt.Printf("No PCBA identifier in test steps at index %d\n", i)
			}

		case '{':
			var probe map[string]interface{}
			if err := json.Unmarshal(raw, &probe); err != nil {
				fmt.Printf("Error probing object at index %d: %v\n", i, err)
				continue
			}

			tsRaw, hasTS := probe["TestStation"]
			if !hasTS {
				fmt.Printf("Object at index %d missing 'TestStation' field\n", i)
				continue
			}

			ts, ok := tsRaw.(string)
			if !ok {
				fmt.Printf("'TestStation' field is not string at index %d\n", i)
				continue
			}

			switch ts {
			case "PCBA", "Final":
				var record dto.TestStationRecordDTO
				if err := json.Unmarshal(raw, &record); err != nil {
					fmt.Printf("Failed to parse TestStationRecordDTO at index %d: %v\n", i, err)
					continue
				}
				if record.TestStation == "Final" {
					fmt.Printf("Parsed FINAL TestStationRecord at index %d: PCBANumber=%s, PartNumber=%s\n",
						i, record.LogisticData.PCBANumber, record.PartNumber)
					finalStations[record.LogisticData.PCBANumber] = record
				}
				results = append(results, record)

			case "Download":
				var download dto.DownloadInfoDTO
				if err := json.Unmarshal(raw, &download); err != nil {
					fmt.Printf("Failed to parse DownloadInfoDTO at index %d: %v\n", i, err)
					continue
				}
				results = append(results, download)

			default:
				fmt.Printf("Unknown TestStation value %q at index %d\n", ts, i)
				continue
			}

		default:
			fmt.Printf("Unexpected JSON token at index %d: %c\n", i, rawTrim[0])
			continue
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
