package parser

import (
	"encoding/json"
	"testing"

	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
)

func TestParseMixedJSONArray(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
		wantErr     bool
		expectTypes []string // "TestStationRecord", "DownloadInfo", "TestSteps"
	}{
		{
			name: "single TestStationRecord PCBA",
			input: `[
				{
					"TestStation": "PCBA",
					"SerialNumber": "SN123",
					"PartNumber": "PN456",
					"LogisticData": {
						"PCBANumber": "PCBA001",
						"SerialNumber": "SN123"
					}
				}
			]`,
			expectedLen: 1,
			wantErr:     false,
			expectTypes: []string{"TestStationRecord"},
		},
		{
			name: "single TestStationRecord Final",
			input: `[
				{
					"TestStation": "Final",
					"SerialNumber": "SN123",
					"PartNumber": "PN456",
					"LogisticData": {
						"PCBANumber": "PCBA001",
						"SerialNumber": "SN123"
					}
				}
			]`,
			expectedLen: 1,
			wantErr:     false,
			expectTypes: []string{"TestStationRecord"},
		},
		{
			name: "single DownloadInfo",
			input: `[
				{
					"TestStation": "Download",
					"SerialNumber": "SN123",
					"PartNumber": "PN456"
				}
			]`,
			expectedLen: 1,
			wantErr:     false,
			expectTypes: []string{"DownloadInfo"},
		},
		{
			name: "TestSteps with matching PCBA",
			input: `[
				{
					"TestStation": "PCBA",
					"SerialNumber": "SN123",
					"LogisticData": {
						"PCBANumber": "PCBA001",
						"SerialNumber": "SN123"
					}
				},
				[
					{
						"TestStepName": "Compare PCBA Serial Number",
						"TestStepResult": "PASS",
						"TestMeasuredValue": "PCBA001"
					},
					{
						"TestStepName": "Voltage Test",
						"TestStepResult": "PASS",
						"TestMeasuredValue": "3.3V"
					}
				]
			]`,
			expectedLen: 2,
			wantErr:     false,
			expectTypes: []string{"TestStationRecord", "TestSteps"},
		},
		{
			name: "TestSteps with PCBA Scan identifier",
			input: `[
				{
					"TestStation": "PCBA",
					"SerialNumber": "SN123",
					"LogisticData": {
						"PCBANumber": "PCBA002",
						"SerialNumber": "SN123"
					}
				},
				[
					{
						"TestStepName": "PCBA Scan",
						"TestStepResult": "PASS",
						"TestMeasuredValue": "PCBA002"
					}
				]
			]`,
			expectedLen: 2,
			wantErr:     false,
			expectTypes: []string{"TestStationRecord", "TestSteps"},
		},
		{
			name: "TestSteps without matching PCBA",
			input: `[
				{
					"TestStation": "PCBA",
					"SerialNumber": "SN123",
					"LogisticData": {
						"PCBANumber": "PCBA001",
						"SerialNumber": "SN123"
					}
				},
				[
					{
						"TestStepName": "Compare PCBA Serial Number",
						"TestStepResult": "PASS",
						"TestMeasuredValue": "PCBA999"
					}
				]
			]`,
			expectedLen: 1,
			wantErr:     false,
			expectTypes: []string{"TestStationRecord"},
		},
		{
			name: "mixed types complete scenario",
			input: `[
				{
					"TestStation": "Download",
					"SerialNumber": "SN123",
					"PartNumber": "PN456"
				},
				{
					"TestStation": "PCBA",
					"SerialNumber": "SN123",
					"LogisticData": {
						"PCBANumber": "PCBA001",
						"SerialNumber": "SN123"
					}
				},
				[
					{
						"TestStepName": "Compare PCBA Serial Number",
						"TestStepResult": "PASS",
						"TestMeasuredValue": "PCBA001"
					}
				],
				{
					"TestStation": "Final",
					"SerialNumber": "SN123",
					"LogisticData": {
						"PCBANumber": "PCBA001",
						"SerialNumber": "SN123"
					}
				}
			]`,
			expectedLen: 4,
			wantErr:     false,
			expectTypes: []string{"DownloadInfo", "TestStationRecord", "TestStationRecord", "TestSteps"},
		},
		{
			name:        "invalid JSON",
			input:       `[{"TestStation": "PCBA", "SerialNumber":}]`,
			expectedLen: 0,
			wantErr:     true,
		},
		{
			name:        "empty array",
			input:       `[]`,
			expectedLen: 0,
			wantErr:     false,
		},
		{
			name: "unknown TestStation",
			input: `[
				{
					"TestStation": "Unknown",
					"SerialNumber": "SN123"
				}
			]`,
			expectedLen: 0,
			wantErr:     false,
		},
		{
			name: "object without TestStation field",
			input: `[
				{
					"SerialNumber": "SN123",
					"PartNumber": "PN456"
				}
			]`,
			expectedLen: 0,
			wantErr:     false,
		},
		{
			name: "TestSteps without PCBA identifier",
			input: `[
				{
					"TestStation": "PCBA",
					"SerialNumber": "SN123",
					"LogisticData": {
						"PCBANumber": "PCBA001",
						"SerialNumber": "SN123"
					}
				},
				[
					{
						"TestStepName": "Voltage Test",
						"TestStepResult": "PASS",
						"TestMeasuredValue": "3.3V"
					}
				]
			]`,
			expectedLen: 1,
			wantErr:     false,
			expectTypes: []string{"TestStationRecord"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMixedJSONArray([]byte(tt.input))
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMixedJSONArray() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(result) != tt.expectedLen {
				t.Errorf("ParseMixedJSONArray() returned %d items, expected %d", len(result), tt.expectedLen)
				return
			}
			
			// Verify types if specified
			if tt.expectTypes != nil && len(tt.expectTypes) == len(result) {
				for i, expectedType := range tt.expectTypes {
					switch expectedType {
					case "TestStationRecord":
						if _, ok := result[i].(dto.TestStationRecordDTO); !ok {
							t.Errorf("Item %d expected to be TestStationRecordDTO, got %T", i, result[i])
						}
					case "DownloadInfo":
						if _, ok := result[i].(dto.DownloadInfoDTO); !ok {
							t.Errorf("Item %d expected to be DownloadInfoDTO, got %T", i, result[i])
						}
					case "TestSteps":
						if _, ok := result[i].([]dto.TestStepDTO); !ok {
							t.Errorf("Item %d expected to be []TestStepDTO, got %T", i, result[i])
						}
					}
				}
			}
		})
	}
}

func TestTrimSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    json.RawMessage
		expected byte
	}{
		{
			name:     "object with leading spaces",
			input:    json.RawMessage("   {\"key\": \"value\"}"),
			expected: '{',
		},
		{
			name:     "array with leading tabs",
			input:    json.RawMessage("\t\t[1, 2, 3]"),
			expected: '[',
		},
		{
			name:     "object with mixed whitespace",
			input:    json.RawMessage(" \n\t {\"key\": \"value\"}"),
			expected: '{',
		},
		{
			name:     "no leading whitespace",
			input:    json.RawMessage("{\"key\": \"value\"}"),
			expected: '{',
		},
		{
			name:     "array with newlines",
			input:    json.RawMessage("\n\r[1, 2, 3]"),
			expected: '[',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimSpaces(tt.input)
			if len(result) == 0 {
				t.Errorf("trimSpaces() returned empty slice")
				return
			}
			if result[0] != tt.expected {
				t.Errorf("trimSpaces() first byte = %c, expected %c", result[0], tt.expected)
			}
		})
	}
}

func TestTrimSpacesEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input json.RawMessage
	}{
		{
			name:  "only whitespace",
			input: json.RawMessage("   \n\t\r   "),
		},
		{
			name:  "empty string",
			input: json.RawMessage(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimSpaces(tt.input)
			if result != nil {
				t.Errorf("trimSpaces() should return nil for whitespace-only input, got %v", result)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseMixedJSONArray(b *testing.B) {
	testJSON := `[
		{
			"TestStation": "Download",
			"SerialNumber": "SN123",
			"PartNumber": "PN456"
		},
		{
			"TestStation": "PCBA",
			"SerialNumber": "SN123",
			"LogisticData": {
				"PCBANumber": "PCBA001",
				"SerialNumber": "SN123"
			}
		},
		[
			{
				"TestStepName": "Compare PCBA Serial Number",
				"TestStepResult": "PASS",
				"TestMeasuredValue": "PCBA001"
			},
			{
				"TestStepName": "Voltage Test",
				"TestStepResult": "PASS",
				"TestMeasuredValue": "3.3V"
			}
		]
	]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseMixedJSONArray([]byte(testJSON))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTrimSpaces(b *testing.B) {
	testData := json.RawMessage("   \n\t {\"key\": \"value\"}")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trimSpaces(testData)
	}
}

// Helper function to test PCBA matching logic specifically
func TestPCBAMatchingLogic(t *testing.T) {
	tests := []struct {
		name           string
		pcbaInStation  string
		pcbaInSteps    string
		stepName       string
		shouldMatch    bool
	}{
		{
			name:          "exact match with Compare PCBA Serial Number",
			pcbaInStation: "PCBA001",
			pcbaInSteps:   "PCBA001",
			stepName:      "Compare PCBA Serial Number",
			shouldMatch:   true,
		},
		{
			name:          "exact match with PCBA Scan",
			pcbaInStation: "PCBA002",
			pcbaInSteps:   "PCBA002",
			stepName:      "PCBA Scan",
			shouldMatch:   true,
		},
		{
			name:          "no match different PCBA",
			pcbaInStation: "PCBA001",
			pcbaInSteps:   "PCBA999",
			stepName:      "Compare PCBA Serial Number",
			shouldMatch:   false,
		},
		{
			name:          "no match wrong step name",
			pcbaInStation: "PCBA001",
			pcbaInSteps:   "PCBA001",
			stepName:      "Voltage Test",
			shouldMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test input
			input := `[
				{
					"TestStation": "PCBA",
					"SerialNumber": "SN123",
					"LogisticData": {
						"PCBANumber": "` + tt.pcbaInStation + `",
						"SerialNumber": "SN123"
					}
				},
				[
					{
						"TestStepName": "` + tt.stepName + `",
						"TestStepResult": "PASS",
						"TestMeasuredValue": "` + tt.pcbaInSteps + `"
					}
				]
			]`

			result, err := ParseMixedJSONArray([]byte(input))
			if err != nil {
				t.Fatal(err)
			}

			// Should always have the TestStationRecord
			expectedLen := 1
			if tt.shouldMatch {
				expectedLen = 2 // TestStationRecord + TestSteps
			}

			if len(result) != expectedLen {
				t.Errorf("Expected %d results, got %d", expectedLen, len(result))
			}

			// Verify the types
			if len(result) >= 1 {
				if _, ok := result[0].(dto.TestStationRecordDTO); !ok {
					t.Error("First result should be TestStationRecordDTO")
				}
			}

			if tt.shouldMatch && len(result) >= 2 {
				if _, ok := result[1].([]dto.TestStepDTO); !ok {
					t.Error("Second result should be []TestStepDTO when matching")
				}
			}
		})
	}
}