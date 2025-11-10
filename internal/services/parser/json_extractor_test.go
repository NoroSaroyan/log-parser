package parser

import (
	"testing"
)

func TestExtractJson(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name: "single JSON object block",
			input: `2024-06-03 14:30:25.123 [INFO]: Starting process
2024-06-03 14:30:25.456 [DEBUG]: Data  {"TestStation": "PCBA", "SerialNumber": "ABC123"}
2024-06-03 14:30:25.789 [INFO]: Process complete`,
			expected: 1,
			wantErr:  false,
		},
		{
			name: "single JSON array block",
			input: `2024-06-03 14:30:25.123 [INFO]: Starting process
2024-06-03 14:30:25.456 [DEBUG]: Data  [{"TestStepName": "Test1", "Result": "PASS"}]
2024-06-03 14:30:25.789 [INFO]: Process complete`,
			expected: 1,
			wantErr:  false,
		},
		{
			name: "multiline JSON object",
			input: `2024-06-03 14:30:25.123 [INFO]: Starting process
2024-06-03 14:30:25.456 [DEBUG]: Data  {
2024-06-03 14:30:25.457 [DEBUG]:   "TestStation": "PCBA",
2024-06-03 14:30:25.458 [DEBUG]:   "SerialNumber": "ABC123"
2024-06-03 14:30:25.459 [DEBUG]: }
2024-06-03 14:30:25.789 [INFO]: Process complete`,
			expected: 1,
			wantErr:  false,
		},
		{
			name: "nested JSON object",
			input: `2024-06-03 14:30:25.123 [INFO]: Starting process
2024-06-03 14:30:25.456 [DEBUG]: Data  {
2024-06-03 14:30:25.457 [DEBUG]:   "TestStation": "PCBA",
2024-06-03 14:30:25.458 [DEBUG]:   "LogisticData": {
2024-06-03 14:30:25.459 [DEBUG]:     "PCBANumber": "12345",
2024-06-03 14:30:25.460 [DEBUG]:     "SerialNumber": "SN123"
2024-06-03 14:30:25.461 [DEBUG]:   }
2024-06-03 14:30:25.462 [DEBUG]: }
2024-06-03 14:30:25.789 [INFO]: Process complete`,
			expected: 1,
			wantErr:  false,
		},
		{
			name: "multiple JSON blocks",
			input: `2024-06-03 14:30:25.123 [INFO]: Starting process
2024-06-03 14:30:25.456 [DEBUG]: Data  {"TestStation": "PCBA", "SerialNumber": "ABC123"}
2024-06-03 14:30:25.789 [DEBUG]: Data  [{"TestStepName": "Test1", "Result": "PASS"}]
2024-06-03 14:30:25.999 [INFO]: Process complete`,
			expected: 2,
			wantErr:  false,
		},
		{
			name:     "no JSON blocks",
			input:    `2024-06-03 14:30:25.123 [INFO]: Starting process\n2024-06-03 14:30:25.789 [INFO]: Process complete`,
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
			wantErr:  false,
		},
		{
			name: "malformed JSON block - unclosed brace",
			input: `2024-06-03 14:30:25.123 [INFO]: Starting process
2024-06-03 14:30:25.456 [DEBUG]: Data  {"TestStation": "PCBA"
2024-06-03 14:30:25.789 [INFO]: Process complete`,
			expected: 1,
			wantErr:  false,
		},
		{
			name: "JSON with arrays in object",
			input: `2024-06-03 14:30:25.456 [DEBUG]: Data  {
2024-06-03 14:30:25.457 [DEBUG]:   "TestStation": "PCBA",
2024-06-03 14:30:25.458 [DEBUG]:   "TestSteps": [
2024-06-03 14:30:25.459 [DEBUG]:     {"name": "test1"},
2024-06-03 14:30:25.460 [DEBUG]:     {"name": "test2"}
2024-06-03 14:30:25.461 [DEBUG]:   ]
2024-06-03 14:30:25.462 [DEBUG]: }`,
			expected: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractJson(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(result) != tt.expected {
				t.Errorf("ExtractJson() returned %d blocks, expected %d", len(result), tt.expected)
				for i, block := range result {
					t.Logf("Block %d: %s", i, block)
				}
			}
		})
	}
}

func TestFilterRelevantJsonBlocks(t *testing.T) {
	tests := []struct {
		name     string
		blocks   []string
		expected int
		wantErr  bool
	}{
		{
			name: "valid TestStationRecord",
			blocks: []string{
				`{"TestStation": "PCBA", "SerialNumber": "ABC123", "LogisticData": {"PCBANumber": "12345"}}`,
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "valid DownloadInfo",
			blocks: []string{
				`{"TestStation": "Download", "SerialNumber": "ABC123"}`,
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "valid TestStep array",
			blocks: []string{
				`[{"TestStepName": "Test1", "TestStepResult": "PASS", "TestMeasuredValue": "123"}]`,
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "valid three-element array structure",
			blocks: []string{
				`[{"TestStation": "Download", "SerialNumber": "ABC123"}, [{"TestStepName": "Test1", "TestStepResult": "PASS"}], {"TestStation": "PCBA", "SerialNumber": "ABC123"}]`,
			},
			expected: 1,
			wantErr:  false,
		},
		{
			name: "mixed valid and invalid blocks",
			blocks: []string{
				`{"TestStation": "PCBA", "SerialNumber": "ABC123"}`,
				`{"invalid": "json"}`,
				`[{"TestStepName": "Test1", "TestStepResult": "PASS"}]`,
				`not json at all`,
			},
			expected: 2,
			wantErr:  false,
		},
		{
			name: "empty TestStation field",
			blocks: []string{
				`{"TestStation": "", "SerialNumber": "ABC123"}`,
			},
			expected: 0,
			wantErr:  true,
		},
		{
			name: "empty test steps array",
			blocks: []string{
				`[]`,
			},
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "no blocks provided",
			blocks:   []string{},
			expected: 0,
			wantErr:  true,
		},
		{
			name: "invalid JSON syntax",
			blocks: []string{
				`{"TestStation": "PCBA", "SerialNumber":}`,
			},
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterRelevantJsonBlocks(tt.blocks)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterRelevantJsonBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if len(result) != tt.expected {
				t.Errorf("FilterRelevantJsonBlocks() returned %d blocks, expected %d", len(result), tt.expected)
				for i, block := range result {
					t.Logf("Filtered block %d: %s", i, block)
				}
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "simple string",
			input:    "test",
			expected: `"test"`,
		},
		{
			name:     "simple number",
			input:    123,
			expected: `123`,
		},
		{
			name:     "simple object",
			input:    map[string]string{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "array",
			input:    []string{"a", "b", "c"},
			expected: `["a","b","c"]`,
		},
		{
			name:     "nil value",
			input:    nil,
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toJSON(tt.input)
			if result != tt.expected {
				t.Errorf("toJSON() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkExtractJson(b *testing.B) {
	largeLog := `2024-06-03 14:30:25.123 [INFO]: Starting process
2024-06-03 14:30:25.456 [DEBUG]: Data  {"TestStation": "PCBA", "SerialNumber": "ABC123", "LogisticData": {"PCBANumber": "12345"}}
2024-06-03 14:30:25.457 [DEBUG]: Data  [{"TestStepName": "Test1", "TestStepResult": "PASS", "TestMeasuredValue": "123"}, {"TestStepName": "Test2", "TestStepResult": "PASS", "TestMeasuredValue": "456"}]
2024-06-03 14:30:25.789 [INFO]: Process complete`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ExtractJson(largeLog)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFilterRelevantJsonBlocks(b *testing.B) {
	blocks := []string{
		`{"TestStation": "PCBA", "SerialNumber": "ABC123", "LogisticData": {"PCBANumber": "12345"}}`,
		`{"TestStation": "Download", "SerialNumber": "ABC123"}`,
		`[{"TestStepName": "Test1", "TestStepResult": "PASS", "TestMeasuredValue": "123"}]`,
		`{"invalid": "json"}`,
		`not json at all`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FilterRelevantJsonBlocks(blocks)
		if err != nil {
			b.Fatal(err)
		}
	}
}