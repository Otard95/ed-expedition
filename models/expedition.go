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
	Status      ExpeditionStatus `json:"status"`

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
	OnRoute    bool      `json:"on_route"`            // System exists in baked route
	Expected   bool      `json:"expected"`            // Was next expected jump
	Synthetic  bool      `json:"synthetic,omitempty"` // Added to fill gaps (app offline)
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

func DeleteExpedition(id string) error {
	path, err := database.PathFor(database.ModelTypeExpeditions, id)
	if err != nil {
		return err
	}

	return os.Remove(path)
}
