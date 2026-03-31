package plotters

import (
	"ed-expedition/lib/slice"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed module.data.json
var moduleDataJSON []byte

var moduleData *ModuleDataStruct

type ModuleDataStruct struct {
	Modules struct {
		Standard struct {
			FSD []FSDModule `json:"fsd"`
		} `json:"standard"`
		Internal struct {
			GFSB []GFSBModule `json:"gfsb"`
		} `json:"internal"`
	} `json:"modules"`
}

type FSDModule struct {
	Symbol    string  `json:"symbol"`
	OptMass   float64 `json:"optmass"`
	MaxFuel   float64 `json:"maxfuel"`
	FuelMul   float64 `json:"fuelmul"`
	FuelPower float64 `json:"fuelpower"`
}

type GFSBModule struct {
	Symbol    string  `json:"symbol"`
	JumpBoost float64 `json:"jumpboost"`
}

func init() {
	if err := json.Unmarshal(moduleDataJSON, &moduleData); err != nil {
		panic(err)
	}
}

func getFsd(fsdItem string) (*FSDModule, error) {
	stdFsd := slice.Find(
		moduleData.Modules.Standard.FSD,
		func(fsd FSDModule) bool { return strings.ToLower(fsd.Symbol) == fsdItem },
	)
	if stdFsd == nil {
		return nil, fmt.Errorf("Unexpected error! Failed to find standard fsd config")
	}
	return stdFsd, nil
}

func getFsdBoost(booster *string) float64 {
	if booster == nil {
		return 0
	}

	stdFsdBooster := slice.Find(
		moduleData.Modules.Internal.GFSB,
		func(fsdBooster GFSBModule) bool {
			return strings.ToLower(fsdBooster.Symbol) == *booster
		},
	)
	if stdFsdBooster == nil {
		return 0
	}

	return stdFsdBooster.JumpBoost
}
