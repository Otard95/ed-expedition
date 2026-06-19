package models

import (
	"ed-expedition/database"
	"ed-expedition/migrations"
	"os"
)

type GalaxyDecision string

const (
	GalaxyNotAsked GalaxyDecision = "not_asked"
	GalaxyDeclined GalaxyDecision = "declined"
	GalaxyAccepted GalaxyDecision = "accepted"
)

type Settings struct {
	JournalDir     *string        `json:"journal_dir,omitempty"`
	GalaxyDecision GalaxyDecision `json:"galaxy_decision,omitempty"`
	Debug          bool           `json:"debug,omitempty"`
}

func LoadSettings() (*Settings, error) {
	if _, err := os.Stat(database.SettingsPath); os.IsNotExist(err) {
		return migrateSettingsFromAppState()
	}

	return database.ReadJSON[Settings](database.SettingsPath)
}

// migrateSettingsFromAppState runs once on first launch after settings.json is
// introduced. Carries over fields that used to live in app-state.json so
// existing users don't lose their galaxy decision or journal dir.
func migrateSettingsFromAppState() (*Settings, error) {
	settings := &Settings{GalaxyDecision: GalaxyNotAsked}

	if _, err := os.Stat(database.AppStatePath); os.IsNotExist(err) {
		return settings, nil
	}

	migrated, err := database.ReadAndMigrateJSON[Settings](database.AppStatePath, migrations.AppStateMigrations)
	if err != nil {
		return settings, nil
	}
	settings = migrated
	if settings.GalaxyDecision == "" {
		settings.GalaxyDecision = GalaxyNotAsked
	}

	_ = SaveSettings(settings)
	return settings, nil
}

func SaveSettings(settings *Settings) error {
	return database.WriteJSON(database.SettingsPath, settings)
}
