package plotters

import (
	"ed-expedition/models"
	"strconv"
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
	Plot(from, to string, inputs PlotterInputs, loadout *models.Loadout) (*models.Route, error)
	InputConfig() PlotterInputConfig
	String() string
}

// getBoolInput retrieves a boolean input, encoding as "1" or "0"
func getBoolInput(inputs PlotterInputs, key string, defaultValue bool) string {
	val, ok := inputs[key]
	if !ok {
		return encodeBool(defaultValue)
	}
	return val // Already encoded as "1" or "0" per the type spec
}

// getNumberInput retrieves a number input as a string
func getNumberInput(inputs PlotterInputs, key string, defaultValue string) string {
	val, ok := inputs[key]
	if !ok {
		return defaultValue
	}
	return val
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
