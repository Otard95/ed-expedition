package services

import (
	"ed-expedition/journal"
	"ed-expedition/lib/slice"
	"ed-expedition/models"
	"fmt"
	"time"
)

type jumpState uint8

const (
	jumpStateNormal jumpState = iota
	jumpStateCharging
	jumpStateCommitted
)

const jumpChargingTimeout = 5 * time.Second

func (e *ExpeditionService) isJumpInProgress() bool {
	e.jumpStateMu.Lock()
	defer e.jumpStateMu.Unlock()
	return e.jumpState != jumpStateNormal
}

func (e *ExpeditionService) handleStartJump(event *journal.StartJumpEvent) {
	e.jumpStateMu.Lock()
	defer e.jumpStateMu.Unlock()

	e.logger.Trace(fmt.Sprintf("[ExpeditionService](Jump) handleStartJump: type=%s, state=%d", event.JumpType, e.jumpState))

	if event.JumpType != journal.JumpTypeHyperspace {
		if e.jumpState == jumpStateCharging {
			e.logger.Trace("[ExpeditionService](Jump) handleStartJump: supercruise jump, resetting to normal")
			e.setJumpStateLocked(jumpStateNormal)
		}
		return
	}

	e.logger.Trace("[ExpeditionService](Jump) handleStartJump: hyperspace jump, transitioning to committed")
	e.setJumpStateLocked(jumpStateCommitted)
}

func (e *ExpeditionService) handleFsdCharging(charging bool) {
	e.jumpStateMu.Lock()
	defer e.jumpStateMu.Unlock()

	e.logger.Trace(fmt.Sprintf("[ExpeditionService](Jump) handleFsdCharging: charging=%v, state=%d", charging, e.jumpState))

	if charging && e.jumpState == jumpStateNormal {
		e.logger.Trace("[ExpeditionService](Jump) handleFsdCharging: entering charging state")
		e.setJumpStateLocked(jumpStateCharging)
		return
	}

	if !charging && e.jumpState == jumpStateCharging {
		e.logger.Trace("[ExpeditionService](Jump) handleFsdCharging: charging stopped, starting timeout to wait for StartJump")
		e.startChargingTimeoutLocked()
	}
}

func (e *ExpeditionService) setJumpState(state jumpState) {
	e.jumpStateMu.Lock()
	defer e.jumpStateMu.Unlock()
	e.setJumpStateLocked(state)
}

func (e *ExpeditionService) setJumpStateLocked(state jumpState) {
	e.logger.Trace(fmt.Sprintf("[ExpeditionService](Jump) setJumpState: %d -> %d", e.jumpState, state))
	e.jumpState = state

	if state == jumpStateNormal {
		e.stopChargingTimeoutLocked()
	}
}

func (e *ExpeditionService) startChargingTimeoutLocked() {
	e.stopChargingTimeoutLocked()

	e.chargingTimer = time.AfterFunc(jumpChargingTimeout, func() {
		e.jumpStateMu.Lock()
		defer e.jumpStateMu.Unlock()

		e.logger.Trace("[ExpeditionService](Jump) charging timeout expired, resetting to normal")
		if e.jumpState == jumpStateCharging {
			e.setJumpStateLocked(jumpStateNormal)
		}
	})
}

func (e *ExpeditionService) stopChargingTimeout() {
	e.jumpStateMu.Lock()
	defer e.jumpStateMu.Unlock()
	e.stopChargingTimeoutLocked()
}

func (e *ExpeditionService) stopChargingTimeoutLocked() {
	if e.chargingTimer != nil {
		e.chargingTimer.Stop()
		e.chargingTimer = nil
	}
}

func (e *ExpeditionService) handleJump(event *journal.FSDJumpEvent) {
	e.logger.Trace(fmt.Sprintf("[ExpeditionService](Jump) handleJump: system=%s, state=%d", event.StarSystem, e.jumpState))
	e.setJumpState(jumpStateNormal)

	if e.activeExpedition == nil {
		return
	}
	jumpHistory := e.activeExpedition.JumpHistory
	if len(jumpHistory) > 0 && !jumpHistory[len(jumpHistory)-1].Timestamp.Before(event.Timestamp) {
		return
	}
	e.logger.Info(fmt.Sprintf("[ExpeditionService](Jump) Handle jump to %s", event.StarSystem))

	if e.activeExpedition.CurrentBakedIndex >= len(e.bakedRoute.Jumps)-1 {
		e.logger.Warning("Received jump but no more expected jumps in route. This should only happen if your have only one jump in your expedition.")
		return
	}

	expectedSystem := e.bakedRoute.Jumps[e.activeExpedition.CurrentBakedIndex+1]
	isExpected := event.SystemAddress == expectedSystem.SystemID

	// This is required because at the time of starting the expedition its not
	// necessarily guarantieed that we know the players current position.
	// If that is the case the expedition would have started with index -1 where
	// it maybe should have been 0
	if !isExpected && e.activeExpedition.CurrentBakedIndex == -1 && len(e.bakedRoute.Jumps) > 1 && e.bakedRoute.Jumps[1].SystemID == event.SystemAddress {
		e.activeExpedition.CurrentBakedIndex++
		isExpected = true
	}

	historicalJump := models.JumpHistoryEntry{
		Timestamp:  event.Timestamp,
		SystemName: event.StarSystem,
		SystemID:   event.SystemAddress,

		Distance:  event.JumpDist,
		FuelUsed:  event.FuelUsed,
		FuelLevel: event.FuelLevel,

		Expected:  isExpected,
		Synthetic: false,
	}

	if isExpected {
		e.activeExpedition.CurrentBakedIndex++

		cpy := e.activeExpedition.CurrentBakedIndex
		historicalJump.BakedIndex = &cpy
	} else {
		for i := e.activeExpedition.CurrentBakedIndex + 2; i < len(e.bakedRoute.Jumps); i++ {
			if e.bakedRoute.Jumps[i].SystemID == event.SystemAddress {
				historicalJump.BakedIndex = &i
				e.activeExpedition.CurrentBakedIndex = i
				break
			}
		}
	}

	e.activeExpedition.JumpHistory = append(e.activeExpedition.JumpHistory, historicalJump)
	e.activeExpedition.LastUpdated = time.Now()

	if e.activeExpedition.CurrentBakedIndex >= len(e.bakedRoute.Jumps)-1 {
		if e.activeExpedition.BakedLoopBackIndex != nil {
			e.activeExpedition.CurrentBakedIndex = *e.activeExpedition.BakedLoopBackIndex
		} else {
			if err := e.completeActiveExpedition(); err != nil {
				panic("Failed to complete expedition")
			}
			return
		}
	}

	e.currentJump = &e.activeExpedition.JumpHistory[len(e.activeExpedition.JumpHistory)-1]

	err := models.SaveExpedition(e.activeExpedition)
	if err != nil {
		panic("Failed to save expedition after jump")
	}

	summary := slice.Find(
		e.Index.Expeditions,
		func(s models.ExpeditionSummary) bool { return s.ID == e.activeExpedition.ID },
	)
	if summary != nil {
		summary.LastUpdated = e.activeExpedition.LastUpdated
		err = models.SaveIndex(e.Index)
		if err != nil {
			e.logger.Error(fmt.Sprintf("[ExpeditionService](Jump) handleJump - Failed to save index: %v", err))
		}
	}

	e.JumpHistory.Publish(&historicalJump)
}
