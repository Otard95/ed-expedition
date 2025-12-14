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

func (a *App) LoadRoutes(expeditionId string) ([]models.Route, error) {
	expedition, err := models.LoadExpedition(expeditionId)
	if err != nil {
		return nil, err
	}

	routeMap, err := expedition.LoadRoutes()
	if err != nil {
		return nil, err
	}

	routes := make([]models.Route, 0, len(routeMap))
	for _, route := range routeMap {
		routes = append(routes, *route)
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

	return nil, fmt.Errorf("Unknown plotter id '%s'", plotterId)
}
