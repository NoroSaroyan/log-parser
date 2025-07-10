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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log-parser/internal/domain/models/dto"
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
	scanner := bufio.NewScanner(strings.NewReader(logs))

	var blocks []string
	var buffer bytes.Buffer
	var insideBlock bool
	var braceCount int
	var bracketCount int

	const prefixMarker = "]:"

	for scanner.Scan() {
		line := scanner.Text()

		prefixEnd := strings.Index(line, prefixMarker)
		if prefixEnd != -1 {
			prefixEnd += len(prefixMarker)
		} else {
			prefixEnd = 0
		}

		strippedLine := ""
		if len(line) > prefixEnd {
			strippedLine = line[prefixEnd:]
		}

		if !insideBlock {
			openBraceIdx := strings.IndexAny(strippedLine, "{[")
			if openBraceIdx != -1 {
				insideBlock = true
				buffer.Reset()
				braceCount = 0
				bracketCount = 0

				buffer.WriteString(strippedLine[openBraceIdx:])
				buffer.WriteByte('\n')

				substr := strippedLine[openBraceIdx:]
				braceCount += strings.Count(substr, "{")
				braceCount -= strings.Count(substr, "}")
				bracketCount += strings.Count(substr, "[")
				bracketCount -= strings.Count(substr, "]")
			}
		} else {
			buffer.WriteString(strippedLine)
			buffer.WriteByte('\n')

			braceCount += strings.Count(strippedLine, "{")
			braceCount -= strings.Count(strippedLine, "}")
			bracketCount += strings.Count(strippedLine, "[")
			bracketCount -= strings.Count(strippedLine, "]")

			if braceCount == 0 && bracketCount == 0 {
				blocks = append(blocks, buffer.String())
				insideBlock = false
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
			println("Accepted TestStep array block")
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
