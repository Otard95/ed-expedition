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
	jw.logger.Trace("handleStatusUpdate called")

	content, err := os.ReadFile(path.Join(jw.dir, "Status.json"))
	if err != nil {
		jw.logger.Error(fmt.Sprintf("Failed to read Status.json: %s", err.Error()))
		return
	}

	var status Status
	err = json.Unmarshal(content, &status)
	if err != nil {
		jw.logger.Error(fmt.Sprintf("Failed to parse Status.json: %s", err.Error()))
		return
	}

	if status.Flags == nil {
		jw.logger.Trace("handleStatusUpdate: Flags is nil, skipping")
		return
	}

	if !binflag.Has(*status.Flags, FlagInMainShip) {
		jw.logger.Trace("handleStatusUpdate: not in main ship, skipping")
		return
	}

	scooping := binflag.Has(*status.Flags, FlagScoopingFuel)
	jw.logger.Trace(fmt.Sprintf("handleStatusUpdate: publishing scooping=%v", scooping))
	jw.Scooping.Publish(scooping)

	if status.Fuel != nil {
		jw.logger.Trace(fmt.Sprintf("handleStatusUpdate: publishing fuel main=%.2f reservoir=%.2f", status.Fuel.FuelMain, status.Fuel.FuelReservoir))
		jw.Fuel.Publish(status.Fuel)
	}
}
