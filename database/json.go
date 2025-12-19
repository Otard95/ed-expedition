package database

import (
	"encoding/json"
	"os"
	"strconv"
	"time"
)

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

// Uses temp file + rename to prevent corruption on crash/power loss
func WriteJSON[T any](path string, data T) error {
	// Marshal with indentation for readability
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + strconv.FormatInt(time.Now().UnixNano(), 10) + ".tmp"
	err = os.WriteFile(tmpPath, content, 0644)
	if err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
