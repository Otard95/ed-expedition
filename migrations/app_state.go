package migrations

var AppStateMigrations = Registry{
	migrateAppStateV0ToV1,
}

func migrateAppStateV0ToV1(data map[string]any) error {
	gd, _ := data["galaxy_decision"].(string)
	if gd != "not_asked" && gd != "declined" && gd != "accepted" {
		data["galaxy_decision"] = "not_asked"
	}

	if loc, ok := data["last_known_location"].(map[string]any); ok {
		if ts, ok := loc["timestamp"].(string); ok {
			data["journal_sync"] = map[string]any{
				"timestamp":  ts,
				"event_hash": "",
			}
		}
	}

	data["version"] = 1
	return nil
}
