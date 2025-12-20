package services

import (
	"ed-expedition/journal"
	"ed-expedition/lib/channels"
	"ed-expedition/lib/slice"
	"ed-expedition/models"
	"fmt"
	"time"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type ExpeditionService struct {
	Index            *models.ExpeditionIndex
	activeExpedition *models.Expedition
	bakedRoute       *models.Route
	watcher          *journal.Watcher
	fsdJumpChan      chan *journal.FSDJumpEvent
	logger           wailsLogger.Logger

	JumpHistory *channels.FanoutChannel[*models.JumpHistoryEntry]
}

func NewExpeditionService(watcher *journal.Watcher, logger wailsLogger.Logger) *ExpeditionService {
	index, err := models.LoadIndex()
	if err != nil {
		panic(err)
	}

	activeExpedition, err := index.LoadActiveExpedition()
	if err != nil {
		panic(err)
	}

	var bakedRoute *models.Route
	if activeExpedition != nil {
		bakedRoute, err = activeExpedition.LoadBaked()
		if err != nil {
			panic(err)
		}
	}

	return &ExpeditionService{
		Index:            index,
		activeExpedition: activeExpedition,
		bakedRoute:       bakedRoute,
		watcher:          watcher,
		logger:           logger,
		JumpHistory: channels.NewFanoutChannel[*models.JumpHistoryEntry](
			"JumpHistory", 0, 5*time.Millisecond, logger,
		),
	}
}

func (e *ExpeditionService) Start() {
	e.fsdJumpChan = e.watcher.FSDJump.Subscribe()

	go func() {
		for event := range e.fsdJumpChan {
			e.handleJump(event)
		}
	}()
}

func (e *ExpeditionService) Stop() error {
	if e.fsdJumpChan != nil {
		e.watcher.FSDJump.Unsubscribe(e.fsdJumpChan)
		e.fsdJumpChan = nil
	}
	return nil
}

func (e *ExpeditionService) handleJump(event *journal.FSDJumpEvent) {
	if e.activeExpedition == nil {
		return
	}
	jumpHistory := e.activeExpedition.JumpHistory
	if len(jumpHistory) > 0 && !jumpHistory[len(jumpHistory)-1].Timestamp.Before(event.Timestamp) {
		return
	}
	e.logger.Info(fmt.Sprintf("[ExpeditionService] Handle jump to %s", event.StarSystem))

	if e.activeExpedition.CurrentBakedIndex >= len(e.bakedRoute.Jumps)-1 {
		e.logger.Warning("Received jump but no more expected jumps in route. This should only happen if your have only one jump in your expedition.")
		return
	}
	expectedSystem := e.bakedRoute.Jumps[e.activeExpedition.CurrentBakedIndex+1]
	isExpected := event.SystemAddress == expectedSystem.SystemID
	if e.activeExpedition.CurrentBakedIndex == 0 && e.bakedRoute.Jumps[e.activeExpedition.CurrentBakedIndex].SystemID == event.SystemAddress {
		e.activeExpedition.CurrentBakedIndex--
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
		for i := e.activeExpedition.CurrentBakedIndex; i < len(e.bakedRoute.Jumps); i++ {
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
			err := e.CompleteActiveExpedition()
			if err != nil {
				panic("Failed to complete expedition")
			}
			return
		}
	}

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
			e.logger.Error(fmt.Sprintf("[ExpeditionService] handleJump - Failed to save index: %v", err))
		}
	}

	e.JumpHistory.Publish(&historicalJump)
}
