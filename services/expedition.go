package services

import (
	"ed-expedition/journal"
	"ed-expedition/models"
	"errors"
	"fmt"
	"slices"
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

	if !expedition.IsEditable() {
		return fmt.Errorf("cannot delete expedition: only planned expeditions can be deleted")
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

func (e *ExpeditionService) RemoveRouteFromExpedition(expeditionId, routeId string) error {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return err
	}

	if !expedition.IsEditable() {
		return fmt.Errorf("cannot delete route: only planned expeditions can be edited")
	}

	routeIndex := slices.IndexFunc(expedition.Routes, func(id string) bool { return id == routeId })
	if routeIndex == -1 {
		return fmt.Errorf("route not found in expedition")
	}

	expedition.Routes = slices.Delete(expedition.Routes, routeIndex, routeIndex+1)

	expedition.Links = slices.DeleteFunc(
		expedition.Links,
		func(l models.Link) bool {
			return l.From.RouteID == routeId || l.To.RouteID == routeId
		},
	)

	if expedition.Start != nil && expedition.Start.RouteID == routeId {
		if len(expedition.Routes) > 0 {
			expedition.Start = &models.RoutePosition{
				RouteID:   expedition.Routes[0],
				JumpIndex: 0,
			}
		} else {
			expedition.Start = nil
		}
	}

	expedition.LastUpdated = time.Now()

	return models.SaveExpedition(expedition)
}

func (e *ExpeditionService) CreateLink(expeditionId string, from, to models.RoutePosition) error {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return err
	}

	if !expedition.IsEditable() {
		return fmt.Errorf("cannot create link: only planned expeditions can be edited")
	}

	link := models.Link{
		ID:   uuid.New().String(),
		From: from,
		To:   to,
	}

	err = validateLink(expedition, link)
	if err != nil {
		return err
	}

	expedition.Links = append(expedition.Links, link)
	expedition.LastUpdated = time.Now()

	return models.SaveExpedition(expedition)
}

func validateLink(expedition *models.Expedition, link models.Link) error {
	if link.From.JumpIndex < 0 {
		return errors.New("The 'from' jump index cannot be negative")
	}
	if link.To.JumpIndex < 0 {
		return errors.New("The 'to' jump index cannot be negative")
	}

	if !expedition.HasRoute(link.From.RouteID) {
		return errors.New("The 'from' route does not exist on this expedition")
	}
	if !expedition.HasRoute(link.To.RouteID) {
		return errors.New("The 'to' route does not exist on this expedition")
	}

	if slices.ContainsFunc(expedition.Links, func(l models.Link) bool {
		return link.From.Equal(&l.From)
	}) {
		return errors.New("There's already an outgoing link from the 'from' system")
	}

	fromRoute, err := models.LoadRoute(link.From.RouteID)
	if err != nil {
		return fmt.Errorf("Failed to load 'from' route: %w", err)
	}
	toRoute, err := models.LoadRoute(link.To.RouteID)
	if err != nil {
		return fmt.Errorf("Failed to load 'to' route: %w", err)
	}

	if link.From.JumpIndex >= len(fromRoute.Jumps) {
		return errors.New("The 'from' route index is out of bounds")
	}
	if link.To.JumpIndex >= len(toRoute.Jumps) {
		return errors.New("The 'to' route index is out of bounds")
	}

	fromSystem := fromRoute.Jumps[link.From.JumpIndex]
	toSystem := toRoute.Jumps[link.To.JumpIndex]

	if fromSystem.SystemID != toSystem.SystemID {
		return errors.New("The 'from' and 'to' systems are not the same")
	}

	return nil
}
