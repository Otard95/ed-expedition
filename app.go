package main

import (
	"context"
	"ed-expedition/journal"
	"ed-expedition/models"
	"ed-expedition/plotters"
	"ed-expedition/services"
	"fmt"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var availablePlotters = map[string]plotters.Plotter{
	"spansh_galaxy_plotter": plotters.SpanshGalaxyPlotter{},
}

type App struct {
	ctx               context.Context
	logger            wailsLogger.Logger
	journalWatcher    *journal.Watcher
	stateService      *services.AppStateService
	expeditionService *services.ExpeditionService

	targetChan             chan *journal.FSDTargetEvent
	jumpHistoryChan        chan *models.JumpHistoryEntry
	completeExpeditionChan chan *models.Expedition
}

func NewApp(
	logger wailsLogger.Logger,
	journalWatcher *journal.Watcher,
	stateService *services.AppStateService,
	expeditionService *services.ExpeditionService,
) *App {
	return &App{
		logger:            logger,
		journalWatcher:    journalWatcher,
		stateService:      stateService,
		expeditionService: expeditionService,
	}
}

// startup is called by Wails. We save the context to enable runtime method calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.jumpHistoryChan = a.expeditionService.JumpHistory.Subscribe()

	go func() {
		for event := range a.jumpHistoryChan {
			runtime.EventsEmit(ctx, "JumpHistory", *event)
		}
	}()

	a.targetChan = a.journalWatcher.FSDTarget.Subscribe()

	go func() {
		for event := range a.targetChan {
			runtime.EventsEmit(ctx, "Target", *event)
		}
	}()

	a.completeExpeditionChan = a.expeditionService.CompleteExpedition.Subscribe()

	go func() {
		for event := range a.completeExpeditionChan {
			runtime.EventsEmit(ctx, "CompleteExpedition", *event)
		}
	}()
}
func (a *App) shutdown(ctx context.Context) {
	if a.jumpHistoryChan != nil {
		a.expeditionService.JumpHistory.Unsubscribe(a.jumpHistoryChan)
	}
	if a.targetChan != nil {
		a.journalWatcher.FSDTarget.Unsubscribe(a.targetChan)
	}
	if a.completeExpeditionChan != nil {
		a.expeditionService.CompleteExpedition.Unsubscribe(a.completeExpeditionChan)
	}
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

func (a *App) EndActiveExpedition() error {
	return a.expeditionService.EndActiveExpedition()
}

type LoadActiveExpeditionPayload struct {
	Expedition *models.Expedition
	BakedRoute *models.Route
}

func (a *App) LoadActiveExpedition() (*LoadActiveExpeditionPayload, error) {
	expedition, err := a.expeditionService.Index.LoadActiveExpedition()
	if err != nil {
		return nil, err
	}
	if expedition == nil {
		return nil, nil
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
