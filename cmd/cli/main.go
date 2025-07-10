package main

import (
	"github.com/NoroSaroyan/log-parser/internal/handlers/cli"
	"log"
)

func main() {
	if err := cli.Run(); err != nil {
		log.Fatalf("CLI error: %v", err)
	}
}
