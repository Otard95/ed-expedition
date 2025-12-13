package main

import (
	"context"
	"fmt"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type App struct {
	ctx    context.Context
	logger wailsLogger.Logger
}

func NewApp(logger wailsLogger.Logger) *App {
	return &App{logger: logger}
}

// startup is called by Wails. We save the context to enable runtime method calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
