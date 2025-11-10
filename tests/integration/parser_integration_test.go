package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NoroSaroyan/log-parser/internal/services/parser"
)

func TestExtractorIntegration(t *testing.T) {
	testDataDir := filepath.Join("..", "testdata", "extractor_test_logs")

	tests := []struct {
		name           string
		filename       string
		expectedBlocks int
	}{
		{
			name:           "simple single object",
			filename:       "simple_single_object.log",
			expectedBlocks: 1,
		},
		{
			name:           "simple single array",
			filename:       "simple_single_array.log",
			expectedBlocks: 1,
		},
		{
			name:           "multiline nested object",
			filename:       "multiline_nested_object.log",
			expectedBlocks: 1,
		},
		{
			name:           "multiple blocks",
			filename:       "multiple_blocks.log",
			expectedBlocks: 4,
		},
		{
			name:           "no json blocks",
			filename:       "no_json_blocks.log",
			expectedBlocks: 0,
		},
		{
			name:           "malformed json",
			filename:       "malformed_json.log",
			expectedBlocks: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(testDataDir, tt.filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", filePath, err)
			}

			blocks, err := parser.ExtractJson(string(content))
			if err != nil {
				t.Fatalf("ExtractJson failed: %v", err)
			}

			if len(blocks) != tt.expectedBlocks {
				t.Errorf("Expected %d blocks, got %d", tt.expectedBlocks, len(blocks))
				for i, block := range blocks {
					t.Logf("Block %d: %s", i, block)
				}
			}
		})
	}
}

func TestParserIntegration(t *testing.T) {
	testDataDir := filepath.Join("..", "testdata", "parser_test_data")

	tests := []struct {
		name          string
		filename      string
		expectedItems int
	}{
		{
			name:          "valid mixed array",
			filename:      "valid_mixed_array.json",
			expectedItems: 4, // Download + PCBA + Final + TestSteps
		},
		{
			name:          "teststeps no matching pcba",
			filename:      "teststeps_no_matching_pcba.json",
			expectedItems: 1, // Only PCBA record, TestSteps filtered out
		},
		{
			name:          "invalid teststation",
			filename:      "invalid_teststation.json",
			expectedItems: 0, // All invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(testDataDir, tt.filename)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", filePath, err)
			}

			items, err := parser.ParseMixedJSONArray(content)
			if err != nil {
				t.Fatalf("ParseMixedJSONArray failed: %v", err)
			}

			if len(items) != tt.expectedItems {
				t.Errorf("Expected %d items, got %d", tt.expectedItems, len(items))
				for i, item := range items {
					t.Logf("Item %d: %T", i, item)
				}
			}
		})
	}
}

func TestFullPipelineIntegration(t *testing.T) {
	// Test the complete pipeline: Extract -> Filter -> Parse
	logContent := `2024-06-03 14:30:25.123 [INFO]: Starting process
2024-06-03 14:30:25.456 [DEBUG]: Data  {"TestStation": "Download", "SerialNumber": "SN123", "PartNumber": "PN456"}
2024-06-03 14:30:25.600 [DEBUG]: Data  {"TestStation": "PCBA", "SerialNumber": "SN123", "LogisticData": {"PCBANumber": "PCBA001"}}
2024-06-03 14:30:25.800 [DEBUG]: Data  [
2024-06-03 14:30:25.801 [DEBUG]:   {"TestStepName": "Compare PCBA Serial Number", "TestStepResult": "PASS", "TestMeasuredValue": "PCBA001"}
2024-06-03 14:30:25.803 [DEBUG]: ]
2024-06-03 14:30:25.999 [INFO]: Process complete`

	// Step 1: Extract JSON blocks
	blocks, err := parser.ExtractJson(logContent)
	if err != nil {
		t.Fatalf("ExtractJson failed: %v", err)
	}

	if len(blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(blocks))
	}

	// Step 2: Filter relevant blocks
	filtered, err := parser.FilterRelevantJsonBlocks(blocks)
	if err != nil {
		t.Fatalf("FilterRelevantJsonBlocks failed: %v", err)
	}

	if len(filtered) != 3 {
		t.Fatalf("Expected 3 filtered blocks, got %d", len(filtered))
	}

	// Step 3: Parse into mixed array format
	combinedJSON := "[" + filtered[0] + "," + filtered[1] + "," + filtered[2] + "]"
	items, err := parser.ParseMixedJSONArray([]byte(combinedJSON))
	if err != nil {
		t.Fatalf("ParseMixedJSONArray failed: %v", err)
	}

	// Should have: DownloadInfo + PCBA Station + TestSteps = 3 items
	expectedItems := 3
	if len(items) != expectedItems {
		t.Errorf("Expected %d items, got %d", expectedItems, len(items))
		for i, item := range items {
			t.Logf("Item %d: %T", i, item)
		}
	}
}
