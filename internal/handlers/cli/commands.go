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
	"time"

	"github.com/NoroSaroyan/log-parser/internal/app"
	"github.com/NoroSaroyan/log-parser/internal/domain/models/dto"
	"github.com/NoroSaroyan/log-parser/internal/infrastructure/logger"
	"github.com/NoroSaroyan/log-parser/internal/services/dispatcher"
	"github.com/NoroSaroyan/log-parser/internal/services/parser"
	"github.com/NoroSaroyan/log-parser/internal/services/processor"
)

func Run() error {
	mode := flag.String("mode", "process", "Mode to run: process (default), ...")
	configPath := flag.String("config", "configs/config.yaml", "Path to config file")
	logLevel := flag.String("log-level", "INFO", "Log level: DEBUG, INFO, WARN, ERROR")
	flag.Parse()

	// Initialize structured logging
	if err := logger.InitLogger(*logLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info("Starting log parser", logger.WithFields(map[string]interface{}{
		"mode":      *mode,
		"log_level": *logLevel,
	}))

	if *mode != "process" {
		return fmt.Errorf("unsupported mode: %s", *mode)
	}

	args := flag.Args()
	if len(args) == 0 {
		logger.Error("No files or directories specified")
		return fmt.Errorf("please specify at least one file or directory to process")
	}

	logger.Info("Processing files", logger.WithField("files", args))

	ctx := context.Background()
	appInstance, err := app.InitializeApp(*configPath)
	if err != nil {
		logger.Error("Failed to initialize app", logger.WithField("error", err))
		return fmt.Errorf("failed to initialize app: %w", err)
	}
	defer func() {
		if err := appInstance.CloseDB(); err != nil {
			logger.Error("Failed to close DB connection", logger.WithField("error", err))
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
	startTime := time.Now()
	
	logger.Info("Starting file processing", logger.WithField("file", filepath))
	
	// Read file
	logData, err := readFileContent(filepath)
	if err != nil {
		logger.Error("Failed to read file", logger.WithFields(map[string]interface{}{
			"file":  filepath,
			"error": err,
		}))
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	logger.Debug("File read successfully", logger.WithFields(map[string]interface{}{
		"file":       filepath,
		"size_bytes": len(logData),
	}))

	// Extract JSON blocks
	allBlocks, err := parser.ExtractJson(string(logData))
	if err != nil {
		logger.Error("Failed to extract JSON blocks", logger.WithFields(map[string]interface{}{
			"file":  filepath,
			"error": err,
		}))
		return fmt.Errorf("failed to extract JSON blocks: %w", err)
	}
	
	logger.Debug("JSON extraction completed", logger.WithFields(map[string]interface{}{
		"file":         filepath,
		"total_blocks": len(allBlocks),
	}))

	// Filter relevant blocks
	filteredBlocks, err := parser.FilterRelevantJsonBlocks(allBlocks)
	if err != nil {
		logger.Error("Failed to filter relevant JSON blocks", logger.WithFields(map[string]interface{}{
			"file":  filepath,
			"error": err,
		}))
		return fmt.Errorf("failed to filter relevant JSON blocks: %w", err)
	}
	
	logger.Debug("Block filtering completed", logger.WithFields(map[string]interface{}{
		"file":             filepath,
		"filtered_blocks":  len(filteredBlocks),
		"discarded_blocks": len(allBlocks) - len(filteredBlocks),
	}))

	// Parse JSON
	combinedJSON := "[\n" + strings.Join(filteredBlocks, ",\n") + "\n]"
	parsedItems, err := parser.ParseMixedJSONArray([]byte(combinedJSON))
	if err != nil {
		logger.Error("Failed to parse mixed JSON array", logger.WithFields(map[string]interface{}{
			"file":  filepath,
			"error": err,
		}))
		return fmt.Errorf("failed to parse mixed JSON array: %w", err)
	}

	// Calculate statistics
	stats := calculateParsingStatistics(parsedItems)
	
	logger.Info("PARSING STATISTICS", logger.WithFields(map[string]interface{}{
		"file":           filepath,
		"Final":          stats.FinalStations,
		"PCBA":           stats.PCBAStations,
		"Download":       stats.DownloadInfo,
		"TestStepArrays": stats.TestStepArrays,
		"TotalTestSteps": stats.TotalTestSteps,
	}))

	// Group data
	groupedData, err := processor.GroupByPCBANumber(parsedItems)
	if err != nil {
		logger.Error("Failed to group data", logger.WithFields(map[string]interface{}{
			"file":  filepath,
			"error": err,
		}))
		return fmt.Errorf("failed to group data: %w", err)
	}
	
	logger.Debug("Data grouping completed", logger.WithFields(map[string]interface{}{
		"file":   filepath,
		"groups": len(groupedData),
	}))

	// Dispatch to database
	if err := dispatcherService.DispatchGroups(ctx, groupedData); err != nil {
		logger.Error("Failed to dispatch groups to database", logger.WithFields(map[string]interface{}{
			"file":  filepath,
			"error": err,
		}))
		return fmt.Errorf("failed to dispatch groups: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("✅ File processing completed", logger.WithFields(map[string]interface{}{
		"file":            filepath,
		"duration":        duration,
		"groups_inserted": len(groupedData),
	}))

	return nil
}

// ParsingStatistics holds statistics about parsed items
type ParsingStatistics struct {
	FinalStations   int
	PCBAStations    int
	DownloadInfo    int
	TestStepArrays  int
	TotalTestSteps  int
}

// calculateParsingStatistics analyzes parsed items and returns detailed statistics
func calculateParsingStatistics(parsedItems []interface{}) ParsingStatistics {
	stats := ParsingStatistics{}
	
	for _, item := range parsedItems {
		switch v := item.(type) {
		case dto.TestStationRecordDTO:
			if v.TestStation == "Final" {
				stats.FinalStations++
			} else if v.TestStation == "PCBA" {
				stats.PCBAStations++
			}
		case dto.DownloadInfoDTO:
			stats.DownloadInfo++
		case []dto.TestStepDTO:
			stats.TestStepArrays++
			stats.TotalTestSteps += len(v)
		}
	}
	
	return stats
}

func readFileContent(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if strings.HasSuffix(strings.ToLower(path), ".gz") {
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

	return os.ReadFile(path)
}
