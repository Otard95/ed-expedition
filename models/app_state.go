package models

import (
	"ed-expedition/database"
	"os"
	"time"
)

type AppState struct {
	LastKnownLoadout  *Loadout  `json:"last_known_loadout,omitempty"`
	LastKnownLocation *Location `json:"last_known_location,omitempty"`
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

// Returns empty index if file doesn't exist
func LoadAppSate() (*AppState, error) {
	path, err := database.AppStatePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &AppState{}, nil
	}

	return database.ReadJSON[AppState](path)
}

func SaveAppState(state *AppState) error {
	path, err := database.AppStatePath()
	if err != nil {
		return err
	}

	return database.WriteJSON(path, state)
}
