package main

import (
	"log"
	"log-parser/internal/handlers/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		log.Fatalf("CLI error: %v", err)
	}
}
