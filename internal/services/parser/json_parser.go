package parser

//
//import (
//	"fmt"
//	"log-parser/internal/domain/models"
//)
//
//// ParsedResult represents a parsed JSON object from the logs.
//type ParsedResult struct {
//	StationInfo *models.StationInformation
//	TestData    *models.TestData
//	RawJSON     string
//}
//
//// ParseJSON takes a JSON string and tries to deserialize it into one of the known models.
//// It first tries StationInformation, then TestData ([]TestStep).
//func ParseJSON(jsonStr string) (*ParsedResult, error) {
//	result := &ParsedResult{RawJSON: jsonStr}
//
//	// Try to parse as StationInformation
//	if station, err := models.UnmarshalStationInformation([]byte(jsonStr)); err == nil && station.PartNumber != "" {
//		result.StationInfo = &station
//		return result, nil
//	}
//
//	// Try to parse as TestData ([]TestStep)
//	if testData, err := models.UnmarshalTestData([]byte(jsonStr)); err == nil && len(testData) > 0 {
//		result.TestData = &testData
//		return result, nil
//	}
//
//	// Unrecognized JSON structure
//	return nil, fmt.Errorf("unable to parse JSON into known models: %s", truncate(jsonStr, 200))
//}
//
//// truncate helps log shortened JSON previews for error output
//func truncate(s string, max int) string {
//	if len(s) <= max {
//		return s
//	}
//	return s[:max] + "..."
//}
