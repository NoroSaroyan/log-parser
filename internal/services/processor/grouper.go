package processor

import (
	"fmt"
	"log-parser/internal/domain/models/dto"
)

func GroupByPCBANumber(parsed []interface{}) ([]dto.GroupedDataDTO, error) {
	// Map: pcbNumber -> GroupedDataDTO
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
			// Тут нужно понять, к какой группе относится этот массив тестов
			// Ищем в массиве TestStep с TestStepName == "PCBA Test", берем ThresholdValue как ключ
			var key string
			for _, step := range v {
				if step.TestStepName == "PCBA Test" {
					key = step.TestThresholdValue
					break
				}
			}
			if key == "" {
				return nil, fmt.Errorf("TestStepDTO array missing PCBA Test step with threshold value")
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

	// Преобразуем map в срез
	result := make([]dto.GroupedDataDTO, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}

	return result, nil
}
