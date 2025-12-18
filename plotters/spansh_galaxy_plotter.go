package plotters

import (
	"bytes"
	"ed-expedition/lib/slice"
	"ed-expedition/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type SpanshGalaxyPlotterResult struct {
	Job        string `json:"job"`
	Parameters struct {
		Algorithm             string      `json:"algorithm"`
		BaseMass              float64     `json:"base_mass"`
		Cargo                 int         `json:"cargo"`
		DestinationSystem     string      `json:"destination_system"`
		ExcludeSecondary      bool        `json:"exclude_secondary"`
		FuelMultiplier        float64     `json:"fuel_multiplier"`
		FuelPower             float64     `json:"fuel_power"`
		InternalTankSize      float64     `json:"internal_tank_size"`
		IsSupercharged        bool        `json:"is_supercharged"`
		MaxFuelPerJump        float64     `json:"max_fuel_per_jump"`
		OptimalMass           float64     `json:"optimal_mass"`
		RangeBoost            float64     `json:"range_boost"`
		RefuelEveryScoopable  int         `json:"refuel_every_scoopable"`
		ReserveSize           int         `json:"reserve_size"`
		ShipBuild             interface{} `json:"ship_build"`
		SourceSystem          string      `json:"source_system"`
		SuperchargeMultiplier int         `json:"supercharge_multiplier"`
		TankSize              int         `json:"tank_size"`
		UseInjections         bool        `json:"use_injections"`
		UseSupercharge        bool        `json:"use_supercharge"`
	} `json:"parameters"`
	Result struct {
		Jumps []struct {
			Distance              float64 `json:"distance"`
			DistanceToDestination float64 `json:"distance_to_destination"`
			FuelInTank            float64 `json:"fuel_in_tank"`
			FuelUsed              float64 `json:"fuel_used"`
			HasNeutron            bool    `json:"has_neutron"`
			ID64                  int64   `json:"id64"`
			IsScoopable           bool    `json:"is_scoopable"`
			MustRefuel            bool    `json:"must_refuel"`
			Name                  string  `json:"name"`
			X                     float64 `json:"x"`
			Y                     float64 `json:"y"`
			Z                     float64 `json:"z"`
		} `json:"jumps"`
		RefuelEveryScoopable bool `json:"refuel_every_scoopable"`
	} `json:"result"`
	State  string `json:"state"`
	Status string `json:"status"`
}

type SpanshGalaxyPlotter struct{}

func (p SpanshGalaxyPlotter) String() string { return "Spansh Galaxy Plotter" }

func (p SpanshGalaxyPlotter) Plot(
	from, to string,
	inputs PlotterInputs,
	loadout *models.Loadout,
) (*models.Route, error) {
	params, err := p.buildQueryParams(from, to, inputs, loadout)
	if err != nil {
		return nil, err
	}

	jobID, err := p.submitPlotRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to submit plot request: %w", err)
	}

	result, err := p.pollForResult(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plot result: %w", err)
	}

	route := p.transformToRoute(result, from, to, inputs)
	return route, nil
}

func (p SpanshGalaxyPlotter) submitPlotRequest(params map[string]string) (string, error) {
	formData := url.Values{}
	for key, value := range params {
		formData.Set(key, value)
	}

	resp, err := http.Post(
		"https://www.spansh.co.uk/api/generic/route",
		"application/x-www-form-urlencoded",
		bytes.NewBufferString(formData.Encode()),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("spansh API returned status %d: %s", resp.StatusCode, string(body))
	}

	var submitResponse struct {
		Job    string `json:"job"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&submitResponse); err != nil {
		return "", fmt.Errorf("failed to decode submit response: %w", err)
	}

	if submitResponse.Status != "queued" && submitResponse.Status != "ok" {
		return "", fmt.Errorf("unexpected initial status: %s", submitResponse.Status)
	}

	return submitResponse.Job, nil
}

func (p SpanshGalaxyPlotter) pollForResult(jobID string) (*SpanshGalaxyPlotterResult, error) {
	pollURL := fmt.Sprintf("https://www.spansh.co.uk/api/results/%s", jobID)
	maxAttempts := 60 // 60 attempts with 4s delay = 4 minute timeout
	pollDelay := 4 * time.Second
	logger := wailsLogger.NewDefaultLogger()

	for a := range maxAttempts {
		time.Sleep(pollDelay)
		logger.Info(fmt.Sprintf("[SpanshGalaxyPlotter] pull attempts %d", a))

		resp, err := http.Get(pollURL)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("spansh API returned status %d: %s", resp.StatusCode, string(body))
		}

		var result SpanshGalaxyPlotterResult
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to decode result: %w", err)
		}

		// Check if job is complete
		if result.Status == "ok" && result.State == "completed" {
			return &result, nil
		}

		// If job failed, return error
		if result.Status == "error" {
			return nil, fmt.Errorf("spansh plot job failed: %s", result.State)
		}
	}

	return nil, fmt.Errorf("plot request timed out after %d attempts", maxAttempts)
}

func (p SpanshGalaxyPlotter) transformToRoute(
	result *SpanshGalaxyPlotterResult,
	from, to string,
	inputs PlotterInputs,
) *models.Route {
	jumps := make([]models.RouteJump, len(result.Result.Jumps))

	for i, spanshJump := range result.Result.Jumps {
		// Convert integer fuel values to float64 pointers
		fuelInTank := float64(spanshJump.FuelInTank)
		fuelUsed := float64(spanshJump.FuelUsed)
		hasNeutron := spanshJump.HasNeutron

		jumps[i] = models.RouteJump{
			SystemName: spanshJump.Name,
			SystemID:   spanshJump.ID64,
			Scoopable:  spanshJump.IsScoopable,
			MustRefuel: spanshJump.MustRefuel,
			Distance:   float64(spanshJump.Distance),
			FuelInTank: &fuelInTank,
			FuelUsed:   &fuelUsed,
			HasNeutron: &hasNeutron,
			Position: &models.Position{
				X: spanshJump.X,
				Y: spanshJump.Y,
				Z: spanshJump.Z,
			},
		}
	}

	// Store all plotter parameters for reference
	plotterParams := make(map[string]any)
	plotterParams["from"] = from
	plotterParams["to"] = to
	for key, value := range inputs {
		plotterParams[key] = value
	}

	// Store Spansh metadata (job ID, original parameters)
	plotterMetadata := make(map[string]any)
	plotterMetadata["job_id"] = result.Job
	plotterMetadata["spansh_parameters"] = result.Parameters

	return &models.Route{
		ID:              result.Job, // Use Spansh job ID as route ID
		Name:            fmt.Sprintf("%s → %s", from, to),
		Plotter:         "spansh_galaxy",
		PlotterParams:   plotterParams,
		PlotterMetadata: plotterMetadata,
		Jumps:           jumps,
		CreatedAt:       time.Now(),
	}
}

func (p SpanshGalaxyPlotter) buildQueryParams(
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

func (p SpanshGalaxyPlotter) InputConfig() PlotterInputConfig {
	return PlotterInputConfig{
		{
			Name:    "is_supercharged",
			Label:   "Already Supercharged",
			Type:    BoolInput,
			Default: "0",
			Info:    "Is your ship already supercharged?",
		},
		{
			Name:    "use_supercharge",
			Label:   "Use Supercharge",
			Type:    BoolInput,
			Default: "1",
			Info:    "Use neutron stars to supercharge your FSD",
		},
		{
			Name:    "use_injections",
			Label:   "Use FSD Injections",
			Type:    BoolInput,
			Default: "0",
			Info:    "Use FSD synthesis to boost when a neutron star is not available.",
		},
		{
			Name:    "exclude_secondary",
			Label:   "Exclude Secondary Stars",
			Type:    BoolInput,
			Default: "0",
			Info:    "Prevent the system using secondary neutron and scoopable stars to help with the route",
		},
		{
			Name:    "refuel_every_scoopable",
			Label:   "Refuel Every Scoopable",
			Type:    BoolInput,
			Default: "0",
			Info:    "Refuel every time you encounter a scoopable star",
		},
		{
			Name:    "cargo",
			Label:   "Cargo",
			Type:    NumberInput,
			Default: "0",
			Info:    "Amount of cargo in tons",
		},
		{
			Name:    "algorithm",
			Label:   "Route Algorithm",
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
