package plotters

import (
	"ed-expedition/lib/vec"
	"ed-expedition/models"
	"ed-expedition/services"
	"errors"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/google/uuid"
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

var (
	ErrorNoScoopable = errors.New("Found no scoopable system")
)

type GalaxyQueryier interface {
	GetSystemWithName(name string) (*services.GalaxySystem, error)
	GetSystemsAround(pos vec.Vec3, radius float64) ([]*services.GalaxySystem, error)
}

type BasicPlotter struct {
	GalaxyQuerier GalaxyQueryier
}

func (p BasicPlotter) String() string { return "Basic Built-in Plotter" }

func (p BasicPlotter) Plot(
	from, to string,
	inputs PlotterInputs,
	loadout *models.Loadout,
	logger wailsLogger.Logger,
) (*models.Route, error) {
	tag := "[BasicPlotter]"
	logger.Info(fmt.Sprintf("%s plotting route: %q -> %q", tag, from, to))

	fromSystem, err := p.GalaxyQuerier.GetSystemWithName(from)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("%s from system: %q id=%d pos=%v", tag, fromSystem.Name, fromSystem.Id, fromSystem.Position))

	toSystem, err := p.GalaxyQuerier.GetSystemWithName(to)
	if err != nil {
		return nil, err
	}
	logger.Debug(fmt.Sprintf("%s to system: %q id=%d pos=%v", tag, toSystem.Name, toSystem.Id, toSystem.Position))

	fsd, err := getFsd(loadout.FSD.Item)
	if err != nil {
		return nil, fmt.Errorf("Failed to get the FSD module data: %s", err.Error())
	}
	maxRange := maxJumpRange(loadout, fsd)

	searchRadius := 20.0
	targetJumpDistance := min(getNumberInput(inputs, "target_jump_distance", 20), maxRange)
	scoopableOnly := getBoolInput(inputs, "scoopable_only", false)

	jumps, err := p.findRoute(loadout, fsd, fromSystem, toSystem, maxRange, loadout.FuelCapacity.Main, targetJumpDistance, scoopableOnly, logger, tag, 0)
	if err != nil {
		return nil, err
	}
	jumps = append(jumps, models.RouteJump{
		SystemName: fromSystem.Name,
		SystemID:   int64(fromSystem.Id),
		Scoopable:  fromSystem.IsScoopable(),
		MustRefuel: false,
		Distance:   0,
		FuelInTank: &loadout.FuelCapacity.Main,
		FuelUsed:   nil,
		HasNeutron: nil,
		Position:   &fromSystem.Position,
	})
	slices.Reverse(jumps)

	plotterParams := make(map[string]any, len(inputs)+2)
	plotterParams["from"] = from
	plotterParams["to"] = to
	for key, value := range inputs {
		plotterParams[key] = value
	}

	plotterMetadata := make(map[string]any, 1)
	plotterMetadata["search_radius"] = searchRadius

	route := models.Route{
		ID:              uuid.New().String(),
		Name:            fmt.Sprintf("%s → %s", fromSystem.Name, toSystem.Name),
		Plotter:         "basic_plotter",
		PlotterParams:   plotterParams,
		PlotterMetadata: plotterMetadata,
		Jumps:           jumps,
		CreatedAt:       time.Now(),
	}

	logger.Info(fmt.Sprintf("%s route generated: %d jumps", tag, len(jumps)))
	return &route, nil
}

func (p BasicPlotter) InputConfig() PlotterInputConfig {
	return PlotterInputConfig{
		{
			Name:    "target_jump_distance",
			Label:   "Target Jump Distance",
			Type:    NumberInput,
			Default: "20",
			Info:    "Preferred jump length in light years. This is a routing target, not a hard max or min jump range.",
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

func (p BasicPlotter) getSystems(v vec.Vec3, r float64, mustHaveScoopable bool, logger wailsLogger.Logger, tag string) ([]*services.GalaxySystem, error) {
	logger.Debug(fmt.Sprintf("%s getSystems: pos=%v radius=%.1f mustHaveScoopable=%v", tag, v, r, mustHaveScoopable))
	s, err := p.GalaxyQuerier.GetSystemsAround(v, r)
	if err != nil || (mustHaveScoopable && !containsScoopable(s)) {
		if err != nil {
			logger.Debug(fmt.Sprintf("%s getSystems: initial search failed: %v, retrying radius=%.1f", tag, err, r*2))
		} else {
			logger.Debug(fmt.Sprintf("%s getSystems: no scoopable in %d results, retrying radius=%.1f", tag, len(s), r*2))
		}
		s, err = p.GalaxyQuerier.GetSystemsAround(v, r*2)
		if err != nil || (mustHaveScoopable && !containsScoopable(s)) {
			msg := "No scoopable systems"
			if err != nil {
				msg = err.Error()
			}
			logger.Debug(fmt.Sprintf("%s getSystems: retry also failed: %s", tag, msg))
			return nil, fmt.Errorf("Failed to get systems: %s", msg)
		}
	}
	logger.Debug(fmt.Sprintf("%s getSystems: found %d systems", tag, len(s)))
	return s, nil
}

func (p BasicPlotter) findRoute(
	loadout *models.Loadout,
	fsd *FSDModule,
	from, to *services.GalaxySystem,
	maxRange, fuelLeft, targetJumpDistance float64,
	scoopableOnly bool,
	logger wailsLogger.Logger,
	tag string,
	depth int,
) ([]models.RouteJump, error) {
	remaining := from.Position.Distance(to.Position)
	logger.Debug(fmt.Sprintf("%s findRoute[%d]: from=%q to=%q remaining=%.2f ly fuel=%.2f t", tag, depth, from.Name, to.Name, remaining, fuelLeft))

	if remaining < targetJumpDistance*1.1 {
		dst := remaining
		fCost := fuelCost(loadout, fsd, maxRange, dst)
		fuelInTank := fuelLeft - fCost
		logger.Debug(fmt.Sprintf("%s findRoute[%d]: close enough to destination, final jump=%.2f ly fuel_cost=%.2f t", tag, depth, dst, fCost))
		return []models.RouteJump{{
			SystemName: to.Name,
			SystemID:   int64(to.Id),
			Scoopable:  to.IsScoopable(),
			MustRefuel: false,
			Distance:   dst,
			FuelInTank: &fuelInTank,
			FuelUsed:   &fCost,
			HasNeutron: nil,
			Position:   &to.Position,
		}}, nil
	}

	shouldScoop := fuelLeft-fuelCost(loadout, fsd, maxRange, targetJumpDistance) < fuelCost(loadout, fsd, maxRange, targetJumpDistance)
	logger.Debug(fmt.Sprintf("%s findRoute[%d]: shouldScoop=%v scoopableOnly=%v", tag, depth, shouldScoop, scoopableOnly))

	jump, system, err := p.findJump(
		loadout,
		fsd,
		from, to,
		maxRange, fuelLeft, targetJumpDistance,
		scoopableOnly, shouldScoop,
		logger, tag, depth,
	)
	if err != nil {
		return nil, err
	}
	fuelLeft = *jump.FuelInTank

	jumps, err := p.findRoute(
		loadout,
		fsd,
		system, to,
		maxRange, fuelLeft, targetJumpDistance,
		scoopableOnly,
		logger, tag, depth+1,
	)
	if err == nil {
		return append(jumps, *jump), nil
	}
	if !errors.Is(err, ErrorNoScoopable) || shouldScoop || scoopableOnly {
		return nil, err
	}

	// If we could not find a route ahead because of a lack of fuel and we didn't
	// refuel on this jump we try finding a scoopable for this jump and try
	// again to plot a route ahead.
	logger.Debug(fmt.Sprintf("%s findRoute[%d]: no route ahead without fuel, retrying with forced scoop", tag, depth))

	jump, system, err = p.findJump(
		loadout,
		fsd,
		from, to,
		maxRange, fuelLeft, targetJumpDistance,
		scoopableOnly, true,
		logger, tag, depth,
	)
	if err != nil {
		return nil, err
	}
	fuelLeft = *jump.FuelInTank

	jumps, err = p.findRoute(
		loadout,
		fsd,
		system, to,
		maxRange, fuelLeft, targetJumpDistance,
		scoopableOnly,
		logger, tag, depth+1,
	)
	if err != nil {
		return nil, err
	}
	return append(jumps, *jump), nil
}

func (p BasicPlotter) findJump(
	loadout *models.Loadout,
	fsd *FSDModule,
	from, to *services.GalaxySystem,
	maxRange, fuelLeft, targetJumpDistance float64,
	scoopableOnly, shouldScoop bool,
	logger wailsLogger.Logger,
	tag string,
	depth int,
) (*models.RouteJump, *services.GalaxySystem, error) {
	target := to.Position.Sub(from.Position).Mag(targetJumpDistance).Add(from.Position)
	logger.Debug(fmt.Sprintf("%s findJump[%d]: from=%q target=%v shouldScoop=%v", tag, depth, from.Name, target, shouldScoop))

	candidates, err := p.getSystems(target, 20, shouldScoop, logger, tag)
	if err != nil {
		logger.Debug(fmt.Sprintf("%s findJump[%d]: no candidates found", tag, depth))
		return nil, nil, err
	}
	logger.Debug(fmt.Sprintf("%s findJump[%d]: %d candidates", tag, depth, len(candidates)))

	system, err := findBestCandidate(loadout, fsd, candidates, from, target, maxRange, fuelLeft, shouldScoop, scoopableOnly, logger, tag, depth)
	if err != nil {
		if shouldScoop || scoopableOnly {
			return nil, nil, ErrorNoScoopable
		}
		return nil, nil, err
	}

	distance := from.Position.Distance(system.Position)
	fCost := fuelCost(loadout, fsd, maxRange, distance)
	fuelLeft -= fCost
	mustScoop := (shouldScoop || scoopableOnly) && system.IsScoopable()
	if mustScoop {
		fuelLeft = loadout.FuelCapacity.Main
	}

	logger.Debug(fmt.Sprintf("%s findJump[%d]: selected %q dist=%.2f ly fuel_cost=%.2f t fuel_after=%.2f t scoop=%v",
		tag, depth, system.Name, distance, fCost, fuelLeft, mustScoop))

	return &models.RouteJump{
		SystemName: system.Name,
		SystemID:   int64(system.Id),
		Scoopable:  system.IsScoopable(),
		MustRefuel: mustScoop,
		Distance:   distance,
		FuelInTank: &fuelLeft,
		FuelUsed:   &fCost,
		HasNeutron: nil,
		Position:   &system.Position,
	}, system, nil
}

func findBestCandidate(
	loadout *models.Loadout,
	fsd *FSDModule,
	candidates []*services.GalaxySystem,
	prevSystem *services.GalaxySystem,
	target vec.Vec3,
	maxRange, fuel float64,
	shouldScoop, scoopableOnly bool,
	logger wailsLogger.Logger,
	tag string,
	depth int,
) (*services.GalaxySystem, error) {
	best := -1
	bestDst := math.Inf(1)
	skippedScoopable := 0
	skippedRange := 0
	skippedFuel := 0
	for i, s := range candidates {
		if (shouldScoop || scoopableOnly) && !s.IsScoopable() {
			skippedScoopable++
			continue
		}
		if prevSystem.Position.Distance(s.Position) > maxRange {
			skippedRange++
			continue
		}
		dst := target.Distance(s.Position)
		fCost := fuelCost(loadout, fsd, maxRange, prevSystem.Position.Distance(s.Position))
		if fuel-fCost < 0.2 {
			skippedFuel++
			continue
		}

		if dst < bestDst {
			best = i
			bestDst = dst
		}
	}

	logger.Debug(fmt.Sprintf("%s findBest[%d]: %d candidates, skipped: %d non-scoopable, %d out-of-range, %d insufficient-fuel",
		tag, depth, len(candidates), skippedScoopable, skippedRange, skippedFuel))

	if best == -1 {
		logger.Debug(fmt.Sprintf("%s findBest[%d]: no suitable system found", tag, depth))
		return nil, fmt.Errorf("Failed to find a suitable system.\n")
	}

	logger.Debug(fmt.Sprintf("%s findBest[%d]: best=%q dist_from_target=%.2f ly", tag, depth, candidates[best].Name, bestDst))
	return candidates[best], nil
}
