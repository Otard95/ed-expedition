package main

import (
	"ed-expedition/plotters"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	inputPath := flag.String("i", "", "Input spansh.data.json file path")
	outputPath := flag.String("o", "", "Output pruned JSON file path")
	force := flag.Bool("force", false, "Allow input and output to be the same file")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" {
		fmt.Println("Usage: prune-spansh-data -i <input> -o <output> [--force]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *inputPath == *outputPath && !*force {
		fmt.Println("Error: input and output paths are the same. Use --force to override.")
		os.Exit(1)
	}

	data, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var spanshData plotters.SpanshDataStruct
	if err := json.Unmarshal(data, &spanshData); err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(spanshData, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling output: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputPath, output, 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully pruned %s -> %s\n", *inputPath, *outputPath)
}
