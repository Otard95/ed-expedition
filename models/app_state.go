package models

import (
	"ed-expedition/database"
	"os"
	"time"
)

type GalaxyDecision string

const (
	GalaxyNotAsked GalaxyDecision = "not_asked"
	GalaxyDeclined GalaxyDecision = "declined"
	GalaxyAccepted GalaxyDecision = "accepted"
)

type AppState struct {
	LastKnownLoadout  *Loadout  `json:"last_known_loadout,omitempty"`
	LastKnownLocation *Location `json:"last_known_location,omitempty"`

	GalaxyDecision     GalaxyDecision `json:"galaxy_decision,omitempty"`
	GalaxyDownloadedAt *time.Time     `json:"galaxy_downloaded_at,omitempty"`
}

type Loadout struct {
	Timestamp    time.Time    `json:"timestamp"`
	UnladenMass  float64      `json:"unladen_mass"`
	FuelCapacity FuelCapacity `json:"fuel_capacity"`
	FSD          struct {
		Item           string   `json:"item"`
		OptimalMass    *float64 `json:"optimal_mass,omitempty"`
		MaxFuelPerJump *float64 `json:"max_fuel_per_jump,omitempty"`
	} `json:"fsd"`
	FSDBooster *string `json:"fsd_booster,omitempty"`
}

type FuelCapacity struct {
	Main    float64 `json:"main"`
	Reserve float64 `json:"reserve"`
}

type LoadoutFSD struct {
	Item           string   `json:"item"`
	OptimalMass    *float64 `json:"optimal_mass,omitempty"`
	MaxFuelPerJump *float64 `json:"max_fuel_per_jump,omitempty"`
}

type Location struct {
	Timestamp time.Time `json:"timestamp"`
	SystemID  int64     `json:"system_id"`
}

// Returns empty state if file doesn't exist
func LoadAppState() (*AppState, error) {
	if _, err := os.Stat(database.AppStatePath); os.IsNotExist(err) {
		return &AppState{}, nil
	}

	state, err := database.ReadJSON[AppState](database.AppStatePath)
	if err != nil {
		return nil, err
	}

	if state.GalaxyDecision == "" {
		state.GalaxyDecision = GalaxyNotAsked
	}

	return state, nil
}

func SaveAppState(state *AppState) error {
	return database.WriteJSON(database.AppStatePath, state)
}

func TSaveAppState(t *database.Transaction, state *AppState) error {
	return t.WriteJSON(database.AppStatePath, state)
}
