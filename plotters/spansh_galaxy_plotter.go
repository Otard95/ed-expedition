package plotters

import (
	"ed-expedition/lib/slice"
	"ed-expedition/models"
	"fmt"
	"strconv"
	"strings"
)

type SpanshGalaxyPlotter struct{}

func (p *SpanshGalaxyPlotter) String() string { return "Spansh Galaxy Plotter" }

func (p *SpanshGalaxyPlotter) Plot(
	from, to string,
	inputs PlotterInputs,
	loadout *models.Loadout,
) (*models.Route, error) {
	params, err := p.buildQueryParams(from, to, inputs, loadout)
	if err != nil {
		return nil, err
	}

	_ = params // TODO: Make HTTP request to Spansh API

	return nil, nil
}

func (p *SpanshGalaxyPlotter) buildQueryParams(
	from, to string,
	inputs PlotterInputs,
	loadout *models.Loadout,
) (map[string]string, error) {
	stdFsd := slice.Find(
		spanshData.Modules.Standard.FSD,
		func(fsd SpanshFSDModule) bool { return strings.ToLower(fsd.Symbol) == loadout.FSD.Item },
	)
	if stdFsd == nil {
		return nil, fmt.Errorf("Unexpected error! Failed to find standard fsd config")
	}

	params := make(map[string]string, 19)

	// Source and destination
	params["source"] = from
	params["destination"] = to

	// Routing options from inputs
	params["is_supercharged"] = getBoolInput(inputs, "is_supercharged", false)
	params["use_supercharge"] = getBoolInput(inputs, "use_supercharge", true)
	params["use_injections"] = getBoolInput(inputs, "use_injections", false)
	params["exclude_secondary"] = getBoolInput(inputs, "exclude_secondary", false)
	params["refuel_every_scoopable"] = getBoolInput(inputs, "refuel_every_scoopable", false)
	params["max_time"] = getNumberInput(inputs, "max_time", "60")
	params["cargo"] = getNumberInput(inputs, "cargo", "0")
	params["algorithm"] = getStringInput(inputs, "algorithm", "optimistic")

	var stdFsdBooster *SpanshGFSBModule
	if loadout.FSDBooster != nil {
		stdFsdBooster = slice.Find(
			spanshData.Modules.Internal.GFSB,
			func(fsdBooster SpanshGFSBModule) bool {
				return strings.ToLower(fsdBooster.Symbol) == *loadout.FSDBooster
			},
		)
	}

	params["optimal_mass"] = strconv.FormatFloat(resolveOptional(loadout.FSD.OptimalMass, stdFsd.OptMass), 'f', -1, 64)
	params["max_fuel_per_jump"] = strconv.FormatFloat(resolveOptional(loadout.FSD.MaxFuelPerJump, stdFsd.MaxFuel), 'f', -1, 64)
	params["fuel_multiplier"] = strconv.FormatFloat(stdFsd.FuelMul, 'f', -1, 64)
	params["fuel_power"] = strconv.FormatFloat(stdFsd.FuelPower, 'f', -1, 64)
	params["tank_size"] = strconv.FormatFloat(loadout.FuelCapacity.Main, 'f', -1, 64)
	params["internal_tank_size"] = strconv.FormatFloat(loadout.FuelCapacity.Reserve, 'f', -1, 64)
	params["base_mass"] = strconv.FormatFloat(loadout.UnladenMass+loadout.FuelCapacity.Main+loadout.FuelCapacity.Reserve, 'f', -1, 64)
	params["range_boost"] = strconv.FormatFloat(resolveOptional(
		get(stdFsdBooster, func(t *SpanshGFSBModule) *float64 { return &t.JumpBoost }),
		0,
	), 'f', -1, 64)
	params["supercharge_multiplier"] = "4"
	if loadout.FSD.Item == "int_hyperdrive_overcharge_size8_class5_overchargebooster_mkii" {
		params["supercharge_multiplier"] = "6"
	}

	return params, nil
}

func (p *SpanshGalaxyPlotter) InputConfig() PlotterInputConfig {
	return PlotterInputConfig{
		"is_supercharged": {
			Type:    BoolInput,
			Default: "0",
			Info:    "Is your ship already supercharged?",
		},
		"use_supercharge": {
			Type:    BoolInput,
			Default: "1",
			Info:    "Use neutron stars to supercharge your FSD",
		},
		"use_injections": {
			Type:    BoolInput,
			Default: "0",
			Info:    "Use FSD synthesis to boost when a neutron star is not available.",
		},
		"exclude_secondary": {
			Type:    BoolInput,
			Default: "0",
			Info:    "Prevent the system using secondary neutron and scoopable stars to help with the route",
		},
		"refuel_every_scoopable": {
			Type:    BoolInput,
			Default: "0",
			Info:    "Refuel every time you encounter a scoopable star",
		},
		"cargo": {
			Type:    NumberInput,
			Default: "0",
		},
		"algorithm": {
			Type:    StringInput,
			Default: "optimistic",
			Options: []PlotterInputOption{
				{
					Value:       "fuel",
					Label:       "Fuel",
					Description: "Prioritises saving fuel, will not scoop fuel or supercharge. Will make the smallest jumps possible in order to preserve fuel as much as possible.",
				},
				{
					Value:       "fuel_jumps",
					Label:       "Fuel Jumps",
					Description: "Prioritises saving fuel, will not scoop fuel or supercharge. Will make the smallest jumps possible in order to preserve fuel as much as possible. Once it has generated a route it will then attempt to minimise the number of jumps to use the entire fuel tank. It will attempt to save only enough fuel to recharge the internal fuel tank once. If you have generated a particularly long route it is likely that you will need to recharge more than once and as such you will most likely run out of fuel.",
				},
				{
					Value:       "guided",
					Label:       "Guided",
					Description: "Generates a standard Neutron Plotter Route and then uses that as a guide to follow. Penalises routes which diverge more than 100LY off the guide, meaning it preserves the general path of a typical Neutron Plotter route, but does not account for more optimal routes farther than 100LY away, and the calculation might time out if jumping through regions of space with sparse stars.",
				},
				{
					Value:       "optimistic",
					Label:       "Optimistic",
					Description: "Prioritises Neutron jumps. Penalises areas of the galaxy which have large gaps between neutron stars. Typically generates the fastest route with fewest total jumps.",
				},
				{
					Value:       "pessimistic",
					Label:       "Pessimistic",
					Description: "Prioritises calculation speed. Overestimates the average star distance to filter out routes. This means it calculates routes faster but the routes are typically less optimal.",
				},
			},
		},
	}
}
