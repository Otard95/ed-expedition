package services

import (
	"ed-expedition/journal"
	"ed-expedition/lib/slice"
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
	bakedRoute       *models.Route
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
	if len(jumpHistory) > 0 && jumpHistory[len(jumpHistory)-1].Timestamp.After(event.Timestamp) {
		return
	}
	e.logger.Info(fmt.Sprintf("[ExpeditionService] Handle jump to %s", event.StarSystem))

	if e.activeExpedition.CurrentBakedIndex >= len(e.bakedRoute.Jumps)-1 {
		e.logger.Warning("Received jump but no more expected jumps in route. This should only happen if your have only one jump in your expedition.")
		return
	}
	expectedSystem := e.bakedRoute.Jumps[e.activeExpedition.CurrentBakedIndex+1]
	isExpected := event.SystemAddress == expectedSystem.SystemID

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

func (e *ExpeditionService) DeleteLink(expeditionId, linkId string) error {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return err
	}

	if !expedition.IsEditable() {
		return fmt.Errorf("cannot delete link: only planned expeditions can be edited")
	}

	linkIndex := slices.IndexFunc(expedition.Links, func(l models.Link) bool { return l.ID == linkId })
	if linkIndex == -1 {
		return fmt.Errorf("link not found in expedition")
	}

	expedition.Links = slices.Delete(expedition.Links, linkIndex, linkIndex+1)
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
