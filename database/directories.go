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

func GetOSCacheDir() (string, error) {
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

func GetDataDir() (string, error) {
	var dataDir string

	if envDataDir := os.Getenv("ED_EXPEDITION_DATA_DIR"); envDataDir != "" {
		dataDir = envDataDir
	} else {
		baseDir, err := GetOSDataDir()
		if err != nil {
			return "", err
		}
		dataDir = filepath.Join(baseDir, "ed-expedition")
	}

	dirs := []string{
		dataDir,
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

func GetCacheDir() (string, error) {
	var cacheDir string

	if envCacheDir := os.Getenv("ED_EXPEDITION_CACHE_DIR"); envCacheDir != "" {
		cacheDir = envCacheDir
	} else {
		baseDir, err := GetOSCacheDir()
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

func BuildStatePath() (string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "build.state.json"), nil
}

func SystemsBinPath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "galaxy", "systems.bin"), nil
}

func SystemsIdxPath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "galaxy", "systems.idx"), nil
}

func NamesBinPath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "galaxy", "names.bin"), nil
}

func NamesTriePath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "galaxy", "names.trie"), nil
}
