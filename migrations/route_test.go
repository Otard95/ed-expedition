package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteV0ToV1_NeutronTrue(t *testing.T) {
	data := routeV0([]map[string]any{
		{"system_name": "Sol", "has_neutron": true},
	})

	migrated, err := Migrate(data, RouteMigrations)

	require.NoError(t, err)
	assert.True(t, migrated)

	jump := getJump(t, data, 0)
	assert.Equal(t, float64(1), jump["fsd_boost"])
	assert.Nil(t, jump["has_neutron"])
}

func TestRouteV0ToV1_NeutronFalse(t *testing.T) {
	data := routeV0([]map[string]any{
		{"system_name": "Sol", "has_neutron": false},
	})

	migrated, err := Migrate(data, RouteMigrations)

	require.NoError(t, err)
	assert.True(t, migrated)

	jump := getJump(t, data, 0)
	assert.Nil(t, jump["fsd_boost"])
	assert.Nil(t, jump["has_neutron"])
}

func TestRouteV0ToV1_NeutronAbsent(t *testing.T) {
	data := routeV0([]map[string]any{
		{"system_name": "Sol"},
	})

	migrated, err := Migrate(data, RouteMigrations)

	require.NoError(t, err)
	assert.True(t, migrated)

	jump := getJump(t, data, 0)
	assert.Nil(t, jump["fsd_boost"])
	assert.Nil(t, jump["has_neutron"])
}

func TestRouteV0ToV1_MultipleJumps(t *testing.T) {
	data := routeV0([]map[string]any{
		{"system_name": "Sol", "has_neutron": false},
		{"system_name": "Neutron A", "has_neutron": true},
		{"system_name": "Alpha Centauri"},
		{"system_name": "Neutron B", "has_neutron": true},
	})

	migrated, err := Migrate(data, RouteMigrations)

	require.NoError(t, err)
	assert.True(t, migrated)

	assert.Nil(t, getJump(t, data, 0)["fsd_boost"])
	assert.Equal(t, float64(1), getJump(t, data, 1)["fsd_boost"])
	assert.Nil(t, getJump(t, data, 2)["fsd_boost"])
	assert.Equal(t, float64(1), getJump(t, data, 3)["fsd_boost"])

	for i := range 4 {
		assert.Nil(t, getJump(t, data, i)["has_neutron"])
	}
}

func TestRouteV0ToV1_NoJumps(t *testing.T) {
	data := routeV0([]map[string]any{})

	migrated, err := Migrate(data, RouteMigrations)

	require.NoError(t, err)
	assert.True(t, migrated)
	assert.Equal(t, 1, data["version"])
}

func TestRouteV0ToV1_MissingJumpsField(t *testing.T) {
	data := map[string]any{"id": "test-route"}

	migrated, err := Migrate(data, RouteMigrations)

	require.NoError(t, err)
	assert.True(t, migrated)
	assert.Equal(t, 1, data["version"])
}

func TestRouteV0ToV1_PreservesOtherFields(t *testing.T) {
	data := routeV0([]map[string]any{
		{
			"system_name": "Sol",
			"system_id":   float64(123),
			"distance":    float64(45.5),
			"scoopable":   true,
			"has_neutron":  true,
		},
	})

	_, err := Migrate(data, RouteMigrations)
	require.NoError(t, err)

	jump := getJump(t, data, 0)
	assert.Equal(t, "Sol", jump["system_name"])
	assert.Equal(t, float64(123), jump["system_id"])
	assert.Equal(t, float64(45.5), jump["distance"])
	assert.Equal(t, true, jump["scoopable"])
}

func TestRouteV0ToV1_AlreadyV1(t *testing.T) {
	data := map[string]any{
		"version": float64(1),
		"jumps": []any{
			map[string]any{"system_name": "Sol", "has_neutron": true},
		},
	}

	migrated, err := Migrate(data, RouteMigrations)

	require.NoError(t, err)
	assert.False(t, migrated)
	assert.Equal(t, true, getJump(t, data, 0)["has_neutron"])
}

// --- helpers ---

func routeV0(jumps []map[string]any) map[string]any {
	jumpsAny := make([]any, len(jumps))
	for i, j := range jumps {
		jumpsAny[i] = j
	}
	return map[string]any{
		"id":    "test-route",
		"name":  "Test Route",
		"jumps": jumpsAny,
	}
}

func getJump(t *testing.T, data map[string]any, index int) map[string]any {
	t.Helper()
	jumps, ok := data["jumps"].([]any)
	require.True(t, ok, "jumps field is not []any")
	require.Greater(t, len(jumps), index, "jump index out of range")
	jump, ok := jumps[index].(map[string]any)
	require.True(t, ok, "jump is not map[string]any")
	return jump
}
