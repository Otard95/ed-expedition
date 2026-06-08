package migrations

var RouteMigrations = Registry{
	migrateRouteV0ToV1,
}

func migrateRouteV0ToV1(data map[string]any) error {
	jumps, ok := data["jumps"].([]any)
	if !ok {
		data["version"] = 1
		return nil
	}

	for _, j := range jumps {
		jump, ok := j.(map[string]any)
		if !ok {
			continue
		}

		if hasNeutron, ok := jump["has_neutron"].(bool); ok && hasNeutron {
			jump["fsd_boost"] = float64(1)
		}
		delete(jump, "has_neutron")
	}

	data["version"] = 1
	return nil
}
