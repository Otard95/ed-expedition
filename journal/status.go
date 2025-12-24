package journal

import (
	"ed-expedition/lib/binflag"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
)

type Flags int

// https://elite-journal.readthedocs.io/en/latest/Status%20File.html
const (
	FlagScoopingFuel Flags = 1 << 11
	FlagInMainShip   Flags = 1 << 24
)

type Status struct {
	Timestamp time.Time   `json:"timestamp"`
	Event     string      `json:"event"`
	Flags     *Flags      `json:"Flags"`
	Fuel      *FuelStatus `json:"Fuel"`
}

type FuelStatus struct {
	FuelMain      float64 `json:"FuelMain"`
	FuelReservoir float64 `json:"FuelReservoir"`
}

func (jw *Watcher) handleStatusUpdate() {
	content, err := os.ReadFile(path.Join(jw.dir, "Status.json"))
	if err != nil {
		jw.logger.Error(fmt.Sprintf("Failed to read Status.json: %s", err.Error()))
	}

	var status Status
	err = json.Unmarshal(content, &status)
	if err != nil {
		jw.logger.Error(fmt.Sprintf("Failed to parse Status.json: %s", err.Error()))
	}

	if status.Flags == nil || !binflag.Has(*status.Flags, FlagInMainShip) {
		return
	}

	jw.Scooping.Publish(binflag.Has(*status.Flags, FlagScoopingFuel))

	if status.Fuel != nil {
		jw.Fuel.Publish(status.Fuel)
	}
}
