package parser

import (
	"encoding/json"
	"errors"
	"fmt"

	"log-parser/internal/domain/models/dto"
)

// ParseMixedJSONArray parses a top-level JSON array containing mixed objects:
// - Objects with "TestStation" field: either TestStationRecordDTO or DownloadInfoDTO
// - Arrays of TestStepDTO objects
func ParseMixedJSONArray(data []byte) ([]interface{}, error) {
	// Unmarshal top-level array as RawMessages
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return nil, fmt.Errorf("unmarshal top-level array: %w", err)
	}

	var results []interface{}

	for _, raw := range rawItems {
		// Detect if raw is an array or object
		rawTrim := trimSpaces(raw)
		if len(rawTrim) == 0 {
			continue // skip empty
		}

		switch rawTrim[0] {
		case '[':
			// It's an array of TestStepDTO
			var steps []dto.TestStepDTO
			if err := json.Unmarshal(raw, &steps); err != nil {
				return nil, fmt.Errorf("unmarshal TestStepDTO array: %w", err)
			}
			results = append(results, steps)

		case '{':
			// It's an object - detect by TestStation field
			var probe map[string]interface{}
			if err := json.Unmarshal(raw, &probe); err != nil {
				return nil, fmt.Errorf("unmarshal object probe: %w", err)
			}

			tsRaw, hasTS := probe["TestStation"]
			if !hasTS {
				return nil, errors.New("object missing TestStation field, cannot detect type")
			}

			ts, ok := tsRaw.(string)
			if !ok {
				return nil, errors.New("TestStation field is not a string")
			}

			switch ts {
			case "PCBA", "Final":
				var record dto.TestStationRecordDTO
				if err := json.Unmarshal(raw, &record); err != nil {
					return nil, fmt.Errorf("unmarshal TestStationRecordDTO: %w", err)
				}
				results = append(results, record)

			case "Download":
				var download dto.DownloadInfoDTO
				if err := json.Unmarshal(raw, &download); err != nil {
					return nil, fmt.Errorf("unmarshal DownloadInfoDTO: %w", err)
				}
				results = append(results, download)

			default:
				return nil, fmt.Errorf("unknown TestStation value: %q", ts)
			}

		default:
			return nil, fmt.Errorf("unexpected JSON token: %c", rawTrim[0])
		}
	}

	return results, nil
}

// trimSpaces returns raw JSON bytes with leading spaces removed, for quick first-byte check
func trimSpaces(raw json.RawMessage) []byte {
	for i, b := range raw {
		if b != ' ' && b != '\n' && b != '\t' && b != '\r' {
			return raw[i:]
		}
	}
	return nil
}
