package plotters

import (
	_ "embed"
	"encoding/json"
)

//go:embed spansh.data.json
var spanshDataJSON []byte

var spanshData *SpanshDataStruct

type SpanshDataStruct struct {
	Modules struct {
		Standard struct {
			FSD []SpanshFSDModule `json:"fsd"`
		} `json:"standard"`
		Internal struct {
			GFSB []SpanshGFSBModule `json:"gfsb"`
		} `json:"internal"`
	} `json:"Modules"`
}

type SpanshFSDModule struct {
	Symbol    string  `json:"symbol"`
	OptMass   float64 `json:"optmass"`
	MaxFuel   float64 `json:"maxfuel"`
	FuelMul   float64 `json:"fuelmul"`
	FuelPower float64 `json:"fuelpower"`
}

type SpanshGFSBModule struct {
	Symbol    string  `json:"symbol"`
	JumpBoost float64 `json:"jumpboost"`
}

func init() {
	if err := json.Unmarshal(spanshDataJSON, &spanshData); err != nil {
		panic(err)
	}
}
