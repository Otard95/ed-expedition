package plotters

import (
	"ed-expedition/models"
	"ed-expedition/services"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type GalaxyQueryier interface {
	GetSystemsAround(x, y, z float64, radius float64) []*services.GalaxySystem
}

type BasicPlotter struct {
	galaxyQueryier GalaxyQueryier
}

func (p BasicPlotter) String() string { return "Basic Built-in Plotter" }

func (p BasicPlotter) Plot(
	from, to string,
	inputs PlotterInputs,
	loadout *models.Loadout,
	logger wailsLogger.Logger,
) (*models.Route, error) {
	return nil, nil
}

func (p BasicPlotter) InputConfig() PlotterInputConfig {
	return PlotterInputConfig{
		{
			Name:    "target_jump_distance",
			Label:   "Target Jump Distance",
			Type:    NumberInput,
			Default: "40",
			Info:    "Preferred jump length in light years. This is a routing target, not a hard max jump range.",
		},
		{
			Name:    "scoopable_only",
			Label:   "Scoopable Only",
			Type:    BoolInput,
			Default: "0",
			Info:    "Only consider systems whose main star is scoopable.",
		},
	}
}
