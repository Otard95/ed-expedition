package form

import (
	"strconv"
	"strings"
)

type InputType string

const (
	StringInput      InputType = "string"
	NumberInput      InputType = "number"
	BoolInput        InputType = "boolean"
	MultiSelectInput InputType = "multiselect"
	DirectoryInput   InputType = "directory"
)

var AllInputType = []struct {
	Value  InputType
	TSName string
}{
	{StringInput, "STRING"},
	{NumberInput, "NUMBER"},
	{BoolInput, "BOOLEAN"},
	{MultiSelectInput, "MULTISELECT"},
	{DirectoryInput, "DIRECTORY"},
}

type InputOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type InputFieldConfig struct {
	Name    string        `json:"name"`
	Label   string        `json:"label"`
	Type    InputType     `json:"type"`
	Default string        `json:"default"`
	Info    string        `json:"info,omitempty"`
	Options []InputOption `json:"options,omitempty"`
	Section string        `json:"section,omitempty"`
}

type InputConfig = []InputFieldConfig

// Value encoding:
//   - Bools:       "1" / "0"
//   - Numbers:     "123.456"
//   - MultiSelect: "a,b,c"
type InputValues map[string]string

func GetBool(inputs InputValues, key string, defaultValue bool) bool {
	val, ok := inputs[key]
	if !ok {
		return defaultValue
	}
	return val == "1"
}

func GetNumber(inputs InputValues, key string, defaultValue float64) float64 {
	val, ok := inputs[key]
	if !ok {
		return defaultValue
	}
	return ParseFloat(val)
}

func GetString(inputs InputValues, key string, defaultValue string) string {
	val, ok := inputs[key]
	if !ok {
		return defaultValue
	}
	return val
}

func GetMultiSelect(inputs InputValues, key string, defaultValue []string) []string {
	val, ok := inputs[key]
	if !ok {
		return defaultValue
	}
	return strings.Split(val, ",")
}

func EncodeBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func ParseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func EncodeNumber(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}
