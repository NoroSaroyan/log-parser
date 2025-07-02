package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"log-parser/internal/services/parser"
	"time"
)

func main() {
	start := time.Now()
	inputFile := "large2.log"
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

	outputData := "[\n"
	for i, block := range filteredBlocks {
		outputData += block
		if i < len(filteredBlocks)-1 {
			outputData += ",\n"
		}
	}
	outputData += "\n]"

	err = ioutil.WriteFile(outputFile, []byte(outputData), 0644)
	if err != nil {
		log.Fatalf("Failed to write output JSON file: %v", err)
	}

	fmt.Printf("Extracted %d relevant JSON blocks into %s\n", len(filteredBlocks), outputFile)
	fmt.Println()
	fmt.Println(time.Since(start))
}
