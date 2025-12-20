package main

import (
	"bufio"
	"ed-expedition/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type REPL struct {
	expedition  *models.Expedition
	bakedRoute  *models.Route
	journalFile *os.File
	journalPath string
	scanner     *bufio.Scanner
}

func NewREPL(journalDir string) (*REPL, error) {
	// Load active expedition
	index, err := models.LoadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	expedition, err := index.LoadActiveExpedition()
	if err != nil {
		return nil, fmt.Errorf("failed to load active expedition: %w", err)
	}
	if expedition == nil {
		return nil, fmt.Errorf("no active expedition found")
	}

	// Load baked route
	bakedRoute, err := expedition.LoadBaked()
	if err != nil {
		return nil, fmt.Errorf("failed to load baked route: %w", err)
	}

	// Create/open journal file
	now := time.Now()
	filename := fmt.Sprintf("Journal.%s.01.json", now.Format("2006-01-02T150405"))
	journalPath := filepath.Join(journalDir, filename)

	journalFile, err := os.OpenFile(journalPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open journal file: %w", err)
	}

	return &REPL{
		expedition:  expedition,
		bakedRoute:  bakedRoute,
		journalFile: journalFile,
		journalPath: journalPath,
		scanner:     bufio.NewScanner(os.Stdin),
	}, nil
}

func (r *REPL) Close() {
	if r.journalFile != nil {
		r.journalFile.Close()
	}
}

func (r *REPL) writeJump(systemName string, systemID int64, distance float64) error {
	event := map[string]interface{}{
		"timestamp":                     time.Now().UTC().Format(time.RFC3339),
		"event":                         "FSDJump",
		"Taxi":                          false,
		"Multicrew":                     false,
		"StarSystem":                    systemName,
		"SystemAddress":                 systemID,
		"StarPos":                       []float64{0, 0, 0},
		"SystemAllegiance":              "",
		"SystemEconomy":                 "",
		"SystemEconomy_Localised":       "",
		"SystemSecondEconomy":           "",
		"SystemSecondEconomy_Localised": "",
		"SystemGovernment":              "",
		"SystemGovernment_Localised":    "",
		"SystemSecurity":                "",
		"SystemSecurity_Localised":      "",
		"Population":                    0,
		"Body":                          "",
		"BodyID":                        0,
		"BodyType":                      "",
		"JumpDist":                      distance,
		"FuelUsed":                      2.5,
		"FuelLevel":                     28.0,
	}

	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := fmt.Fprintf(r.journalFile, "%s\n", line); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	if err := r.journalFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Give the watcher time to process the event
	time.Sleep(50 * time.Millisecond)

	// Reload expedition to get updated state
	if err := r.reloadExpedition(); err != nil {
		return fmt.Errorf("failed to reload expedition: %w", err)
	}

	return nil
}

func (r *REPL) writeTarget(systemName string, systemID int64, starClass string) error {
	event := map[string]interface{}{
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"event":         "FSDTarget",
		"Name":          systemName,
		"SystemAddress": systemID,
		"StarClass":     starClass,
	}

	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := fmt.Fprintf(r.journalFile, "%s\n", line); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	if err := r.journalFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Give the watcher time to process the event
	time.Sleep(50 * time.Millisecond)

	return nil
}

func (r *REPL) reloadExpedition() error {
	index, err := models.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	expedition, err := index.LoadActiveExpedition()
	if err != nil {
		return fmt.Errorf("failed to load active expedition: %w", err)
	}
	if expedition == nil {
		return fmt.Errorf("no active expedition found")
	}

	r.expedition = expedition
	return nil
}

func (r *REPL) getNextSystem() (*models.RouteJump, error) {
	if r.expedition.CurrentBakedIndex >= len(r.bakedRoute.Jumps)-1 {
		return nil, fmt.Errorf("already at last jump in route")
	}
	return &r.bakedRoute.Jumps[r.expedition.CurrentBakedIndex+1], nil
}

func (r *REPL) findSystemInRoute(query string) (*models.RouteJump, int, error) {
	query = strings.ToLower(query)
	for i, jump := range r.bakedRoute.Jumps {
		if strings.ToLower(jump.SystemName) == query {
			return &jump, i, nil
		}
		// Also check partial matches
		if strings.Contains(strings.ToLower(jump.SystemName), query) {
			return &jump, i, nil
		}
	}
	return nil, -1, fmt.Errorf("system not found in route")
}

func (r *REPL) handleCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	command := strings.ToLower(parts[0])

	switch command {
	case "help", "h":
		r.printHelp()
		return nil

	case "status", "s":
		r.printStatus()
		return nil

	case "jump", "j":
		if len(parts) < 2 {
			return fmt.Errorf("usage: jump <next|detour|system-name>")
		}

		target := strings.ToLower(parts[1])
		switch target {
		case "next", "n":
			next, err := r.getNextSystem()
			if err != nil {
				return err
			}
			fmt.Printf("Jumping to next system: %s (ID: %d)\n", next.SystemName, next.SystemID)
			if err := r.writeJump(next.SystemName, next.SystemID, next.Distance); err != nil {
				return err
			}
			fmt.Println("✓ Jump written to journal")
			return nil

		case "detour", "d":
			detourName := fmt.Sprintf("Detour-%d", time.Now().Unix()%10000)
			detourID := int64(900000 + time.Now().Unix()%100000)
			fmt.Printf("Jumping to detour: %s (ID: %d)\n", detourName, detourID)
			if err := r.writeJump(detourName, detourID, 25.5); err != nil {
				return err
			}
			fmt.Println("✓ Jump written to journal")
			return nil

		default:
			// Try to find system by name
			systemQuery := strings.Join(parts[1:], " ")
			jump, idx, err := r.findSystemInRoute(systemQuery)
			if err != nil {
				return fmt.Errorf("system '%s' not found in route", systemQuery)
			}
			fmt.Printf("Jumping to: %s (ID: %d, index: %d)\n", jump.SystemName, jump.SystemID, idx)
			if err := r.writeJump(jump.SystemName, jump.SystemID, jump.Distance); err != nil {
				return err
			}
			fmt.Println("✓ Jump written to journal")
			return nil
		}

	case "target", "t":
		if len(parts) < 2 {
			return fmt.Errorf("usage: target <next|system-name>")
		}

		targetArg := strings.ToLower(parts[1])
		switch targetArg {
		case "next", "n":
			next, err := r.getNextSystem()
			if err != nil {
				return err
			}
			fmt.Printf("Targeting next system: %s (ID: %d)\n", next.SystemName, next.SystemID)
			if err := r.writeTarget(next.SystemName, next.SystemID, "G"); err != nil {
				return err
			}
			fmt.Println("✓ Target written to journal")
			return nil

		default:
			// Try to find system by name
			systemQuery := strings.Join(parts[1:], " ")
			jump, idx, err := r.findSystemInRoute(systemQuery)
			if err != nil {
				return fmt.Errorf("system '%s' not found in route", systemQuery)
			}
			fmt.Printf("Targeting: %s (ID: %d, index: %d)\n", jump.SystemName, jump.SystemID, idx)
			if err := r.writeTarget(jump.SystemName, jump.SystemID, "G"); err != nil {
				return err
			}
			fmt.Println("✓ Target written to journal")
			return nil
		}

	case "exit", "quit", "q":
		fmt.Println("Exiting...")
		os.Exit(0)
		return nil

	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", command)
	}
}

func (r *REPL) printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  jump next, j n          - Jump to next expected system")
	fmt.Println("  jump detour, j d        - Jump to random detour system")
	fmt.Println("  jump <system>, j <sys>  - Jump to specific system by name")
	fmt.Println("  target next, t n        - Target next expected system")
	fmt.Println("  target <system>, t <sys>- Target specific system by name")
	fmt.Println("  status, s               - Show current expedition status")
	fmt.Println("  help, h                 - Show this help")
	fmt.Println("  exit, quit, q           - Exit REPL")
	fmt.Println()
}

func (r *REPL) printStatus() {
	fmt.Printf("\nExpedition: %s\n", r.expedition.Name)
	fmt.Printf("Status: %s\n", r.expedition.Status)
	fmt.Printf("Current Index: %d/%d\n", r.expedition.CurrentBakedIndex, len(r.bakedRoute.Jumps)-1)
	fmt.Printf("Total Jumps in History: %d\n", len(r.expedition.JumpHistory))

	if r.expedition.CurrentBakedIndex < len(r.bakedRoute.Jumps) {
		current := r.bakedRoute.Jumps[r.expedition.CurrentBakedIndex]
		fmt.Printf("Current System: %s\n", current.SystemName)
	}

	if next, err := r.getNextSystem(); err == nil {
		fmt.Printf("Next System: %s (%.2f LY)\n", next.SystemName, next.Distance)
	} else {
		fmt.Printf("Next System: None (at end of route)\n")
	}
	fmt.Printf("Journal: %s\n", r.journalPath)
	fmt.Println()
}

func (r *REPL) Run() {
	fmt.Println("=== Elite Dangerous Jump REPL ===")
	fmt.Printf("Writing to: %s\n", r.journalPath)
	r.printStatus()
	fmt.Println("Type 'help' for available commands")
	fmt.Println()

	for {
		fmt.Print("> ")
		if !r.scanner.Scan() {
			break
		}

		line := strings.TrimSpace(r.scanner.Text())
		if line == "" {
			continue
		}

		if err := r.handleCommand(line); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	if err := r.scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

func main() {
	journalDir := "."
	if len(os.Args) > 1 {
		journalDir = os.Args[1]
	}

	// Check if ED_EXPEDITION_DATA_DIR is set
	if dataDir := os.Getenv("ED_EXPEDITION_DATA_DIR"); dataDir != "" {
		fmt.Printf("Using data directory: %s\n", dataDir)
	}

	repl, err := NewREPL(journalDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing REPL: %v\n", err)
		os.Exit(1)
	}
	defer repl.Close()

	repl.Run()
}
