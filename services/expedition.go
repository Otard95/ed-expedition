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

// Your are expected to add the last historical jump before calling this
// function, however you do not need to save the expedition
func (e *ExpeditionService) CompleteActiveExpedition() error {
	if e.activeExpedition == nil {
		return errors.New("There is no active expedition to complete")
	}

	e.activeExpedition.EndedOn = time.Now()
	e.activeExpedition.LastUpdated = time.Now()
	e.activeExpedition.Status = models.StatusCompleted

	err := models.SaveExpedition(e.activeExpedition)
	if err != nil {
		return nil
	}

	expeditionSummary := slice.Find(
		e.Index.Expeditions,
		func(exp models.ExpeditionSummary) bool { return exp.ID == e.activeExpedition.ID },
	)
	if expeditionSummary == nil {
		return errors.New("Unable to find active expedition summary")
	}
	expeditionSummary.Status = models.StatusCompleted
	expeditionSummary.LastUpdated = time.Now()
	e.Index.ActiveExpeditionID = nil

	err = models.SaveIndex(e.Index)
	if err != nil {
		return err
	}

	e.activeExpedition = nil
	e.bakedRoute = nil

	return nil
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

func (e *ExpeditionService) EndActiveExpedition() error {
	if e.activeExpedition == nil {
		// TODO: This should probably be an error
		return nil
	}

	e.activeExpedition.EndedOn = time.Now()
	e.activeExpedition.LastUpdated = time.Now()
	e.activeExpedition.Status = models.StatusEnded

	err := models.SaveExpedition(e.activeExpedition)
	if err != nil {
		return nil
	}

	expeditionSummary := slice.Find(
		e.Index.Expeditions,
		func(exp models.ExpeditionSummary) bool { return exp.ID == *e.Index.ActiveExpeditionID },
	)
	if expeditionSummary == nil {
		return errors.New("Unable to find active expedition summary")
	}
	expeditionSummary.Status = models.StatusEnded
	expeditionSummary.LastUpdated = time.Now()
	e.Index.ActiveExpeditionID = nil

	err = models.SaveIndex(e.Index)
	if err != nil {
		return err
	}

	e.activeExpedition = nil
	e.bakedRoute = nil

	return nil
}

func (e *ExpeditionService) StartExpedition(expeditionId string) error {
	expeditionSummary := slice.Find(
		e.Index.Expeditions,
		func(exp models.ExpeditionSummary) bool { return exp.ID == expeditionId },
	)
	if expeditionSummary == nil {
		return errors.New("Failed to find this expedition in the index")
	}

	expedition, err := expeditionSummary.LoadFull()
	if err != nil {
		return err
	}

	err = ensureExpeditionCanBeStarted(expedition)
	if err != nil {
		return err
	}

	route, loopBackIndex, err := bakeExpeditionRoute(expedition)
	if err != nil {
		return err
	}

	// TODO: Fix orphan baked route if any of the following fail
	err = models.SaveRoute(route)
	if err != nil {
		return err
	}

	err = e.EndActiveExpedition()
	if err != nil {
		return err
	}

	// TODO: Fix orphan baked route if any of the following fail
	expedition.BakedRouteID = &route.ID
	expedition.CurrentBakedIndex = 0
	if loopBackIndex > -1 {
		expedition.BakedLoopBackIndex = &loopBackIndex
	}
	expedition.StartedOn = time.Now()
	expedition.LastUpdated = time.Now()
	expedition.Status = models.StatusActive

	err = models.SaveExpedition(expedition)
	if err != nil {
		return err
	}

	// TODO: Fix inconsistent state if saving index fails

	e.activeExpedition = expedition
	e.Index.ActiveExpeditionID = &expedition.ID
	expeditionSummary.Status = models.StatusActive
	expeditionSummary.LastUpdated = expedition.LastUpdated

	err = models.SaveIndex(e.Index)
	if err != nil {
		return err
	}

	return nil
}

func bakeExpeditionRoute(expedition *models.Expedition) (*models.Route, int, error) {
	routes, err := expedition.LoadRoutes()
	if err != nil {
		return nil, -1, err
	}

	routeById := make(map[string]*models.Route, len(expedition.Routes))

	for _, route := range routes {
		routeById[route.ID] = route
	}

	newRouteJumps := make([]models.RouteJump, 0, 64)
	loopBackIndex := -1

	next := expedition.Start.Clone()
	visited := make([]*models.RouteJump, 0, 64)
	for next != nil {
		currentRoute, ok := routeById[next.RouteID]
		if !ok {
			// We panic here because if we end up in this state then we have at some
			// point created an invalid link, and/or the route is invalid/corrupt.
			panic("Failed to find route in 'routeById[next.RouteID]' while baking expedition route")
		}
		if len(currentRoute.Jumps) <= next.JumpIndex {
			// We panic here because if we end up in this state then we have at some
			// point created an invalid link, and/or the route is invalid/corrupt.
			panic("The next.JumpIndex is out of bounds")
		}
		currentJump := &currentRoute.Jumps[next.JumpIndex]
		if i := slices.Index(visited, currentJump); i > -1 {
			loopBackIndex = i
			break
		}
		visited = append(visited, currentJump)

		// Because links connect two identical systems, we should expect two
		// identical systems in a row, and skip them if we do encounter them.
		if len(visited) > 1 && visited[len(visited)-2].SystemID == currentJump.SystemID {
			continue
		}

		newRouteJumps = append(newRouteJumps, *currentJump.Clone())

		link := slice.Find(
			expedition.Links,
			func(l models.Link) bool { return l.From.Equal(next) },
		)
		if link != nil {
			next = link.To.Clone()
		} else if len(currentRoute.Jumps) > next.JumpIndex+1 {
			next.JumpIndex++
		} else {
			next = nil
		}
	}

	return &models.Route{
		ID:      uuid.NewString(),
		Name:    fmt.Sprintf("Baked route for expedition: %s", expedition.Name),
		Plotter: "ed-expedition-baker",
		PlotterParams: map[string]any{
			"expedition_id": expedition.ID,
		},
		PlotterMetadata: nil,
		Jumps:           newRouteJumps,
		CreatedAt:       time.Now(),
	}, loopBackIndex, nil
}

func ensureExpeditionCanBeStarted(expedition *models.Expedition) error {
	if !expedition.IsEditable() {
		return errors.New("The expedition is not in the planned state, it cannot be started.")
	}
	if len(expedition.Routes) == 0 || expedition.Start == nil {
		return errors.New("The expedition is needs at least one route and a start.")
	}
	return nil
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
