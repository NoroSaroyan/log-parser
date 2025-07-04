package cli

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"log-parser/internal/app"
	"log-parser/internal/services/dispatcher"
	"log-parser/internal/services/parser"
	"log-parser/internal/services/processor"
)

// Run parses flags and runs the CLI application.
func Run() error {
	mode := flag.String("mode", "process", "Mode to run: process (default), ...")
	configPath := flag.String("config", "configs/config.yaml", "Path to config file")
	flag.Parse()

	if *mode != "process" {
		return fmt.Errorf("unsupported mode: %s", *mode)
	}

	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("please specify at least one file or directory to process")
	}

	ctx := context.Background()
	appInstance, err := app.InitializeApp(*configPath)
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
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

	for _, path := range args {
		fi, err := os.Stat(path)
		if err != nil {
			log.Printf("Skipping %s: %v", path, err)
			continue
		}

		if fi.IsDir() {
			err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && isSupportedFile(p) {
					if err := processSingleFile(ctx, p, appInstance, dispatcherService); err != nil {
						log.Printf("Error processing file %s: %v", p, err)
					}
				}
				return nil
			})
			if err != nil {
				log.Printf("Error walking directory %s: %v", path, err)
			}
		} else {
			if err := processSingleFile(ctx, path, appInstance, dispatcherService); err != nil {
				log.Printf("Error processing file %s: %v", path, err)
			}
		}
	}

	return nil
}

func isSupportedFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".log") || strings.HasSuffix(lower, ".txt") ||
		strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".gz")
}

func processSingleFile(ctx context.Context, filepath string, appInstance *app.App, dispatcherService dispatcher.DispatcherService) error {
	logData, err := readFileContent(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	allBlocks, err := parser.ExtractJson(string(logData))
	if err != nil {
		return fmt.Errorf("failed to extract JSON blocks: %w", err)
	}

	filteredBlocks, err := parser.FilterRelevantJsonBlocks(allBlocks)
	if err != nil {
		return fmt.Errorf("failed to filter relevant JSON blocks: %w", err)
	}

	combinedJSON := "[\n" + strings.Join(filteredBlocks, ",\n") + "\n]"

	parsedItems, err := parser.ParseMixedJSONArray([]byte(combinedJSON))
	if err != nil {
		return fmt.Errorf("failed to parse mixed JSON array: %w", err)
	}

	groupedData, err := processor.GroupByPCBANumber(parsedItems)
	if err != nil {
		return fmt.Errorf("failed to group data: %w", err)
	}

	if err := dispatcherService.DispatchGroups(ctx, groupedData); err != nil {
		return fmt.Errorf("failed to dispatch groups: %w", err)
	}

	fmt.Printf("Processed file %s: %d relevant blocks, %d items parsed\n", filepath, len(filteredBlocks), len(parsedItems))
	return nil
}

func readFileContent(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if strings.HasSuffix(strings.ToLower(path), ".gz") {
		// Decompress gzip
		gr, err := gzip.NewReader(f)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gr.Close()

		var buf bytes.Buffer
		scanner := bufio.NewScanner(gr)
		for scanner.Scan() {
			buf.Write(scanner.Bytes())
			buf.WriteByte('\n')
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading gzip content: %w", err)
		}
		return buf.Bytes(), nil
	}

	// Regular file read
	return os.ReadFile(path)
}
