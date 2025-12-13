package models

import (
	"ed-expedition/database"
	"os"
	"time"
)

// ExpeditionIndex tracks all expeditions and the currently active one
type ExpeditionIndex struct {
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

func (summary *ExpeditionSummary) LoadFull() (*Expedition, error) {
	expedition, err := LoadExpedition(summary.ID)
	if err != nil {
		return nil, err
	}

	return expedition, nil
}

// Returns empty index if file doesn't exist
func LoadIndex() (*ExpeditionIndex, error) {
	path, err := database.IndexPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &ExpeditionIndex{
			ActiveExpeditionID: nil,
			Expeditions:        []ExpeditionSummary{},
		}, nil
	}

	return database.ReadJSON[ExpeditionIndex](path)
}

func SaveIndex(index *ExpeditionIndex) error {
	path, err := database.IndexPath()
	if err != nil {
		return err
	}

	return database.WriteJSON(path, index)
}

func (e *ExpeditionIndex) LoadActiveExpedition() (*Expedition, error) {
	if e.ActiveExpeditionID == nil {
		return nil, nil
	}

	expedition, err := LoadExpedition(*e.ActiveExpeditionID)
	if err != nil {
		return nil, err
	}

	return expedition, nil

}
