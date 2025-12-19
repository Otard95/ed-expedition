package main

import (
	"context"
	"ed-expedition/models"
	"ed-expedition/plotters"
	"ed-expedition/services"
	"fmt"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

var availablePlotters = map[string]plotters.Plotter{
	"spansh_galaxy_plotter": plotters.SpanshGalaxyPlotter{},
}

type App struct {
	ctx               context.Context
	logger            wailsLogger.Logger
	stateService      *services.AppStateService
	expeditionService *services.ExpeditionService
}

func NewApp(logger wailsLogger.Logger, stateService *services.AppStateService, expeditionService *services.ExpeditionService) *App {
	return &App{
		logger:            logger,
		stateService:      stateService,
		expeditionService: expeditionService,
	}
}

// startup is called by Wails. We save the context to enable runtime method calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetExpeditionSummaries() []models.ExpeditionSummary {
	return a.expeditionService.Index.Expeditions
}

func (a *App) CreateExpedition() (string, error) {
	return a.expeditionService.CreateExpedition()
}

func (a *App) LoadExpedition(id string) (*models.Expedition, error) {
	return models.LoadExpedition(id)
}

func (a *App) LoadRoutes(expeditionId string) ([]*models.Route, error) {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return nil, err
	}

	routes, err := expedition.LoadRoutes()
	if err != nil {
		return nil, err
	}

	return routes, nil
}

func (a *App) GetPlotterOptions() map[string]string {
	options := make(map[string]string, len(availablePlotters))

	for k, v := range availablePlotters {
		options[k] = v.String()
	}

	return options
}

func (a *App) GetPlotterInputConfig(plotterId string) (plotters.PlotterInputConfig, error) {
	if plotter, ok := availablePlotters[plotterId]; ok {
		return plotter.InputConfig(), nil
	}

	return plotters.PlotterInputConfig{}, fmt.Errorf("Unknown plotter id '%s'", plotterId)
}

func (a *App) PlotRoute(expeditionId, plotterId, from, to string, inputs plotters.PlotterInputs) (*models.Route, error) {
	plotter, ok := availablePlotters[plotterId]
	if !ok {
		return nil, fmt.Errorf("Unknown plotter id '%s'", plotterId)
	}

	loadout := a.stateService.State.LastKnownLoadout
	if loadout == nil {
		return nil, fmt.Errorf("No ship loadout available - please load game first")
	}

	route, err := plotter.Plot(from, to, inputs, loadout)
	if err != nil {
		return nil, err
	}

	if err := a.expeditionService.AddRouteToExpedition(expeditionId, route); err != nil {
		return nil, fmt.Errorf("failed to add route to expedition: %w", err)
	}

	return route, nil
}

func (a *App) DeleteExpedition(id string) error {
	return a.expeditionService.DeleteExpedition(id)
}

func (a *App) RenameExpedition(id, name string) error {
	return a.expeditionService.RenameExpedition(id, name)
}

func (a *App) RemoveRouteFromExpedition(expeditionId, routeId string) error {
	return a.expeditionService.RemoveRouteFromExpedition(expeditionId, routeId)
}

func (a *App) CreateLink(expeditionId string, from, to models.RoutePosition) error {
	return a.expeditionService.CreateLink(expeditionId, from, to)
}

func (a *App) DeleteLink(expeditionId, linkId string) error {
	return a.expeditionService.DeleteLink(expeditionId, linkId)
}

func (a *App) StartExpedition(expeditionId string) error {
	return a.expeditionService.StartExpedition(expeditionId)
}

type LoadActiveExpeditionPayload struct {
	Expedition *models.Expedition
	BakedRoute *models.Route
}

func (a *App) LoadActiveExpedition(expeditionId string) (*LoadActiveExpeditionPayload, error) {
	expedition, err := a.expeditionService.Index.LoadActiveExpedition()
	if err != nil {
		return nil, err
	}

	bakedRoute, err := expedition.LoadBaked()
	if err != nil {
		return nil, err
	}

	return &LoadActiveExpeditionPayload{
		Expedition: expedition,
		BakedRoute: bakedRoute,
	}, nil
}
