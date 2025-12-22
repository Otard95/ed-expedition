package models

import (
	"ed-expedition/database"
	"os"
	"slices"
	"time"
)

// ExpeditionStatus represents the state of an expedition
type ExpeditionStatus string

const (
	StatusPlanned   ExpeditionStatus = "planned"
	StatusActive    ExpeditionStatus = "active"
	StatusCompleted ExpeditionStatus = "completed"
	StatusEnded     ExpeditionStatus = "ended"
)

// Expedition represents a journey through connected routes
type Expedition struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	CreatedAt   time.Time        `json:"created_at"`
	LastUpdated time.Time        `json:"last_updated"`
	Status      ExpeditionStatus `json:"status" ts_type:"'planned'|'active'|'completed'|'ended'"`

	StartedOn time.Time `json:"started_on,omitempty"`
	EndedOn   time.Time `json:"ended_on,omitempty"`

	// Start point (can be mid-route)
	Start *RoutePosition `json:"start,omitempty"`

	// Route library (IDs of routes used in this expedition)
	Routes []string `json:"routes"`

	Links []Link `json:"links"`

	// Baked route (only exists when active/completed/ended)
	BakedRouteID       *string `json:"baked_route_id,omitempty"`
	CurrentBakedIndex  int     `json:"current_baked_index"`
	BakedLoopBackIndex *int    `json:"baked_loop_back_index,omitempty"`

	JumpHistory []JumpHistoryEntry `json:"jump_history"`
}

func (e *Expedition) IsEditable() bool {
	return e.Status == StatusPlanned
}

type RoutePosition struct {
	RouteID   string `json:"route_id"`
	JumpIndex int    `json:"jump_index"`
}

func (routePos *RoutePosition) Equal(otherPos *RoutePosition) bool {
	return routePos.RouteID == otherPos.RouteID && routePos.JumpIndex == otherPos.JumpIndex
}
func (routePos *RoutePosition) Clone() *RoutePosition {
	return &RoutePosition{RouteID: routePos.RouteID, JumpIndex: routePos.JumpIndex}
}

// Link connects two routes at an identical system
type Link struct {
	ID   string        `json:"id"`
	From RoutePosition `json:"from"`
	To   RoutePosition `json:"to"`
}

// JumpHistoryEntry records a single jump taken during expedition
type JumpHistoryEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	SystemName string    `json:"system_name"`
	SystemID   int64     `json:"system_id"`
	BakedIndex *int      `json:"baked_index,omitempty"`

	Distance  float64 `json:"distance"`
	FuelUsed  float64 `json:"fuel_used"`
	FuelLevel float64 `json:"fuel_in_tank"`

	Expected  bool `json:"expected"`
	Synthetic bool `json:"synthetic"`
}

func (expedition *Expedition) LoadRoutes() ([]*Route, error) {
	result := make([]*Route, len(expedition.Routes))

	for i, routeId := range expedition.Routes {
		route, err := LoadRoute(routeId)
		if err != nil {
			return nil, err
		}
		result[i] = route
	}

	return result, nil
}

func (expedition *Expedition) LoadBaked() (*Route, error) {
	if expedition.BakedRouteID == nil {
		return nil, nil
	}

	route, err := LoadRoute(*expedition.BakedRouteID)
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (expedition *Expedition) HasRoute(routeId string) bool {
	return slices.Contains(expedition.Routes, routeId)
}

func LoadExpedition(id string) (*Expedition, error) {
	path, err := database.PathFor(database.ModelTypeExpeditions, id)
	if err != nil {
		return nil, err
	}

	return database.ReadJSON[Expedition](path)
}

func SaveExpedition(expedition *Expedition) error {
	path, err := database.PathFor(database.ModelTypeExpeditions, expedition.ID)
	if err != nil {
		return err
	}

	return database.WriteJSON(path, expedition)
}

func TSaveExpedition(t *database.Transaction, expedition *Expedition) error {
	path, err := database.PathFor(database.ModelTypeExpeditions, expedition.ID)
	if err != nil {
		return err
	}

	return t.WriteJSON(path, expedition)
}

func DeleteExpedition(id string) error {
	path, err := database.PathFor(database.ModelTypeExpeditions, id)
	if err != nil {
		return err
	}

	return os.Remove(path)
}
