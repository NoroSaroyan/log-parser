package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log-parser/internal/domain/models/dto"
	"strings"
)

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

				// Start writing from the opening brace/bracket
				buffer.WriteString(strippedLine[openBraceIdx:])
				buffer.WriteByte('\n')

				// Count braces and brackets from that position
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

			// Close block only when both counts balanced
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

// FilterRelevantJsonBlocks and toJSON unchanged...
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
			//println("Accepted triple structure block")
			filtered = append(filtered, block)
			continue
		}

		var d dto.DownloadInfoDTO
		if json.Unmarshal([]byte(block), &d) == nil && d.TestStation != "" {
			//println("Accepted DownloadInfo block")
			filtered = append(filtered, block)
			continue
		}

		var tsr dto.TestStationRecordDTO
		if json.Unmarshal([]byte(block), &tsr) == nil && tsr.TestStation != "" {
			//println("Accepted TestStationRecord block")
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

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
