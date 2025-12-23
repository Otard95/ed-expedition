package main

import (
	"ed-expedition/journal"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TestLogger implements wails Logger interface for this test tool
type TestLogger struct{}

func (l *TestLogger) Print(message string)   {}
func (l *TestLogger) Trace(message string)   {}
func (l *TestLogger) Debug(message string)   {}
func (l *TestLogger) Info(message string)    {}
func (l *TestLogger) Warning(message string) {}
func (l *TestLogger) Error(message string)   {}
func (l *TestLogger) Fatal(message string)   {}

// Fixed timestamp for normalized output comparison
const fixedTimestamp = "2025-01-01T00:00:00Z"

func main() {
	// Command-line flags
	normalizeTime := flag.Bool("normalize-time", false, "Use fixed timestamp for output (for comparison)")
	flag.Parse()

	// Create watcher for ./data/journals
	journalDir := "./data/journals"

	watcher, err := journal.NewWatcher(journalDir, &TestLogger{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create watcher: %v\n", err)
		os.Exit(1)
	}
	defer watcher.Close()

	// Subscribe to all event types
	loadoutCh := watcher.Loadout.Subscribe()
	fsdJumpCh := watcher.FSDJump.Subscribe()
	fsdTargetCh := watcher.FSDTarget.Subscribe()

	// Start goroutines to print events
	go func() {
		for event := range loadoutCh {
			fmt.Printf("[Loadout] Ship: %s (%s) | Max Jump: %.2f LY\n",
				event.Ship, event.ShipName, event.MaxJumpRange)
		}
	}()

	go func() {
		for event := range fsdJumpCh {
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
	}()

	go func() {
		for event := range fsdTargetCh {
			if *normalizeTime {
				timestamp, _ := time.Parse(time.RFC3339, fixedTimestamp)
				event.Timestamp = timestamp
			}
			fmt.Printf("[FSDTarget] %+v\n", event)
		}
	}()

	// Start watching
	fmt.Printf("Watching journal directory: %s\n", journalDir)
	fmt.Println("Waiting for events... (Ctrl+C to exit)")
	watcher.Start()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
}
