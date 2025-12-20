package services

import (
	"ed-expedition/lib/slice"
	"ed-expedition/models"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

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

	e.CompleteExpedition.Publish(e.activeExpedition)

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
