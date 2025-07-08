package parser

import (
	"encoding/json"
	"fmt"

	"log-parser/internal/domain/models/dto"
)

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

func trimSpaces(raw json.RawMessage) []byte {
	for i, b := range raw {
		if b != ' ' && b != '\n' && b != '\t' && b != '\r' {
			return raw[i:]
		}
	}
	return nil
}
