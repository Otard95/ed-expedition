package migrations

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrate_NoVersion_AppliesAll(t *testing.T) {
	data := map[string]any{"name": "test"}
	registry := Registry{
		func(d map[string]any) error { d["v1"] = true; d["version"] = 1; return nil },
		func(d map[string]any) error { d["v2"] = true; d["version"] = 2; return nil },
	}

	migrated, err := Migrate(data, registry)

	require.NoError(t, err)
	assert.True(t, migrated)
	assert.Equal(t, 2, data["version"])
	assert.Equal(t, true, data["v1"])
	assert.Equal(t, true, data["v2"])
}

func TestMigrate_VersionZero_AppliesAll(t *testing.T) {
	data := map[string]any{"version": float64(0), "name": "test"}
	registry := Registry{
		func(d map[string]any) error { d["v1"] = true; d["version"] = 1; return nil },
		func(d map[string]any) error { d["v2"] = true; d["version"] = 2; return nil },
	}

	migrated, err := Migrate(data, registry)

	require.NoError(t, err)
	assert.True(t, migrated)
	assert.Equal(t, 2, data["version"])
	assert.Equal(t, true, data["v1"])
	assert.Equal(t, true, data["v2"])
}

func TestMigrate_PartiallyMigrated_AppliesRemaining(t *testing.T) {
	data := map[string]any{"version": float64(1)}
	calls := 0
	registry := Registry{
		func(d map[string]any) error { calls++; d["version"] = 1; return nil },
		func(d map[string]any) error { d["v2"] = true; d["version"] = 2; return nil },
		func(d map[string]any) error { d["v3"] = true; d["version"] = 3; return nil },
	}

	migrated, err := Migrate(data, registry)

	require.NoError(t, err)
	assert.True(t, migrated)
	assert.Equal(t, 3, data["version"])
	assert.Equal(t, 0, calls)
	assert.Equal(t, true, data["v2"])
	assert.Equal(t, true, data["v3"])
}

func TestMigrate_AlreadyCurrent_NoOp(t *testing.T) {
	data := map[string]any{"version": float64(2), "name": "test"}
	registry := Registry{
		func(d map[string]any) error { d["v1"] = true; d["version"] = 1; return nil },
		func(d map[string]any) error { d["v2"] = true; d["version"] = 2; return nil },
	}

	migrated, err := Migrate(data, registry)

	require.NoError(t, err)
	assert.False(t, migrated)
	assert.Nil(t, data["v1"])
	assert.Nil(t, data["v2"])
}

func TestMigrate_AheadOfRegistry_NoOp(t *testing.T) {
	data := map[string]any{"version": float64(5)}
	registry := Registry{
		func(d map[string]any) error { d["version"] = 1; return nil },
	}

	migrated, err := Migrate(data, registry)

	require.NoError(t, err)
	assert.False(t, migrated)
	assert.Equal(t, float64(5), data["version"])
}

func TestMigrate_EmptyRegistry_NoOp(t *testing.T) {
	data := map[string]any{"name": "test"}
	registry := Registry{}

	migrated, err := Migrate(data, registry)

	require.NoError(t, err)
	assert.False(t, migrated)
}

func TestMigrate_MigrationError_StopsAndReturns(t *testing.T) {
	data := map[string]any{"name": "test"}
	registry := Registry{
		func(d map[string]any) error { d["v1"] = true; d["version"] = 1; return nil },
		func(d map[string]any) error { return fmt.Errorf("disk full") },
		func(d map[string]any) error { d["v3"] = true; d["version"] = 3; return nil },
	}

	migrated, err := Migrate(data, registry)

	assert.Error(t, err)
	assert.False(t, migrated)
	assert.Contains(t, err.Error(), "v1 → v2")
	assert.Contains(t, err.Error(), "disk full")
	assert.Equal(t, true, data["v1"])
	assert.Nil(t, data["v3"])
}

func TestMigrate_SequentialExecution(t *testing.T) {
	var order []int
	data := map[string]any{}
	registry := Registry{
		func(d map[string]any) error { order = append(order, 0); d["version"] = 1; return nil },
		func(d map[string]any) error { order = append(order, 1); d["version"] = 2; return nil },
		func(d map[string]any) error { order = append(order, 2); d["version"] = 3; return nil },
	}

	_, err := Migrate(data, registry)

	require.NoError(t, err)
	assert.Equal(t, []int{0, 1, 2}, order)
}

func TestRegistryLatestVersion(t *testing.T) {
	assert.Equal(t, 0, Registry{}.LatestVersion())
	assert.Equal(t, 1, Registry{nil}.LatestVersion())
	assert.Equal(t, 3, Registry{nil, nil, nil}.LatestVersion())
}
