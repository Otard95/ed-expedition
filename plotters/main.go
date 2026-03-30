package plotters

import (
	"ed-expedition/models"
	"ed-expedition/services"
	"math"
	"slices"
	"strconv"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

// Should be the string returned from JS `typeof`
type PlotterInputType string

const (
	StringInput PlotterInputType = "string"
	NumberInput PlotterInputType = "number"
	BoolInput   PlotterInputType = "boolean"
)

type PlotterInputOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type PlotterInputFieldConfig struct {
	Name    string               `json:"name"`
	Label   string               `json:"label"`
	Type    PlotterInputType     `json:"type"`
	Default string               `json:"default"`
	Info    string               `json:"info,omitempty"`
	Options []PlotterInputOption `json:"options,omitempty"` // If set, renders as select/dropdown
}

type PlotterInputConfig = []PlotterInputFieldConfig

// Value encoding based on types:
//   - Bools:   true = "1" | false = "0"
//   - Numbers: "123.456"
type PlotterInputs map[string]string

type Plotter interface {
	Plot(
		from, to string,
		inputs PlotterInputs,
		loadout *models.Loadout,
		logger wailsLogger.Logger,
	) (*models.Route, error)
	InputConfig() PlotterInputConfig
	String() string
}

// getBoolInput retrieves a boolean input, encoding as "1" or "0"
func getBoolInput(inputs PlotterInputs, key string, defaultValue bool) bool {
	val, ok := inputs[key]
	if !ok {
		return defaultValue
	}
	return val == "1"
}

// getNumberInput retrieves a number input as a string
func getNumberInput(inputs PlotterInputs, key string, defaultValue float64) float64 {
	val, ok := inputs[key]
	if !ok {
		return defaultValue
	}
	return parseFloat(val)
}

// getStringInput retrieves a string input
func getStringInput(inputs PlotterInputs, key string, defaultValue string) string {
	val, ok := inputs[key]
	if !ok {
		return defaultValue
	}
	return val
}

// encodeBool converts a bool to "1" or "0"
func encodeBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// parseFloat parses a string to float64, returning 0 on error
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func resolveOptional[T any](val *T, defaultValue T) T {
	if val == nil {
		return defaultValue
	}
	return *val
}

func oneOf[T any](vals ...*T) *T {
	for _, val := range vals {
		if val != nil {
			return val
		}
	}
	return nil
}

func get[T any, R any](val *T, getter func(*T) *R) *R {
	if val != nil {
		return getter(val)
	}
	return nil
}

func maxJumpRange(loadout *models.Loadout, fsd *FSDModule) float64 {
	boost := getFsdBoost(loadout.FSDBooster)
	mass := loadout.UnladenMass + loadout.FuelCapacity.Main + loadout.FuelCapacity.Reserve
	optMass := resolveOptional(loadout.FSD.OptimalMass, fsd.OptMass)
	maxFuel := resolveOptional(loadout.FSD.MaxFuelPerJump, fsd.MaxFuel)

	maxRange := math.Pow(maxFuel/fsd.FuelMul, 1.0/fsd.FuelPower)*
		(optMass/mass) +
		boost

	return maxRange
}

func fuelCost(loadout *models.Loadout, fsd *FSDModule, maxRange, distance float64) float64 {
	maxFuel := resolveOptional(loadout.FSD.MaxFuelPerJump, fsd.MaxFuel)

	return math.Pow(distance/maxRange, fsd.FuelPower) * maxFuel
}

func containsScoopable(systems []*services.GalaxySystem) bool {
	return slices.ContainsFunc(systems, func(s *services.GalaxySystem) bool { return s.IsScoopable() })
}
