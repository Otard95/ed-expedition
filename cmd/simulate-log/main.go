package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// generateFilename takes a filename and part number, returns filename with updated part number
func generateFilename(baseFilename string, partNum int) string {
	// Pattern: Journal.2025-12-06T173759.01.json
	pattern := regexp.MustCompile(`^(Journal\.\d{4}-\d{2}-\d{2}T\d{6}\.)(\d+)(\.json)$`)

	if matches := pattern.FindStringSubmatch(baseFilename); matches != nil {
		// Use the base filename pattern with new part number
		return fmt.Sprintf("%s%02d%s", matches[1], partNum, matches[3])
	}

	// Fallback: append part number before extension
	ext := filepath.Ext(baseFilename)
	base := baseFilename[:len(baseFilename)-len(ext)]
	return fmt.Sprintf("%s.%02d%s", base, partNum, ext)
}

func main() {
	// Generate default filename with current timestamp
	now := time.Now()
	defaultFilename := fmt.Sprintf("Journal.%s.01.json", now.Format("2006-01-02T150405"))

	// Command-line flags
	inputFile := flag.String("input", "", "Path to JSON array file")
	outputDir := flag.String("output-dir", ".", "Directory to write log file to")
	outputFile := flag.String("output-file", defaultFilename, "Name of output log file")
	delay := flag.Int("delay", 100, "Delay between writes in milliseconds")
	count := flag.Int("count", 1, "Number of files to split output into (increments part number)")
	flag.Parse()

	// Validate required flags
	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Read input JSON array
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON array
	var events []map[string]any
	if err := json.Unmarshal(data, &events); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Validate count
	if *count < 1 {
		fmt.Fprintln(os.Stderr, "Error: -count must be at least 1")
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total events: %d\n", len(events))
	fmt.Printf("Delay between writes: %dms\n", *delay)
	fmt.Printf("Splitting into %d file(s)\n\n", *count)

	// Calculate events per file
	eventsPerFile := len(events) / *count
	remainder := len(events) % *count

	delayDuration := time.Duration(*delay) * time.Millisecond
	eventIndex := 0

	for fileNum := 0; fileNum < *count; fileNum++ {
		// Calculate how many events go in this file
		numEvents := eventsPerFile
		if fileNum < remainder {
			numEvents++ // Distribute remainder across first files
		}

		// Generate filename with incremented part number
		filename := generateFilename(*outputFile, fileNum+1)
		outPath := filepath.Join(*outputDir, filename)

		// Open output file
		out, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Writing to: %s (%d events)\n", filename, numEvents)

		// Write events for this file
		for i := 0; i < numEvents; i++ {
			event := events[eventIndex]
			event["timestamp"] = time.Now().Format(time.RFC3339)

			// Marshal event to JSON
			line, err := json.Marshal(event)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling event %d: %v\n", eventIndex, err)
				eventIndex++
				continue
			}

			// Write line
			if _, err := fmt.Fprintf(out, "%s\n", line); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing event %d: %v\n", eventIndex, err)
				eventIndex++
				continue
			}

			// Flush to ensure write is visible immediately
			if err := out.Sync(); err != nil {
				fmt.Fprintf(os.Stderr, "Error syncing event %d: %v\n", eventIndex, err)
			}

			fmt.Printf("  [%d/%d] %s\n", eventIndex+1, len(events), event["event"])

			eventIndex++

			// Delay before next write (except after last event)
			if eventIndex < len(events) {
				time.Sleep(delayDuration)
			}
		}

		out.Close()
		fmt.Println()
	}

	fmt.Println("Simulation complete!")
}
