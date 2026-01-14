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
type Flags2 int

// https://elite-journal.readthedocs.io/en/latest/Status%20File.html
const (
	FlagScoopingFuel Flags = 1 << 11
	FlagFsdCharging  Flags = 1 << 17
	FlagInMainShip   Flags = 1 << 24
)

const (
	Flag2HyperdriveCharging Flags2 = 1 << 19
)

type Status struct {
	Timestamp time.Time   `json:"timestamp"`
	Event     string      `json:"event"`
	Flags     *Flags      `json:"Flags"`
	Flags2    *Flags2     `json:"Flags2"`
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

	fsdChargingFlag := binflag.Has(*status.Flags, FlagFsdCharging)
	hyperdriveCFlag := status.Flags2 != nil && binflag.Has(*status.Flags2, Flag2HyperdriveCharging)
	fsdCharging := fsdChargingFlag || hyperdriveCFlag

	fuelStr := "n/a"
	if status.Fuel != nil {
		fuelStr = fmt.Sprintf("%.2f", status.Fuel.FuelMain)
	}
	if fsdChargingFlag != jw.prevFsdChargingFlag {
		jw.logger.Trace(fmt.Sprintf("[FSD_TIMING] FsdCharging flag: %v -> %v (status.timestamp=%v, fuel=%s)",
			jw.prevFsdChargingFlag, fsdChargingFlag, status.Timestamp, fuelStr))
		jw.prevFsdChargingFlag = fsdChargingFlag
	}
	if hyperdriveCFlag != jw.prevHyperdriveCFlag {
		jw.logger.Trace(fmt.Sprintf("[FSD_TIMING] HyperdriveCharging flag: %v -> %v (status.timestamp=%v, fuel=%s)",
			jw.prevHyperdriveCFlag, hyperdriveCFlag, status.Timestamp, fuelStr))
		jw.prevHyperdriveCFlag = hyperdriveCFlag
	}

	jw.logger.Trace(fmt.Sprintf("handleStatusUpdate: publishing fsdCharging=%v", fsdCharging))
	jw.FsdCharging.Publish(fsdCharging)

	if status.Fuel != nil {
		jw.logger.Trace(fmt.Sprintf("handleStatusUpdate: publishing fuel main=%.2f reservoir=%.2f", status.Fuel.FuelMain, status.Fuel.FuelReservoir))
		jw.Fuel.Publish(status.Fuel)
	}
}
