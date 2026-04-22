package parser

import (
	"strings"

	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
)

// InferStationTypeFromSteps inspects a TestSteps array and infers which station
// it belongs to, based on the first recognized scan step:
//
//	"PCBA Scan"                                                 -> "PCBA"
//	"Compare PCBA Serial Number" / "Valid PCBA Serial Number"   -> "Final"
//
// The returned pcba is the TestMeasuredValue of that scan step (trimmed).
// Returns ("", "") when no recognized scan step is present.
func InferStationTypeFromSteps(steps []dto.TestStepDTO) (stationType, pcba string) {
	for _, s := range steps {
		switch s.TestStepName {
		case "PCBA Scan":
			return "PCBA", strings.TrimSpace(s.GetMeasuredValueString())
		case "Compare PCBA Serial Number", "Valid PCBA Serial Number":
			return "Final", strings.TrimSpace(s.GetMeasuredValueString())
		}
	}
	return "", ""
}
