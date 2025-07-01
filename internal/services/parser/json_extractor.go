package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log-parser/internal/domain/models"
	"strings"
)

func ExtractJson(logs string) ([]string, error) {
	scanner := bufio.NewScanner(strings.NewReader(logs))

	var blocks []string
	var buffer bytes.Buffer
	var insideBlock bool
	var braceCount int

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
			openBraceIdx := strings.Index(strippedLine, "{")
			if openBraceIdx != -1 {
				insideBlock = true
				buffer.Reset()
				braceCount = 0

				buffer.WriteString(strippedLine[openBraceIdx:])
				buffer.WriteByte('\n')

				braceCount += strings.Count(strippedLine[openBraceIdx:], "{")
				braceCount -= strings.Count(strippedLine[openBraceIdx:], "}")
			}
		} else {
			buffer.WriteString(strippedLine)
			buffer.WriteByte('\n')

			braceCount += strings.Count(strippedLine, "{")
			braceCount -= strings.Count(strippedLine, "}")

			if braceCount == 0 {
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

func FilterRelevantJsonBlocks(blocks []string) ([]string, error) {
	var filtered []string

	for _, block := range blocks {
		block = strings.TrimSpace(block)

		var fullStructure []interface{}
		err := json.Unmarshal([]byte(block), &fullStructure)
		if err == nil && len(fullStructure) == 3 {
			var d models.DownloadInfo
			if err := json.Unmarshal([]byte(toJSON(fullStructure[0])), &d); err != nil {
				continue
			}
			var steps []models.TestStep
			if err := json.Unmarshal([]byte(toJSON(fullStructure[1])), &steps); err != nil {
				continue
			}
			var tsr models.TestStationRecord
			if err := json.Unmarshal([]byte(toJSON(fullStructure[2])), &tsr); err != nil {
				continue
			}

			filtered = append(filtered, block)
			continue
		}

		var d models.DownloadInfo
		if json.Unmarshal([]byte(block), &d) == nil && d.TestStation != "" {
			filtered = append(filtered, block)
			continue
		}

		var tsr models.TestStationRecord
		if json.Unmarshal([]byte(block), &tsr) == nil && tsr.TestStation != "" {
			filtered = append(filtered, block)
			continue
		}

		var steps []models.TestStep
		if json.Unmarshal([]byte(block), &steps) == nil && len(steps) > 0 {
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
