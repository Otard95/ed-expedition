package main

import (
	"ed-expedition/journal"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// Fixed timestamp for normalized output comparison
const fixedTimestamp = "2025-01-01T00:00:00Z"

func main() {
	// Command-line flags
	inputFile := flag.String("input", "", "Path to journal JSON array file")
	fromTime := flag.String("from", "", "Only show events after this time (RFC3339 format)")
	normalizeTime := flag.Bool("normalize-time", false, "Use fixed timestamp for output (for comparison)")
	flag.Parse()

	// Validate required flags
	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse fromTime if provided
	var filterTime time.Time
	if *fromTime != "" {
		var err error
		filterTime, err = time.Parse(time.RFC3339, *fromTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing -from time: %v\n", err)
			os.Exit(1)
		}
	}

	// Read input JSON array
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON array
	var events []json.RawMessage
	if err := json.Unmarshal(data, &events); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Processing %d events from %s\n", len(events), *inputFile)
	if !filterTime.IsZero() {
		fmt.Printf("Filtering events after: %s\n", filterTime.Format(time.RFC3339))
	}
	fmt.Println()

	// Process each event
	for _, eventData := range events {
		// Determine event type
		var base struct {
			Timestamp time.Time         `json:"timestamp"`
			Event     journal.EventType `json:"event"`
		}
		if err := json.Unmarshal(eventData, &base); err != nil {
			continue
		}

		// Filter by time if specified
		if !filterTime.IsZero() && base.Timestamp.Before(filterTime) {
			continue
		}

		// Process based on event type
		switch base.Event {
		case journal.Loadout:
			var event journal.LoadoutEvent
			if err := json.Unmarshal(eventData, &event); err == nil {
				fmt.Printf("[Loadout] Ship: %s (%s) | Max Jump: %.2f LY\n",
					event.Ship, event.ShipName, event.MaxJumpRange)
			}

		case journal.FSDJump:
			var event journal.FSDJumpEvent
			if err := json.Unmarshal(eventData, &event); err == nil {
				timestamp := event.Timestamp
				if *normalizeTime {
					timestamp, _ = time.Parse(time.RFC3339, fixedTimestamp)
				}
				fmt.Printf("[FSDJump] %s → %s | Distance: %.2f LY | Fuel: %.2f\n",
					timestamp.Format("15:04:05"),
					event.StarSystem,
					event.JumpDist,
					event.FuelLevel)
			}

		case journal.FSDTarget:
			var event journal.FSDTargetEvent
			if err := json.Unmarshal(eventData, &event); err == nil {
				if *normalizeTime {
					timestamp, _ := time.Parse(time.RFC3339, fixedTimestamp)
					event.Timestamp = timestamp
				}
				fmt.Printf("[FSDTarget] &%+v\n", event)
			}
		}
	}
}
