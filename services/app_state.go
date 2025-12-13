package services

import (
	"ed-expedition/journal"
	"ed-expedition/lib/slice"
	"ed-expedition/models"
	"strings"
)

type AppStateService struct {
	State       *models.AppState
	watcher     *journal.Watcher
	loadoutChan chan *journal.LoadoutEvent
}

func NewAppStateService(watcher *journal.Watcher) *AppStateService {
	state, err := models.LoadAppSate()
	if err != nil {
		panic(err)
	}

	return &AppStateService{
		State:   state,
		watcher: watcher,
	}
}

func (s *AppStateService) Start() {
	s.loadoutChan = s.watcher.Loadout.Subscribe()

	go func() {
		for event := range s.loadoutChan {
			if !s.State.LastKnownLoadout.Timestamp.Before(event.Timestamp) {
				continue
			}

			s.State.LastKnownLoadout = transformLoadoutEventToStateLoadout(event)
			if err := models.SaveAppState(s.State); err != nil {
				// TODO: Proper error handling (log, retry, etc.)
				panic(err)
			}
		}
	}()
}

func (s *AppStateService) Stop() error {
	if s.loadoutChan != nil {
		s.watcher.Loadout.Unsubscribe(s.loadoutChan)
		s.loadoutChan = nil
	}
	return nil
}

func transformLoadoutEventToStateLoadout(event *journal.LoadoutEvent) *models.Loadout {
	fsd := slice.Find(event.Modules, func(module journal.LoadoutModule) bool {
		return module.Slot == "FrameShiftDrive"
	})
	if fsd == nil {
		// TODO: Needs proper error handling
		panic("Journal Loadout event missing fsd module")
	}

	fsdBooster := slice.Find(event.Modules, func(module journal.LoadoutModule) bool {
		return strings.HasPrefix(module.Item, "int_guardianfsdbooster")
	})

	loadout := &models.Loadout{
		Timestamp:   event.Timestamp,
		UnladenMass: event.UnladenMass,
		FuelCapacity: models.FuelCapacity{
			Main:    event.FuelCapacity.Main,
			Reserve: event.FuelCapacity.Reserve,
		},
		FSD: models.LoadoutFSD{
			Item: fsd.Item,
		},
	}

	if fsd.Engineering != nil && len(fsd.Engineering.Modifiers) > 0 {
		if optMass := slice.Find(
			fsd.Engineering.Modifiers,
			func(mod journal.EngineeringModifier) bool {
				return mod.Label == "FSDOptimalMass"
			},
		); optMass != nil {
			loadout.FSD.OptimalMass = &optMass.Value
		}

		if maxFuel := slice.Find(
			fsd.Engineering.Modifiers,
			func(mod journal.EngineeringModifier) bool {
				return mod.Label == "MaxFuelPerJump"
			},
		); maxFuel != nil {
			loadout.FSD.MaxFuelPerJump = &maxFuel.Value
		}
	}

	if fsdBooster != nil {
		loadout.FSDBooster = &fsdBooster.Item
	}

	return loadout
}
