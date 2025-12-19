package main

import (
	"ed-expedition/journal"
	"ed-expedition/lib/fs"
	"ed-expedition/services"
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	logger := wailsLogger.NewDefaultLogger()

	journalDir := flag.String("j", "./data/journals", "The path to the journal files")
	flag.Parse()
	if len(*journalDir) == 0 || !fs.IsDir(*journalDir) {
		logger.Error("Missing or invalid `-j` flag. You must provide a valid directory.")
		os.Exit(1)
	}

	journalWatcher, err := journal.NewWatcher(*journalDir, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to watch journal directory: %e", err))
		os.Exit(1)
	}
	defer journalWatcher.Close()

	expeditionService := services.NewExpeditionService(journalWatcher, logger)
	defer expeditionService.Stop()

	stateService := services.NewAppStateService(journalWatcher, logger)
	defer stateService.Stop()

	expeditionService.Start()
	stateService.Start()

	if stateService.State.LastKnownLocation != nil {
		err := journalWatcher.Sync(stateService.State.LastKnownLocation.Timestamp)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to sync journal: %e", err))
			os.Exit(1)
		}
	}

	journalWatcher.Start()

	app := NewApp(logger, stateService, expeditionService)

	err = wails.Run(&options.App{
		Title:     "ed-expedition",
		Width:     1024,
		Height:    768,
		MaxWidth:  4096,
		MaxHeight: 2160,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []any{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
