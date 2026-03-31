package main

import (
	"ed-expedition/lib/fs"
	"embed"
	"flag"
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

	app := NewApp(logger, *journalDir)

	err := wails.Run(&options.App{
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
		OnShutdown:       app.shutdown,
		Bind: []any{
			app,
		},
		EnumBind: []interface{}{
			AllGalaxyStatus,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
