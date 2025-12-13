package main

import (
	"context"
	"ed-expedition/models"
	"ed-expedition/services"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

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
