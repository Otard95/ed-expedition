package database

import (
	"encoding/json"
	"os"
)

// ReadJSON reads and unmarshals a JSON file into type T
func ReadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result T
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// WriteJSON marshals data to JSON and writes atomically to path
// Uses temp file + rename for atomic writes (no corruption on crash)
func WriteJSON[T any](path string, data T) error {
	// Marshal with indentation for readability
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first
	tmpPath := path + ".tmp"
	err = os.WriteFile(tmpPath, content, 0644)
	if err != nil {
		return err
	}

	// Atomic rename (replaces old file)
	return os.Rename(tmpPath, path)
}
