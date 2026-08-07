// set-region-coords walks every region directory and prompts for an optional
// name-position override. Paste the JSON copied from the map debug overlay
// (e.g. {"x":500,"z":1200}) or press Enter to leave unchanged / clear.
//
// Usage:
//
//	go run ./cmd/set-region-coords <regions-dir>
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: set-region-coords <regions-dir>")
		os.Exit(1)
	}
	regionsDir := os.Args[1]

	entries, err := os.ReadDir(regionsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(regionsDir, e.Name())

		// Read display name
		name := e.Name()
		if b, err := os.ReadFile(filepath.Join(dir, "name.txt")); err == nil {
			name = strings.TrimSpace(string(b))
		}

		// Show existing override if present
		existing := ""
		if b, err := os.ReadFile(filepath.Join(dir, "name-coords.json")); err == nil {
			existing = " [" + strings.TrimSpace(string(b)) + "]"
		}

		fmt.Printf("\n%s%s\n  Paste XZ JSON (Enter = skip, 'c' = clear): ", name, existing)

		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		coordsPath := filepath.Join(dir, "name-coords.json")

		if input == "c" {
			os.Remove(coordsPath)
			fmt.Println("  Cleared.")
			continue
		}

		// Validate — must parse as {x, z}
		var coords struct {
			X float64 `json:"x"`
			Z float64 `json:"z"`
		}
		if err := json.Unmarshal([]byte(input), &coords); err != nil {
			fmt.Printf("  Invalid JSON (%v), skipping.\n", err)
			continue
		}

		out, _ := json.Marshal(coords)
		if err := os.WriteFile(coordsPath, out, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  Write error: %v\n", err)
			continue
		}
		fmt.Printf("  Saved: x=%.0f z=%.0f\n", coords.X, coords.Z)
	}

	fmt.Println("\nDone.")
}
