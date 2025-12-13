package database

import (
	"os"
	"path/filepath"
	"runtime"
)

type ModelType string

const (
	ModelTypeExpeditions ModelType = "expeditions"
	ModelTypeRoutes      ModelType = "routes"
)

func PathFor(modelType ModelType, id string) (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, string(modelType), id+".json"), nil
}

func GetOSDataDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		baseDir := os.Getenv("APPDATA")
		if baseDir == "" {
			return "", os.ErrNotExist
		}
		return baseDir, nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
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

func GetDataDir() (string, error) {
	baseDir, err := GetOSDataDir()
	if err != nil {
		return "", err
	}

	dataDir := filepath.Join(baseDir, "ed-expedition")

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

func AppStatePath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "app-state.json"), nil
}

func IndexPath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "index.json"), nil
}
