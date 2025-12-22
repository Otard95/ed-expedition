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

func (e *ExpeditionService) AddRouteToExpedition(expeditionId string, route *models.Route) error {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return fmt.Errorf("Failed to load expedition with id '%s': %s", expeditionId, err.Error())
	}

	if !expedition.IsEditable() {
		return errors.New("Expedition is not editable")
	}

	indexExpIndex := slices.IndexFunc(
		e.Index.Expeditions,
		func(s models.ExpeditionSummary) bool { return s.ID == expeditionId },
	)
	isFirstRoute := len(expedition.Routes) == 0

	name := expedition.Name
	expeditionLastUpdate := expedition.LastUpdated
	undo := func() {
		if indexExpIndex > -1 {
			e.Index.Expeditions[indexExpIndex].Name = name
			e.Index.Expeditions[indexExpIndex].LastUpdated = expeditionLastUpdate
		}
	}

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

		if indexExpIndex > -1 {
			e.Index.Expeditions[indexExpIndex].Name = route.Name
			e.Index.Expeditions[indexExpIndex].LastUpdated = expedition.LastUpdated
		}
	}

	t := database.NewTransaction("ExpeditionService.AddRouteToExpedition")

	if err := models.TSaveRoute(t, route); err != nil {
		undo()
		return fmt.Errorf("Failed to save route: %s", err.Error())
	}

	if err := models.TSaveExpedition(t, expedition); err != nil {
		undo()
		if err := t.Rewind(); err != nil {
			e.logger.Error("[ExpeditionService] AddRouteToExpedition transaction rewind failed after save expedition.")
		}
		return fmt.Errorf("Failed to save expedition: %s", err.Error())
	}

	if isFirstRoute && expedition.Name != "" {
		if err := models.TSaveIndex(t, e.Index); err != nil {
			undo()
			if err := t.Rewind(); err != nil {
				e.logger.Error("[ExpeditionService] AddRouteToExpedition transaction rewind failed after save index.")
			}
			return fmt.Errorf("Failed to save index: %s", err.Error())
		}
	}

	if err := t.Apply(); err != nil {
		undo()
		e.logger.Error("[ExpeditionService] AddRouteToExpedition transaction failed to apply.")
		return fmt.Errorf("Failed to add route to expedition: %s", err.Error())
	}

	return nil
}

func (e *ExpeditionService) RenameExpedition(expeditionId, name string) error {
	summary := slice.Find(
		e.Index.Expeditions,
		func(s models.ExpeditionSummary) bool { return s.ID == expeditionId },
	)

	if summary == nil {
		return fmt.Errorf("Failed to find expedition in index")
	}

	expedition, err := summary.LoadFull()
	if err != nil {
		return fmt.Errorf("Failed to load expedition")
	}

	if !expedition.IsEditable() {
		return fmt.Errorf("Expedition is not editable")
	}

	prevName := expedition.Name
	prevLastUpdated := expedition.LastUpdated
	undo := func() {
		summary.Name = prevName
		summary.LastUpdated = prevLastUpdated
	}

	expedition.Name = name
	expedition.LastUpdated = time.Now()

	summary.Name = name
	summary.LastUpdated = expedition.LastUpdated

	t := database.NewTransaction("ExpeditionSummary.RenameExpedition")

	if err := models.TSaveExpedition(t, expedition); err != nil {
		undo()
		return fmt.Errorf("Failed to save expedition: %s", err.Error())
	}

	if err := models.TSaveIndex(t, e.Index); err != nil {
		undo()
		if rErr := t.Rewind(); rErr != nil {
			e.logger.Error("[ExpeditionService] RenameExpedition transaction rewind failed.")
		}
		return fmt.Errorf("Failed to save index: %s", err.Error())
	}

	if err := t.Apply(); err != nil {
		undo()
		e.logger.Error("[ExpeditionService] RenameExpedition transaction failed to apply.")
		return fmt.Errorf("Failed to rename expedition: %s", err.Error())
	}

	return nil
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
