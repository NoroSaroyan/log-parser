package main

import (
	"fmt"
	"log"
	"log-parser/internal/services/parser"
	"os"
	"reflect"
)

func main() {
	filePath := "/Users/noriksaroyan/GolandProjects/log-parser/corporate_resources/output.json"
	outputPath := "/Users/noriksaroyan/GolandProjects/log-parser/corporate_resources/parsed_output.txt"

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	results, err := parser.ParseMixedJSONArray(data)
	if err != nil {
		log.Fatalf("Parsing failed: %v", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	for i, obj := range results {
		_, _ = fmt.Fprintf(f, "Parsed object #%d:\n", i)
		writeStructIndented(f, obj, "  ")
		_, _ = fmt.Fprintln(f, "\n")
	}

	fmt.Printf("Parsing result saved to %s\n", outputPath)
}

func writeStructIndented(f *os.File, obj interface{}, indent string) {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			_, _ = fmt.Fprintf(f, "%s- Item %d:\n", indent, i)
			writeStructIndented(f, v.Index(i).Interface(), indent+"  ")
		}
		return
	}

	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i).Interface()
			_, _ = fmt.Fprintf(f, "%s%s: %v\n", indent, field.Name, fieldValue)
		}
		return
	}

	_, _ = fmt.Fprintf(f, "%s%v\n", indent, obj)
}
