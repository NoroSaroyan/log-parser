package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"log-parser/internal/app"
	"log-parser/internal/services/dispatcher"
	"log-parser/internal/services/parser"
	"log-parser/internal/services/processor"
)

func main() {
	inputFile := "/Users/noriksaroyan/GolandProjects/log-parser/corporate_resources/mesrestapi2.log"
	outputFile := "output.json"

	logData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read log file: %v", err)
	}

	allBlocks, err := parser.ExtractJson(string(logData))
	if err != nil {
		log.Fatalf("Failed to extract JSON blocks: %v", err)
	}

	filteredBlocks, err := parser.FilterRelevantJsonBlocks(allBlocks)
	if err != nil {
		log.Fatalf("Failed to filter relevant JSON blocks: %v", err)
	}

	combinedJSON := "[\n"
	for i, block := range filteredBlocks {
		combinedJSON += block
		if i < len(filteredBlocks)-1 {
			combinedJSON += ",\n"
		}
	}
	combinedJSON += "\n]"

	parsedItems, err := parser.ParseMixedJSONArray([]byte(combinedJSON))
	if err != nil {
		log.Fatalf("Failed to parse mixed JSON array: %v", err)
	}

	err = ioutil.WriteFile(outputFile, []byte(combinedJSON), 0644)
	if err != nil {
		log.Fatalf("Failed to write output JSON file: %v", err)
	}

	fmt.Printf("Extracted %d relevant JSON blocks into %s\n", len(filteredBlocks), outputFile)
	fmt.Printf("Parsed %d items from combined JSON array\n", len(parsedItems))

	ctx := context.Background()

	appInstance, err := app.InitializeApp("configs/config.yaml")
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer func() {
		if err := appInstance.CloseDB(); err != nil {
			log.Printf("failed to close DB connection: %v", err)
		}
	}()

	dispatcherService := dispatcher.NewDispatcherService(
		appInstance.DownloadInfoService,
		appInstance.LogisticService,
		appInstance.TestStationService,
		appInstance.TestStepService,
	)

	groupedData, err := processor.GroupByPCBANumber(parsedItems)
	if err != nil {
		log.Fatalf("failed to group data: %v", err)
	}

	if err := dispatcherService.DispatchGroups(ctx, groupedData); err != nil {
		log.Fatalf("failed to dispatch groups: %v", err)
	}
}
