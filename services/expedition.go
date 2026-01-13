package services

import (
	"ed-expedition/journal"
	"ed-expedition/lib/channels"
	"ed-expedition/models"
	"sync"
	"time"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type ExpeditionService struct {
	Index              *models.ExpeditionIndex
	activeExpedition   *models.Expedition
	bakedRoute         *models.Route
	currentJump        *models.JumpHistoryEntry
	previouslyScooping bool

	watcher         *journal.Watcher
	fsdJumpChan     chan *journal.FSDJumpEvent
	startJumpChan   chan *journal.StartJumpEvent
	fsdChargingChan chan bool
	scoopingChan    chan bool
	fuelChan        chan *journal.FuelStatus
	logger          wailsLogger.Logger

	jumpState     jumpState
	jumpStateMu   sync.Mutex
	chargingTimer *time.Timer

	JumpHistory        *channels.FanoutChannel[*models.JumpHistoryEntry]
	CompleteExpedition *channels.FanoutChannel[*models.Expedition]
	CurrentJump        *channels.FanoutChannel[*models.JumpHistoryEntry]
	FuelAlert          *channels.FanoutChannel[*FuelAlert]
}

func NewExpeditionService(watcher *journal.Watcher, logger wailsLogger.Logger, currentSystem *int64) *ExpeditionService {
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

	var currentJump *models.JumpHistoryEntry
	if activeExpedition != nil &&
		currentSystem != nil &&
		len(activeExpedition.JumpHistory) > 0 &&
		activeExpedition.JumpHistory[len(activeExpedition.JumpHistory)-1].SystemID == *currentSystem {
		currentJump = &activeExpedition.JumpHistory[len(activeExpedition.JumpHistory)-1]
	}

	return &ExpeditionService{
		Index:              index,
		activeExpedition:   activeExpedition,
		bakedRoute:         bakedRoute,
		currentJump:        currentJump,
		previouslyScooping: false,

		watcher: watcher,
		logger:  logger,

		JumpHistory: channels.NewFanoutChannel[*models.JumpHistoryEntry](
			"JumpHistory", 0, 5*time.Millisecond, logger,
		),
		CompleteExpedition: channels.NewFanoutChannel[*models.Expedition](
			"CompleteExpedition", 0, 5*time.Millisecond, logger,
		),
		CurrentJump: channels.NewFanoutChannel[*models.JumpHistoryEntry](
			"CurrentJump", 0, 5*time.Millisecond, logger,
		),
		FuelAlert: channels.NewFanoutChannel[*FuelAlert](
			"FuelAlert", 0, 5*time.Millisecond, logger,
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

	e.startJumpChan = e.watcher.StartJump.Subscribe()
	go func() {
		for event := range e.startJumpChan {
			e.handleStartJump(event)
		}
	}()

	e.fsdChargingChan = e.watcher.FsdCharging.Subscribe()
	go func() {
		for event := range e.fsdChargingChan {
			e.handleFsdCharging(event)
		}
	}()

	e.scoopingChan = e.watcher.Scooping.Subscribe()
	go func() {
		for event := range e.scoopingChan {
			e.handleRefueling(event)
		}
	}()

	e.fuelChan = e.watcher.Fuel.Subscribe()
	go func() {
		for event := range e.fuelChan {
			e.handleFuelChange(event)
		}
	}()
}

func (e *ExpeditionService) Stop() error {
	if e.fsdJumpChan != nil {
		e.watcher.FSDJump.Unsubscribe(e.fsdJumpChan)
		e.fsdJumpChan = nil
	}
	if e.startJumpChan != nil {
		e.watcher.StartJump.Unsubscribe(e.startJumpChan)
		e.startJumpChan = nil
	}
	if e.fsdChargingChan != nil {
		e.watcher.FsdCharging.Unsubscribe(e.fsdChargingChan)
		e.fsdChargingChan = nil
	}
	e.stopChargingTimeout()
	return nil
}

func (e *ExpeditionService) GetNextSystemName() *string {
	if e.activeExpedition == nil || e.bakedRoute == nil {
		return nil
	}
	nextIndex := e.activeExpedition.CurrentBakedIndex + 1
	if nextIndex >= len(e.bakedRoute.Jumps) {
		return nil
	}
	return &e.bakedRoute.Jumps[nextIndex].SystemName
}
