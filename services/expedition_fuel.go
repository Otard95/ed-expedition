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
	FuelLevelOk FuelAlertLevel = iota
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
	if e.currentJump != nil {
		e.currentJump.FuelLevel = fuel.FuelMain
		e.CurrentJump.Publish(e.currentJump)
	}

	if e.activeExpedition == nil || e.bakedRoute == nil || e.currentJump == nil {
		return
	}

	if e.currentJump.BakedIndex == nil {
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelOk,
			Message: "You're off route. You're on your own. Good luck commander!",
		})
	}

	if e.bakedRoute.Jumps[*e.currentJump.BakedIndex].Scoopable {
		return
	}

	currentFuel := fuel.FuelMain
	for _, jump := range e.bakedRoute.Jumps[(*e.currentJump.BakedIndex)+1:] {
		if jump.FuelUsed == nil {
			return
		}
		currentFuel -= *jump.FuelUsed
		if jump.Scoopable {
			break
		}
	}

	if currentFuel < 0 {
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelCritical,
			Message: "You will run out of fuel before the next scoopable system",
		})
	} else if currentFuel < 1 {
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelWarn,
			Message: fmt.Sprintf("You'll arrive at the next scoopable system with %st fuel left.", strconv.FormatFloat(currentFuel, 'f', 1, 64)),
		})
	} else {
		e.FuelAlert.Publish(&FuelAlert{
			Level:   FuelLevelOk,
			Message: "",
		})
	}
}
