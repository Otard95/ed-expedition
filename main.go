package main

import (
	"ed-expedition/lib/form"
	"ed-expedition/models"
	"embed"
	"flag"
	"os"

	"github.com/wailsapp/wails/v2"
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	logger := wailsLogger.NewDefaultLogger()

	journalDir := flag.String("j", os.Getenv("ED_EXPEDITION_JOURNAL_DIR"), "Elite Dangerous journal directory (default: $ED_EXPEDITION_JOURNAL_DIR)")
	flag.Parse()

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
		Linux: &linux.Options{
			Icon: icon,
		},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []any{
			app,
		},
		EnumBind: []interface{}{
			AllGalaxyStatus,
			models.AllFSDBoost,
			form.AllInputType,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
