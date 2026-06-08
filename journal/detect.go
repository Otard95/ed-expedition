package journal

import (
	"os"
	"path/filepath"
	"runtime"
)

// edProtonSuffix is the fixed path from a Steam root to the ED journal directory.
const edProtonSuffix = "steamapps/compatdata/359320/pfx/drive_c/users/steamuser/Saved Games/Frontier Developments/Elite Dangerous"

func journalDirCandidates() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		return []string{
			filepath.Join(home, "Saved Games", "Frontier Developments", "Elite Dangerous"),
		}
	case "linux":
		steamRoots := []string{
			filepath.Join(home, ".local", "share", "Steam"),
			filepath.Join(home, ".steam", "steam"),
			filepath.Join(home, ".steam", "debian-installation"),
			filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam"),
		}
		candidates := make([]string, len(steamRoots))
		for i, root := range steamRoots {
			candidates[i] = filepath.Join(root, edProtonSuffix)
		}
		return candidates
	default:
		return nil
	}
}

// DetectJournalDir returns the first auto-detected ED journal directory that
// exists on the system. Returns empty string if none found.
func DetectJournalDir() string {
	for _, dir := range journalDirCandidates() {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}
