package main

import (
	"fmt"
	"os"

	"github.com/NoroSaroyan/log-parser/internal/handlers/cli"
	"github.com/NoroSaroyan/log-parser/internal/infrastructure/logger"
)

func main() {
	if err := cli.Run(); err != nil {
		// Initialize logger if not already done, then log the fatal error
		if err := logger.InitLogger("ERROR"); err == nil {
			logger.Fatal("CLI error", err)
		}
		// Fallback to fmt if logger initialization fails
		fmt.Fprintf(os.Stderr, "CLI error: %v\n", err)
		os.Exit(1)
	}
}
