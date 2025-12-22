package services

import (
	"ed-expedition/database"
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
func (e *ExpeditionService) completeActiveExpedition() error {
	if e.activeExpedition == nil {
		return errors.New("There is no active expedition to complete")
	}

	expedition := e.activeExpedition
	expeditionSummary := slice.Find(
		e.Index.Expeditions,
		func(exp models.ExpeditionSummary) bool { return exp.ID == expedition.ID },
	)
	if expeditionSummary == nil {
		return errors.New("Unable to find active expedition summary")
	}

	prevLastUpdated := expedition.LastUpdated
	undo := func() {
		expedition.EndedOn = time.Time{}
		expedition.LastUpdated = prevLastUpdated
		expedition.Status = models.StatusActive

		expeditionSummary.Status = models.StatusActive
		expeditionSummary.LastUpdated = prevLastUpdated
		e.Index.ActiveExpeditionID = &expedition.ID
	}

	expedition.EndedOn = time.Now()
	expedition.LastUpdated = time.Now()
	expedition.Status = models.StatusCompleted

	expeditionSummary.Status = models.StatusCompleted
	expeditionSummary.LastUpdated = time.Now()
	e.Index.ActiveExpeditionID = nil

	t := database.NewTransaction("ExpeditionService.completeActiveExpedition")

	err := models.TSaveExpedition(t, e.activeExpedition)
	if err != nil {
		undo()
		return fmt.Errorf("Failed to save expedition: %s", err.Error())
	}

	err = models.TSaveIndex(t, e.Index)
	if err != nil {
		undo()
		if rErr := t.Rewind(); rErr != nil {
			e.logger.Error("[ExpeditionService] completeActiveExpedition transaction rewind failed.")
		}
		return fmt.Errorf("Failed to save index: %s", err.Error())
	}

	if err := t.Apply(); err != nil {
		undo()
		e.logger.Error("[ExpeditionService] completeActiveExpedition transaction failed to apply.")
		return fmt.Errorf("Failed to complete expedition: %s", err.Error())
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

	summary := models.ExpeditionSummary{
		ID:          id,
		Name:        "",
		Status:      models.StatusPlanned,
		CreatedAt:   now,
		LastUpdated: now,
	}

	t := database.NewTransaction("ExpeditionService.CreateExpedition")

	if err := models.TSaveExpedition(t, expedition); err != nil {
		return "", err
	}

	e.Index.Expeditions = append(e.Index.Expeditions, summary)

	if err := models.TSaveIndex(t, e.Index); err != nil {
		e.Index.Expeditions = e.Index.Expeditions[:len(e.Index.Expeditions)-1]

		rErr := t.Rewind()
		if rErr != nil {
			e.logger.Error("[ExpeditionService] CreateExpedition transaction rewind failed.")
		}
		return "", err
	}

	if err := t.Apply(); err != nil {
		e.Index.Expeditions = e.Index.Expeditions[:len(e.Index.Expeditions)-1]

		e.logger.Error(fmt.Sprintf("[ExpeditionService] CreateExpedition transaction failed to apply: %v", err))
		return "", err
	}

	return id, nil
}

func (e *ExpeditionService) DeleteExpedition(expeditionId string) error {
	summaryIndex := slices.IndexFunc(
		e.Index.Expeditions,
		func(s models.ExpeditionSummary) bool { return s.ID == expeditionId },
	)

	if summaryIndex < 0 {
		return fmt.Errorf("Unable to find expedition in index")
	}

	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return fmt.Errorf("Unable to load expedition: %s", err.Error())
	}

	if !expedition.IsEditable() {
		return fmt.Errorf("cannot delete expedition: only planned expeditions can be deleted")
	}

	prevExpeditions := make([]models.ExpeditionSummary, len(e.Index.Expeditions))
	copy(prevExpeditions, e.Index.Expeditions)
	e.Index.Expeditions = slices.Delete(e.Index.Expeditions, summaryIndex, summaryIndex+1)

	if err := models.SaveIndex(e.Index); err != nil {
		e.Index.Expeditions = prevExpeditions
		return fmt.Errorf("Failed to save index: %s", err.Error())
	}

	if err := models.DeleteExpedition(expeditionId); err != nil {
		// This is not great. But the only side-effect is that we have an undeleted
		// unreachable expedition file.
		// TODO: Add cleanup on app start
		e.logger.Error(fmt.Sprintf("Failed to delete expedition: %s", err.Error()))
	}

	return nil
}

func (e *ExpeditionService) EndActiveExpedition(t *database.Transaction) error {
	if e.activeExpedition == nil {
		// TODO: This should maybe be an error?
		return nil
	}

	expeditionSummary := slice.Find(
		e.Index.Expeditions,
		func(exp models.ExpeditionSummary) bool { return exp.ID == *e.Index.ActiveExpeditionID },
	)
	if expeditionSummary == nil {
		return errors.New("Unable to find active expedition summary")
	}

	prevActiveExpeditionLastUpdated := e.activeExpedition.LastUpdated
	prevActiveExpeditionStatus := e.activeExpedition.Status
	prevActiveExpeditionId := *e.Index.ActiveExpeditionID
	undo := func() {
		e.activeExpedition.EndedOn = time.Time{}
		e.activeExpedition.LastUpdated = prevActiveExpeditionLastUpdated
		e.activeExpedition.Status = prevActiveExpeditionStatus

		expeditionSummary.Status = prevActiveExpeditionStatus
		expeditionSummary.LastUpdated = prevActiveExpeditionLastUpdated
		e.Index.ActiveExpeditionID = &prevActiveExpeditionId
	}

	e.activeExpedition.EndedOn = time.Now()
	e.activeExpedition.LastUpdated = time.Now()
	e.activeExpedition.Status = models.StatusEnded

	expeditionSummary.Status = models.StatusEnded
	expeditionSummary.LastUpdated = time.Now()
	e.Index.ActiveExpeditionID = nil

	// If we inherit the transaction we should not Rewind/Apply, that would be the
	// transaction's owner's job. Otherwise we'll need to handle the Rewind/Apply
	tr := t
	if tr == nil {
		tr = database.NewTransaction("ExpeditionService.EndActiveExpedition")
	}

	err := models.TSaveExpedition(tr, e.activeExpedition)
	if err != nil {
		undo()
		return fmt.Errorf("Failed to save expedition: %s", err.Error())
	}

	err = models.TSaveIndex(tr, e.Index)
	if err != nil {
		undo()
		if t == nil {
			if rErr := tr.Rewind(); rErr != nil {
				e.logger.Error("[ExpeditionService] EndActiveExpedition transaction rewind failed.")
			}
		}
		return fmt.Errorf("Failed to save index: %s", err.Error())
	}

	if t == nil {
		if err := tr.Apply(); err != nil {
			undo()
			e.logger.Error("[ExpeditionService] EndActiveExpedition transaction failed to apply.")
			return fmt.Errorf("Failed to end active expedition: %s", err.Error())
		}
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
	expedition.BakedRouteID = &route.ID
	expedition.CurrentBakedIndex = 0
	if loopBackIndex > -1 {
		expedition.BakedLoopBackIndex = &loopBackIndex
	}
	expedition.StartedOn = time.Now()
	expedition.LastUpdated = time.Now()
	expedition.Status = models.StatusActive

	prevActiveExpeditionId := e.Index.ActiveExpeditionID
	prevActiveExpedition := e.activeExpedition
	prevLastUpdated := expeditionSummary.LastUpdated
	undo := func() {
		e.activeExpedition = prevActiveExpedition
		e.Index.ActiveExpeditionID = prevActiveExpeditionId
		if prevActiveExpedition != nil {
			expeditionSummary.Status = prevActiveExpedition.Status
			expeditionSummary.LastUpdated = prevActiveExpedition.LastUpdated
		} else {
			e.Index.ActiveExpeditionID = nil
			expeditionSummary.Status = models.StatusPlanned
			expeditionSummary.LastUpdated = prevLastUpdated
		}
	}

	e.activeExpedition = expedition
	e.Index.ActiveExpeditionID = &expedition.ID
	expeditionSummary.Status = models.StatusActive
	expeditionSummary.LastUpdated = expedition.LastUpdated

	t := database.NewTransaction("ExpeditionService.StartExpedition")

	err = models.TSaveRoute(t, route)
	if err != nil {
		undo()
		return fmt.Errorf("Failed to save route: %s", err.Error())
	}

	err = e.EndActiveExpedition(t)
	if err != nil {
		undo()
		if rErr := t.Rewind(); rErr != nil {
			e.logger.Error("[ExpeditionService] StartExpedition transaction rewind failed after 'EndActiveExpedition'.")
		}
		return fmt.Errorf("Could not end active expedition first: %s", err.Error())
	}

	err = models.TSaveExpedition(t, expedition)
	if err != nil {
		undo()
		if rErr := t.Rewind(); rErr != nil {
			e.logger.Error("[ExpeditionService] StartExpedition transaction rewind failed after save expedition.")
		}
		return fmt.Errorf("Could not save expedition: %s", err.Error())
	}

	err = models.TSaveIndex(t, e.Index)
	if err != nil {
		undo()
		if rErr := t.Rewind(); rErr != nil {
			e.logger.Error("[ExpeditionService] StartExpedition transaction rewind failed after save index.")
		}
		return fmt.Errorf("Could not save index: %s", err.Error())
	}

	if err := t.Apply(); err != nil {
		undo()
		e.logger.Error("[ExpeditionService] StartExpedition transaction failed to apply.")
		return fmt.Errorf("Failed to start expedition: %s", err.Error())
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
