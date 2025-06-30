package cli

import (
	"flag"
	"fmt"
)

func Run() {
	mode := flag.String("mode", "parser", "Mode to run: parser, extractor, analyzer")
	flag.Parse()

	switch *mode {
	case "parser":
		fmt.Println("Run parser")
	case "extractor":
		fmt.Println("Run extractor")
	case "analyzer":
		fmt.Println("Run analyzer")
	default:
		fmt.Println("Unknown mode")
	}
}
