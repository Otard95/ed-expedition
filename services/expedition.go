package services

import (
	"ed-expedition/journal"
	"ed-expedition/models"
)

type ExpeditionService struct {
	index            *models.ExpeditionIndex
	activeExpedition *models.Expedition
	watcher          *journal.Watcher
	fsdJumpChan      chan *journal.FSDJumpEvent
}

func NewExpeditionService(watcher *journal.Watcher) *ExpeditionService {
	index, err := models.LoadIndex()
	if err != nil {
		panic(err)
	}

	activeExpedition, err := index.LoadActiveExpedition()
	if err != nil {
		panic(err)
	}

	return &ExpeditionService{
		index:            index,
		activeExpedition: activeExpedition,
		watcher:          watcher,
	}
}

func (e *ExpeditionService) Start() {
	e.fsdJumpChan = e.watcher.FSDJump.Subscribe()

	go func() {
		for event := range e.fsdJumpChan {
			if e.activeExpedition == nil {
				continue
			}

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

}
