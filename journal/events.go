package journal

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	Loadout   EventType = "Loadout"
	FSDJump   EventType = "FSDJump"
	FSDTarget EventType = "FSDTarget"
)

type LoadoutEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	Event         EventType `json:"event"`
	Ship          string    `json:"Ship"`
	ShipID        int       `json:"ShipID"`
	ShipName      string    `json:"ShipName"`
	ShipIdent     string    `json:"ShipIdent"`
	HullValue     int       `json:"HullValue"`
	ModulesValue  int       `json:"ModulesValue"`
	HullHealth    float64   `json:"HullHealth"`
	UnladenMass   float64   `json:"UnladenMass"`
	CargoCapacity int       `json:"CargoCapacity"`
	MaxJumpRange  float64   `json:"MaxJumpRange"`
	FuelCapacity  struct {
		Main    float64 `json:"Main"`
		Reserve float64 `json:"Reserve"`
	} `json:"FuelCapacity"`
	Rebuy   int `json:"Rebuy"`
	Modules []struct {
		Slot         string  `json:"Slot"`
		Item         string  `json:"Item"`
		On           bool    `json:"On"`
		Priority     int     `json:"Priority"`
		Health       float64 `json:"Health"`
		Value        int     `json:"Value,omitempty"`
		AmmoInClip   int     `json:"AmmoInClip,omitempty"`
		AmmoInHopper int     `json:"AmmoInHopper,omitempty"`
		Engineering  struct {
			Engineer      string  `json:"Engineer"`
			EngineerID    int     `json:"EngineerID"`
			BlueprintID   int     `json:"BlueprintID"`
			BlueprintName string  `json:"BlueprintName"`
			Level         int     `json:"Level"`
			Quality       float64 `json:"Quality"`
			Modifiers     []struct {
				Label         string  `json:"Label"`
				Value         float64 `json:"Value"`
				OriginalValue float64 `json:"OriginalValue"`
				LessIsGood    int     `json:"LessIsGood"`
			} `json:"Modifiers"`
		} `json:"Engineering,omitempty"`
	} `json:"Modules"`
}

func LoadoutEventFromJson(data []byte) (*LoadoutEvent, error) {
	loadoutEvent := LoadoutEvent{}
	err := json.Unmarshal(data, &loadoutEvent)
	if err != nil {
		return nil, err
	}

	return &loadoutEvent, nil
}

type FSDJumpEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	Event         EventType `json:"event"`
	Taxi          bool      `json:"Taxi"`
	Multicrew     bool      `json:"Multicrew"`
	StarSystem    string    `json:"StarSystem"`
	SystemAddress int64     `json:"SystemAddress"`
	// [x, y, z], in light years
	StarPos                      []float64 `json:"StarPos"`
	SystemAllegiance             string    `json:"SystemAllegiance"`
	SystemEconomy                string    `json:"SystemEconomy"`
	SystemEconomyLocalised       string    `json:"SystemEconomy_Localised"`
	SystemSecondEconomy          string    `json:"SystemSecondEconomy"`
	SystemSecondEconomyLocalised string    `json:"SystemSecondEconomy_Localised"`
	SystemGovernment             string    `json:"SystemGovernment"`
	SystemGovernmentLocalised    string    `json:"SystemGovernment_Localised"`
	SystemSecurity               string    `json:"SystemSecurity"`
	SystemSecurityLocalised      string    `json:"SystemSecurity_Localised"`
	Population                   int       `json:"Population"`
	Body                         string    `json:"Body"`
	BodyID                       int       `json:"BodyID"`
	BodyType                     string    `json:"BodyType"`
	JumpDist                     float64   `json:"JumpDist"`
	FuelUsed                     float64   `json:"FuelUsed"`
	FuelLevel                    float64   `json:"FuelLevel"`
}

func FSDJumpEventFromJson(data []byte) (*FSDJumpEvent, error) {
	fsdJumpEvent := FSDJumpEvent{}
	err := json.Unmarshal(data, &fsdJumpEvent)
	if err != nil {
		return nil, err
	}

	return &fsdJumpEvent, nil
}

type FSDTargetEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	Event         string    `json:"event"`
	Name          string    `json:"Name"`
	SystemAddress int64     `json:"SystemAddress"`
	StarClass     string    `json:"StarClass"`
}

func FSDTargetEventFromJson(data []byte) (*FSDTargetEvent, error) {
	fsdJumpEvent := FSDTargetEvent{}
	err := json.Unmarshal(data, &fsdJumpEvent)
	if err != nil {
		return nil, err
	}

	return &fsdJumpEvent, nil
}
