package database

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type ModelType string

const (
	ModelTypeExpeditions ModelType = "expeditions"
	ModelTypeRoutes      ModelType = "routes"
)

var (
	DataDir        string
	CacheDir       string
	AppStatePath   string
	IndexPath      string
	BuildStatePath string
	SystemsBinPath string
	SystemsIdxPath string
	NamesBinPath   string
	NamesTriePath  string
)

func init() {
	if err := InitDirectories(); err != nil {
		panic(fmt.Sprintf("failed to init directories: %v", err))
	}
}

// InitDirectories initializes all directory paths. Called automatically at package init,
// but can be called again in tests after setting ED_EXPEDITION_DATA_DIR env var.
func InitDirectories() error {
	var err error

	DataDir, err = initDataDir()
	if err != nil {
		return fmt.Errorf("failed to init data dir: %w", err)
	}

	CacheDir, err = initCacheDir()
	if err != nil {
		return fmt.Errorf("failed to init cache dir: %w", err)
	}

	AppStatePath = filepath.Join(DataDir, "app-state.json")
	IndexPath = filepath.Join(DataDir, "index.json")
	BuildStatePath = filepath.Join(CacheDir, "build.state.json")
	SystemsBinPath = filepath.Join(DataDir, "galaxy", "systems.bin")
	SystemsIdxPath = filepath.Join(DataDir, "galaxy", "systems.idx")
	NamesBinPath = filepath.Join(DataDir, "galaxy", "names.bin")
	NamesTriePath = filepath.Join(DataDir, "galaxy", "names.trie")

	return nil
}

func PathFor(modelType ModelType, id string) string {
	return filepath.Join(DataDir, string(modelType), id+".json")
}

func initDataDir() (string, error) {
	var dataDir string

	if envDataDir := os.Getenv("ED_EXPEDITION_DATA_DIR"); envDataDir != "" {
		dataDir = envDataDir
	} else {
		baseDir, err := getOSDataDir()
		if err != nil {
			return "", err
		}
		dataDir = filepath.Join(baseDir, "ed-expedition")
	}

	dirs := []string{
		filepath.Join(dataDir, string(ModelTypeExpeditions)),
		filepath.Join(dataDir, string(ModelTypeRoutes)),
		filepath.Join(dataDir, "galaxy"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	return dataDir, nil
}

func initCacheDir() (string, error) {
	var cacheDir string

	if envCacheDir := os.Getenv("ED_EXPEDITION_CACHE_DIR"); envCacheDir != "" {
		cacheDir = envCacheDir
	} else {
		baseDir, err := getOSCacheDir()
		if err != nil {
			return "", err
		}
		cacheDir = filepath.Join(baseDir, "ed-expedition")
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return cacheDir, nil
}

func getOSDataDir() (string, error) {
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

func getOSCacheDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		baseDir := os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("APPDATA")
		}
		if baseDir == "" {
			return "", os.ErrNotExist
		}
		return baseDir, nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Caches"), nil
	default:
		baseDir := os.Getenv("XDG_CACHE_HOME")
		if baseDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, ".cache"), nil
		}
		return baseDir, nil
	}
}
