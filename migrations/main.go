package migrations

import "fmt"

type MigrationFunc func(data map[string]any) error

type Registry []MigrationFunc

func (r Registry) LatestVersion() int {
	return len(r)
}

func Migrate(data map[string]any, registry Registry) (mitraged bool, err error) {
	version := 0
	if v, ok := data["version"]; ok {
		if vFloat, ok := v.(float64); ok {
			version = int(vFloat)
		}
	}

	if version >= len(registry) {
		return false, nil
	}

	for i := version; i < len(registry); i++ {
		if err := registry[i](data); err != nil {
			return false, fmt.Errorf("migration v%d → v%d failed: %w", i, i+1, err)
		}
	}

	return true, nil
}
