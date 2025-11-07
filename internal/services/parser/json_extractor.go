/*
Package parser provides functionality for extracting and filtering JSON data blocks
from raw log files. This package is designed to handle log files where JSON data
is embedded within noisy, prefixed, or multi-line textual log entries.

The primary goal is to reliably extract JSON fragments that represent meaningful
data structures, such as DownloadInfo, TestStationRecords, and TestSteps,
which are critical to downstream processing.

Functions:

  - ExtractJson: Scans a raw log string, identifies and extracts well-formed JSON blocks
    by matching balanced braces/brackets, even if spanning multiple lines.

  - FilterRelevantJsonBlocks: Filters extracted JSON blocks, returning only those
    which can successfully unmarshal into the known domain data structures, thus
    identifying blocks relevant to the application’s domain logic.
*/
package parser

import (
	"encoding/json"
	"errors"
	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
	"strings"
)

// ExtractJson scans the provided raw log string line-by-line and extracts JSON blocks.
//
// It uses a simple state machine approach:
//   - Detects lines containing JSON opening delimiters '{' or '[' after a known prefix marker.
//   - Tracks balanced curly braces and square brackets to identify complete JSON structures,
//     including nested objects or arrays spanning multiple lines.
//
// The 'prefixMarker' (currently set to "]:") is stripped off to isolate JSON content.
//
// Returns a slice of JSON strings, each representing a complete JSON block extracted from the logs.
//
// If the input contains no JSON blocks or malformed blocks that can't be balanced, those are ignored.
// The function returns an error only if the input scanning encounters an I/O error.
func ExtractJson(logs string) ([]string, error) {
	lines := strings.Split(logs, "\n")
	var blocks []string
	var currentBlock strings.Builder
	var insideBlock bool
	var braceCount int

	for _, line := range lines {
		if strings.Contains(line, " Data  {") {
			// Start of a new Data block
			if insideBlock {
				// Previous block wasn't properly closed, save it anyway
				blocks = append(blocks, currentBlock.String())
			}

			insideBlock = true
			currentBlock.Reset()
			braceCount = 0

			// Extract everything after " Data  "
			dataIdx := strings.Index(line, " Data  ")
			if dataIdx != -1 {
				jsonPart := line[dataIdx+7:] // 7 = len(" Data  ")
				currentBlock.WriteString(jsonPart)
				currentBlock.WriteByte('\n')
				braceCount += strings.Count(jsonPart, "{") - strings.Count(jsonPart, "}")
			}
		} else if strings.Contains(line, " Data  [") {
			// Start of a new Data array block
			if insideBlock {
				// Previous block wasn't properly closed, save it anyway
				blocks = append(blocks, currentBlock.String())
			}

			insideBlock = true
			currentBlock.Reset()
			braceCount = 0

			// Extract everything after " Data  "
			dataIdx := strings.Index(line, " Data  ")
			if dataIdx != -1 {
				jsonPart := line[dataIdx+7:] // 7 = len(" Data  ")
				currentBlock.WriteString(jsonPart)
				currentBlock.WriteByte('\n')
				braceCount += strings.Count(jsonPart, "[") - strings.Count(jsonPart, "]")
			}
		} else if insideBlock {
			// Continue building the current block
			// Extract JSON content after the log prefix
			prefixEnd := strings.Index(line, "]:")
			if prefixEnd != -1 && len(line) > prefixEnd+2 {
				jsonPart := line[prefixEnd+2:]
				currentBlock.WriteString(jsonPart)
				currentBlock.WriteByte('\n')
				braceCount += strings.Count(jsonPart, "{") - strings.Count(jsonPart, "}")
				braceCount += strings.Count(jsonPart, "[") - strings.Count(jsonPart, "]")
			}

			// Check if block is complete
			if braceCount <= 0 && currentBlock.Len() > 10 {
				blocks = append(blocks, currentBlock.String())
				insideBlock = false
				currentBlock.Reset()
			}
		}
	}

	// Handle case where file ends while inside a block
	if insideBlock && currentBlock.Len() > 10 {
		blocks = append(blocks, currentBlock.String())
	}

	return blocks, nil
}

// FilterRelevantJsonBlocks filters a list of JSON strings, returning only those blocks
// that are relevant to the application domain and successfully parse into one or more
// of the known DTO types.
//
// This function attempts to unmarshal each JSON block into the following types:
// - A compound array of exactly three elements: DownloadInfoDTO, []TestStepDTO, TestStationRecordDTO
// - Standalone DownloadInfoDTO with a non-empty TestStation field
// - Standalone TestStationRecordDTO with a non-empty TestStation field
// - A slice of TestStepDTO with one or more elements
//
// Blocks that successfully unmarshal to any of these valid structures are included in
// the returned slice.
//
// Returns an error if no relevant JSON blocks are found.
//
// This filtering ensures that only domain-significant JSON data is further processed,
// reducing noise from unrelated or malformed JSON blocks.
func FilterRelevantJsonBlocks(blocks []string) ([]string, error) {
	var filtered []string

	for _, block := range blocks {
		block = strings.TrimSpace(block)

		var fullStructure []interface{}
		err := json.Unmarshal([]byte(block), &fullStructure)
		if err == nil && len(fullStructure) == 3 {
			var d dto.DownloadInfoDTO
			if err := json.Unmarshal([]byte(toJSON(fullStructure[0])), &d); err != nil {
				continue
			}
			var steps []dto.TestStepDTO
			if err := json.Unmarshal([]byte(toJSON(fullStructure[1])), &steps); err != nil {
				continue
			}
			var tsr dto.TestStationRecordDTO
			if err := json.Unmarshal([]byte(toJSON(fullStructure[2])), &tsr); err != nil {
				continue
			}
			filtered = append(filtered, block)
			continue
		}

		var d dto.DownloadInfoDTO
		if json.Unmarshal([]byte(block), &d) == nil && d.TestStation != "" {
			filtered = append(filtered, block)
			continue
		}

		var tsr dto.TestStationRecordDTO
		if json.Unmarshal([]byte(block), &tsr) == nil && tsr.TestStation != "" {
			filtered = append(filtered, block)
			continue
		}

		var steps []dto.TestStepDTO
		if json.Unmarshal([]byte(block), &steps) == nil && len(steps) > 0 {
			//println("Accepted TestStep array block")
			filtered = append(filtered, block)
			continue
		}
	}

	if len(filtered) == 0 {
		return nil, errors.New("no relevant JSON blocks found")
	}

	return filtered, nil
}

// toJSON is a helper function that marshals a Go value back into JSON string format.
//
// Used internally to convert unmarshaled interface{} values back into JSON strings
// for nested unmarshaling attempts.
//
// Note: Errors during marshaling are ignored here for brevity as input is controlled.
func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
