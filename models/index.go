package models

import (
	"ed-expedition/database"
	"os"
	"time"
)

// Index tracks all expeditions and the currently active one
type Index struct {
	ActiveExpeditionID *string             `json:"active_expedition_id"`
	Expeditions        []ExpeditionSummary `json:"expeditions"`
}

// ExpeditionSummary provides overview info for expedition listing
type ExpeditionSummary struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Status      ExpeditionStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	LastUpdated time.Time        `json:"last_updated"`
}

// LoadFull loads the complete expedition data from disk
func (summary *ExpeditionSummary) LoadFull() (*Expedition, error) {
	expedition, err := LoadExpedition(summary.ID)
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

// LoadIndex loads the expedition index from disk
// Returns empty index if file doesn't exist
func LoadIndex() (*Index, error) {
	path, err := database.IndexPath()
	if err != nil {
		return nil, err
	}

	// If index doesn't exist, return empty index
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Index{
			ActiveExpeditionID: nil,
			Expeditions:        []ExpeditionSummary{},
		}, nil
	}

	return database.ReadJSON[Index](path)
}

// SaveIndex saves the expedition index to disk
func SaveIndex(index *Index) error {
	path, err := database.IndexPath()
	if err != nil {
		return err
	}

	return database.WriteJSON(path, index)
}
