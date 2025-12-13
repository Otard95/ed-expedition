package database

import (
	"os"
	"path/filepath"
	"runtime"
)

// ModelType represents a type of model stored in the database
type ModelType string

const (
	ModelTypeExpeditions ModelType = "expeditions"
	ModelTypeRoutes      ModelType = "routes"
)

// PathFor constructs a path for a given model type and ID
// Example: PathFor(ModelTypeExpeditions, "abc-123") -> ~/.local/share/ed-expedition/expeditions/abc-123.json
func PathFor(modelType ModelType, id string) (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, string(modelType), id+".json"), nil
}

// GetOSDataDir returns the OS-specific user data directory
// - Windows: %APPDATA%
// - macOS: ~/Library/Application Support
// - Linux: ~/.local/share (respects XDG_DATA_HOME)
func GetOSDataDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%
		baseDir := os.Getenv("APPDATA")
		if baseDir == "" {
			return "", os.ErrNotExist
		}
		return baseDir, nil
	case "darwin":
		// macOS: ~/Library/Application Support
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
		// Linux and others: ~/.local/share (XDG_DATA_HOME)
		baseDir := os.Getenv("XDG_DATA_HOME")
		if baseDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, ".local", "share"), nil
		}
		return baseDir, nil
	}
}

// GetDataDir returns the application data directory
// Creates it if it doesn't exist
func GetDataDir() (string, error) {
	baseDir, err := GetOSDataDir()
	if err != nil {
		return "", err
	}

	// Application-specific subdirectory
	dataDir := filepath.Join(baseDir, "ed-expedition")

	// Create directories if they don't exist
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, string(ModelTypeExpeditions)),
		filepath.Join(dataDir, string(ModelTypeRoutes)),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	return dataDir, nil
}

// IndexPath returns the path to index.json
func IndexPath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "index.json"), nil
}
