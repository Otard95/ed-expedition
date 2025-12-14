package services

import (
	"ed-expedition/journal"
	"ed-expedition/models"
	"fmt"
	"time"

	"github.com/google/uuid"
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type ExpeditionService struct {
	Index            *models.ExpeditionIndex
	activeExpedition *models.Expedition
	watcher          *journal.Watcher
	fsdJumpChan      chan *journal.FSDJumpEvent
	logger           wailsLogger.Logger
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

	return &ExpeditionService{
		Index:            index,
		activeExpedition: activeExpedition,
		watcher:          watcher,
		logger:           logger,
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
	e.logger.Info(fmt.Sprintf("[ExpeditionService] Handle jump to %s", event.StarSystem))
}

func (e *ExpeditionService) CreateExpedition() (string, error) {
	now := time.Now()
	id := uuid.New().String()

	expedition := &models.Expedition{
		ID:          id,
		Name:        "",
		CreatedAt:   now,
		LastUpdated: now,
		Status:      models.StatusPlanned,
		Routes:      []string{},
		Links:       []models.Link{},
		JumpHistory: []models.JumpHistoryEntry{},
	}

	if err := models.SaveExpedition(expedition); err != nil {
		return "", err
	}

	summary := models.ExpeditionSummary{
		ID:          id,
		Name:        "",
		Status:      models.StatusPlanned,
		CreatedAt:   now,
		LastUpdated: now,
	}

	e.Index.Expeditions = append(e.Index.Expeditions, summary)

	// TODO: Fix orphan expedition
	if err := models.SaveIndex(e.Index); err != nil {
		return "", err
	}

	return id, nil
}

func (e *ExpeditionService) AddRouteToExpedition(expeditionId string, route *models.Route) error {
	if err := models.SaveRoute(route); err != nil {
		return err
	}

	// TODO: Fix orphan route if any of the following fail

	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return err
	}

	isFirstRoute := len(expedition.Routes) == 0

	expedition.Routes = append(expedition.Routes, route.ID)

	if expedition.Start == nil && len(route.Jumps) > 0 {
		expedition.Start = &models.RoutePosition{
			RouteID:   route.ID,
			JumpIndex: 0,
		}
	}

	if isFirstRoute && expedition.Name == "" {
		expedition.Name = route.Name
		expedition.LastUpdated = time.Now()

		for i := range e.Index.Expeditions {
			if e.Index.Expeditions[i].ID == expeditionId {
				e.Index.Expeditions[i].Name = route.Name
				e.Index.Expeditions[i].LastUpdated = expedition.LastUpdated
				break
			}
		}
	}

	if err := models.SaveExpedition(expedition); err != nil {
		return err
	}

	if isFirstRoute && expedition.Name != "" {
		return models.SaveIndex(e.Index)
	}

	return nil
}

func (e *ExpeditionService) DeleteExpedition(expeditionId string) error {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return err
	}

	if expedition.Status == models.StatusActive {
		return fmt.Errorf("cannot delete active expedition")
	}

	if err := models.DeleteExpedition(expeditionId); err != nil {
		return err
	}

	for i, summary := range e.Index.Expeditions {
		if summary.ID == expeditionId {
			e.Index.Expeditions = append(e.Index.Expeditions[:i], e.Index.Expeditions[i+1:]...)
			break
		}
	}

	return models.SaveIndex(e.Index)
}

func (e *ExpeditionService) RenameExpedition(expeditionId, name string) error {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return err
	}

	expedition.Name = name
	expedition.LastUpdated = time.Now()

	if err := models.SaveExpedition(expedition); err != nil {
		return err
	}

	for i := range e.Index.Expeditions {
		if e.Index.Expeditions[i].ID == expeditionId {
			e.Index.Expeditions[i].Name = name
			e.Index.Expeditions[i].LastUpdated = expedition.LastUpdated
			break
		}
	}

	return models.SaveIndex(e.Index)
}
