package services

import (
	"ed-expedition/journal"
	"ed-expedition/models"
	"fmt"
	"strconv"
	"time"
)

type FuelAlertLevel uint8

const (
	FuelLevelInfo FuelAlertLevel = iota
	FuelLevelOk
	FuelLevelWarn
	FuelLevelCritical
)

type FuelAlert struct {
	Level   FuelAlertLevel `json:"level"`
	Message string         `json:"message"`
}

func (e *ExpeditionService) handleRefueling(scooping bool) {
	if e.activeExpedition != nil && e.previouslyScooping && !scooping {
		go func() {
			time.Sleep(time.Second)

			if e.activeExpedition == nil {
				return
			}

			err := models.SaveExpedition(e.activeExpedition)
			if err != nil {
				e.logger.Error(fmt.Sprintf("Failed to save expedition after refueling: %s", err.Error()))
			}
		}()
	}
	e.previouslyScooping = scooping
}

func (e *ExpeditionService) handleFuelChange(fuel *journal.FuelStatus) {
	e.logger.Trace(fmt.Sprintf("[ExpeditionService](Fuel) handleFuelChange: fuel=%.2f, jumpInProgress=%v", fuel.FuelMain, e.isJumpInProgress()))

	if e.currentJump != nil {
		if e.isJumpInProgress() {
			e.logger.Trace(fmt.Sprintf("[ExpeditionService](Fuel) handleFuelChange: jump in progress, skipping update to current jump '%s'", e.currentJump.SystemName))
		} else {
			e.logger.Trace(fmt.Sprintf("handleFuelChange: update fuel in tank of current jump '%s' to %f", e.currentJump.SystemName, fuel.FuelMain))
			e.currentJump.FuelLevel = fuel.FuelMain
			e.CurrentJump.Publish(e.currentJump)
		}
	}

	if e.activeExpedition == nil || e.bakedRoute == nil || e.currentJump == nil {
		e.logger.Trace("handleFuelChange: no active expedition/route/jump, skipping")
		return
	}

	if e.currentJump.BakedIndex == nil {
		e.logger.Trace("handleFuelChange: off route, publishing ok with message")
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelInfo,
			Message: "You're off route. You're on your own. Good luck commander o7",
		})
		return
	}

	// TODO: Add a setting to enable this check
	// if e.bakedRoute.Jumps[*e.currentJump.BakedIndex].Scoopable {
	// 	e.logger.Trace("handleFuelChange: current system is scoopable, clearing alert")
	// 	e.FuelAlert.Publish(&FuelAlert{
	// 		Level:   FuelLevelOk,
	// 		Message: "",
	// 	})
	// 	return
	// }

	currentFuel := fuel.FuelMain
	for _, jump := range e.bakedRoute.Jumps[(*e.currentJump.BakedIndex)+1:] {
		if jump.FuelUsed == nil {
			e.logger.Trace("handleFuelChange: jump has no fuel data, skipping fuel check")
			return
		}
		currentFuel -= *jump.FuelUsed
		if jump.Scoopable {
			break
		}
	}

	e.logger.Trace(fmt.Sprintf("handleFuelChange: projected fuel at next scoopable=%.2f", currentFuel))

	// TODO: Disable with setting mentioned above
	if currentFuel < 0.1 && e.bakedRoute.Jumps[*e.currentJump.BakedIndex].Scoopable {
		e.logger.Trace("handleFuelChange: publishing must refuel warning")
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelWarn,
			Message: "Remember to refuel before you go",
		})
	} else if currentFuel < 0.1 {
		e.logger.Trace("handleFuelChange: publishing critical alert")
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelCritical,
			Message: "You will run out of fuel before the next scoopable system",
		})
	} else if currentFuel < 1 {
		e.logger.Trace("handleFuelChange: publishing warn alert")
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelWarn,
			Message: fmt.Sprintf("You'll arrive at the next scoopable system with %st fuel left.", strconv.FormatFloat(currentFuel, 'f', 1, 64)),
		})
	} else {
		e.logger.Trace("handleFuelChange: publishing ok (no message)")
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelOk,
			Message: "Fuel levels at required levels",
		})
	}
}
