package processor

import (
	"fmt"
	"log-parser/internal/domain/models/dto"
)

func GroupByPCBANumber(parsed []interface{}) ([]dto.GroupedDataDTO, error) {
	groups := map[string]*dto.GroupedDataDTO{}
	for _, item := range parsed {
		switch v := item.(type) {
		case dto.DownloadInfoDTO:
			key := v.TcuPCBANumber
			if key == "" {
				return nil, fmt.Errorf("DownloadInfoDTO missing TcuPCBANumber")
			}
			group, ok := groups[key]
			if !ok {
				group = &dto.GroupedDataDTO{}
				groups[key] = group
			}
			group.DownloadInfo = v

		case dto.TestStationRecordDTO:
			key := v.LogisticData.PCBANumber
			if key == "" {
				return nil, fmt.Errorf("TestStationRecordDTO missing LogisticData.PCBANumber")
			}
			group, ok := groups[key]
			if !ok {
				group = &dto.GroupedDataDTO{}
				groups[key] = group
			}
			group.TestStationRecords = append(group.TestStationRecords, v)

		case []dto.TestStepDTO:
			var key string
			for _, step := range v {
				if step.TestStepName == "PCBA Scan" || step.TestStepName == "Compare PCBA Serial Number" {
					key = step.GetMeasuredValueString()
					break
				}
			}

			if key == "" {
				return nil, fmt.Errorf("TestStepDTO array missing PCBA Scan step with measured value")
			}
			group, ok := groups[key]
			if !ok {
				group = &dto.GroupedDataDTO{}
				groups[key] = group
			}
			group.TestSteps = append(group.TestSteps, v)

		default:
			return nil, fmt.Errorf("unexpected type in parsed data: %T", v)
		}
	}

	result := make([]dto.GroupedDataDTO, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}

	return result, nil
}
