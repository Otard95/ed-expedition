package plotters

import (
	"ed-expedition/database"
	"ed-expedition/lib/job"
	"ed-expedition/lib/ptr"
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

func (p BasicPlotter) ProgressType() job.PhaseType {
	return job.PhaseTypeObservable
}

type RoutePlottingContext struct {
	loadout            *models.Loadout
	fsd                *FSDModule
	from               *services.GalaxySystem
	to                 *services.GalaxySystem
	maxRange           float64
	totalDistance      float64
	targetJumpDistance float64
	starClasses        []database.StarClass
	logger             wailsLogger.Logger
	tracker            *job.ProgressTracker
}

func (r *RoutePlottingContext) withFrom(system *services.GalaxySystem) *RoutePlottingContext {
	clone := *r
	clone.from = system
	return &clone
}

func (p BasicPlotter) Plot(
	from, to string,
	inputs PlotterInputs,
	loadout *models.Loadout,
	logger wailsLogger.Logger,
	tracker *job.ProgressTracker,
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
	allowInjections := getBoolInput(inputs, "allow_injections", false)
	effectiveMaxRange := maxRange
	if allowInjections {
		effectiveMaxRange = maxRange * 2.0
	}

	searchRadius := 20.0
	targetJumpDistance := min(getNumberInput(inputs, "target_jump_distance", 20), effectiveMaxRange)
	starClasses := parseStarClassInput(getMultiSelectInput(inputs, "star_class", []string{"O", "B", "A", "F", "G", "K", "M"}))

	totalDistance := fromSystem.Position.Distance(toSystem.Position)
	tracker.SetTotal(totalDistance)

	ctx := RoutePlottingContext{
		loadout, fsd, fromSystem, toSystem,
		effectiveMaxRange, totalDistance, targetJumpDistance,
		starClasses, logger, tracker,
	}

	jumps, err := p.findRoute(&ctx, loadout.FuelCapacity.Main, tag, 0)
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
		Version:         1,
		ID:              uuid.New().String(),
		Name:            fmt.Sprintf("%s → %s", fromSystem.Name, toSystem.Name),
		Plotter:         "basic_plotter",
		PlotterParams:   plotterParams,
		PlotterMetadata: plotterMetadata,
		Jumps:           jumps,
		CreatedAt:       time.Now(),
	}
	computeFSDBoostForRoute(&route, maxRange)

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
			Name:    "star_class",
			Label:   "Star Classes",
			Type:    MultiSelectInput,
			Default: "O,B,A,F,G,K,M",
			Info:    "Prefer selected systems. May use others.",
			Options: []PlotterInputOption{
				{Value: "O", Label: "O-Type Stars", Description: ""},
				{Value: "B", Label: "B-Type Stars", Description: ""},
				{Value: "A", Label: "A-Type Stars", Description: ""},
				{Value: "F", Label: "F-Type Stars", Description: ""},
				{Value: "G", Label: "G-Type Stars", Description: ""},
				{Value: "K", Label: "K-Type Stars", Description: ""},
				{Value: "M", Label: "M-Type Stars", Description: ""},
				{Value: "L", Label: "L-Type Stars", Description: ""},
				{Value: "T", Label: "T-Type Stars", Description: ""},
				{Value: "Y", Label: "Y-Type Stars", Description: ""},
				{Value: "PROTO", Label: "Proto Stars", Description: ""},
				{Value: "CARBON", Label: "Carbon Stars", Description: ""},
				{Value: "WOLF-RAYET", Label: "Wolf-Rayet Stars", Description: ""},
				{Value: "WHITE-DWARF", Label: "White Dwarf Stars", Description: ""},
				{Value: "NON-SEQUENCE", Label: "Non Sequence Stars", Description: ""},
			},
		},
		{
			Name:    "allow_injections",
			Label:   "Allow FSD Injections",
			Type:    BoolInput,
			Default: "0",
			Info:    "Allow jumps that require FSD synthesis injections (up to premium, 2x range).",
		},
	}
}

var classInputToClassMap = map[string][]database.StarClass{
	"O": {database.StarClassO},
	"B": {database.StarClassB, database.StarClassBSuperGiant},
	"A": {database.StarClassA, database.StarClassASuperGiant},
	"F": {database.StarClassF, database.StarClassFSuperGiant},
	"G": {database.StarClassG, database.StarClassGSuperGiant},
	"K": {database.StarClassK, database.StarClassKGiant},
	"M": {database.StarClassM, database.StarClassMGiant, database.StarClassMSuperGiant},
	"L": {database.StarClassL},
	"T": {database.StarClassT},
	"Y": {database.StarClassY},

	"PROTO": {database.StarClassTTauri, database.StarClassHerbigAe},
	"CARBON": {
		database.StarClassC,
		database.StarClassCN,
		database.StarClassCJ,
		database.StarClassMSType,
		database.StarClassSType,
	},
	"WOLF-RAYET": {
		database.StarClassWolfRayet,
		database.StarClassWolfRayetC,
		database.StarClassWolfRayetN,
		database.StarClassWolfRayetNC,
		database.StarClassWolfRayetO,
	},
	"WHITE-DWARF": {
		database.StarClassWhiteDwarfD,
		database.StarClassWhiteDwarfDA,
		database.StarClassWhiteDwarfDAB,
		database.StarClassWhiteDwarfDAV,
		database.StarClassWhiteDwarfDAZ,
		database.StarClassWhiteDwarfDB,
		database.StarClassWhiteDwarfDBV,
		database.StarClassWhiteDwarfDBZ,
		database.StarClassWhiteDwarfDC,
		database.StarClassWhiteDwarfDCV,
		database.StarClassWhiteDwarfDQ,
	},
	"NON-SEQUENCE": {
		database.StarClassNeutron,
		database.StarClassBlackHole,
		database.StarClassSupermassiveBlkHole,
	},
}

func parseStarClassInput(inputClasses []string) []database.StarClass {
	selectedClasses := make([]database.StarClass, 0, len(inputClasses)*2)

	for _, inputClass := range inputClasses {
		classes, ok := classInputToClassMap[inputClass]
		if !ok {
			panic(fmt.Sprintf("Unknown class input '%s' passed to parseStarClassInput", inputClass))
		}
		selectedClasses = append(selectedClasses, classes...)
	}

	return selectedClasses
}

func (p BasicPlotter) getSystems(
	v vec.Vec3, r float64,
	mustHaveScoopable bool, preferred []database.StarClass,
	logger wailsLogger.Logger, tag string,
) ([]*services.GalaxySystem, error) {
	logger.Debug(fmt.Sprintf("%s getSystems: pos=%v radius=%.1f mustHaveScoopable=%v", tag, v, r, mustHaveScoopable))
	s, err := p.GalaxyQuerier.GetSystemsAround(v, r)
	if err != nil || !containsAnyOfClasses(s, preferred) || (mustHaveScoopable && !containsScoopable(s)) {
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
	ctx *RoutePlottingContext,
	fuelLeft float64,
	tag string,
	depth int,
) ([]models.RouteJump, error) {
	remaining := ctx.from.Position.Distance(ctx.to.Position)
	ctx.tracker.SetProgress(ctx.totalDistance - remaining)
	ctx.logger.Debug(fmt.Sprintf("%s findRoute[%d]: from=%q to=%q remaining=%.2f ly fuel=%.2f t", tag, depth, ctx.from.Name, ctx.to.Name, remaining, fuelLeft))

	if remaining < ctx.targetJumpDistance*1.1 {
		fCost := fuelCost(ctx.loadout, ctx.fsd, ctx.maxRange, remaining)
		ctx.logger.Debug(fmt.Sprintf("%s findRoute[%d]: close enough to destination, final jump=%.2f ly fuel_cost=%.2f t", tag, depth, remaining, fCost))
		return []models.RouteJump{{
			SystemName: ctx.to.Name,
			SystemID:   int64(ctx.to.Id),
			Scoopable:  ctx.to.IsScoopable(),
			MustRefuel: false,
			Distance:   remaining,
			FuelInTank: ptr.New(fuelLeft - fCost),
			FuelUsed:   &fCost,
			Position:   &ctx.to.Position,
		}}, nil
	}

	fc := fuelCost(ctx.loadout, ctx.fsd, ctx.maxRange, ctx.targetJumpDistance)
	shouldScoop := fuelLeft-fc < fc
	ctx.logger.Debug(fmt.Sprintf("%s findRoute[%d]: shouldScoop=%v", tag, depth, shouldScoop))

	jump, system, err := p.findJump(ctx, fuelLeft, shouldScoop, tag, depth)
	if err != nil {
		return nil, err
	}
	fuelLeft = *jump.FuelInTank

	jumps, err := p.findRoute(ctx.withFrom(system), fuelLeft, tag, depth+1)
	if err == nil {
		return append(jumps, *jump), nil
	}
	if !errors.Is(err, ErrorNoScoopable) || shouldScoop {
		return nil, err
	}

	// If we could not find a route ahead because of a lack of fuel and we didn't
	// refuel on this jump we try finding a scoopable for this jump and try
	// again to plot a route ahead.
	ctx.logger.Debug(fmt.Sprintf("%s findRoute[%d]: no route ahead without fuel, retrying with forced scoop", tag, depth))

	jump, system, err = p.findJump(ctx, fuelLeft, true, tag, depth)
	if err != nil {
		return nil, err
	}
	fuelLeft = *jump.FuelInTank

	jumps, err = p.findRoute(ctx.withFrom(system), fuelLeft, tag, depth+1)
	if err != nil {
		return nil, err
	}
	return append(jumps, *jump), nil
}

func (p BasicPlotter) findJump(
	ctx *RoutePlottingContext,
	fuelLeft float64,
	shouldScoop bool,
	tag string,
	depth int,
) (*models.RouteJump, *services.GalaxySystem, error) {
	target := ctx.to.Position.Sub(ctx.from.Position).Mag(ctx.targetJumpDistance).Add(ctx.from.Position)
	ctx.logger.Debug(fmt.Sprintf("%s findJump[%d]: from=%q target=%v shouldScoop=%v", tag, depth, ctx.from.Name, target, shouldScoop))

	candidates, err := p.getSystems(target, 20, shouldScoop, ctx.starClasses, ctx.logger, tag)
	if err != nil {
		ctx.logger.Debug(fmt.Sprintf("%s findJump[%d]: no candidates found", tag, depth))
		return nil, nil, err
	}
	ctx.logger.Debug(fmt.Sprintf("%s findJump[%d]: %d candidates", tag, depth, len(candidates)))

	system, err := findBestCandidate(ctx, candidates, target, fuelLeft, shouldScoop, tag, depth)
	if err != nil {
		if shouldScoop {
			return nil, nil, ErrorNoScoopable
		}
		return nil, nil, err
	}

	distance := ctx.from.Position.Distance(system.Position)
	fCost := fuelCost(ctx.loadout, ctx.fsd, ctx.maxRange, distance)
	fuelLeft -= fCost
	mustScoop := shouldScoop && system.IsScoopable()
	if mustScoop {
		fuelLeft = ctx.loadout.FuelCapacity.Main
	}

	ctx.logger.Debug(fmt.Sprintf("%s findJump[%d]: selected %q dist=%.2f ly fuel_cost=%.2f t fuel_after=%.2f t scoop=%v",
		tag, depth, system.Name, distance, fCost, fuelLeft, mustScoop))

	return &models.RouteJump{
		SystemName: system.Name,
		SystemID:   int64(system.Id),
		Scoopable:  system.IsScoopable(),
		MustRefuel: mustScoop,
		Distance:   distance,
		FuelInTank: &fuelLeft,
		FuelUsed:   &fCost,
		Position:   &system.Position,
	}, system, nil
}

func findBestCandidate(
	ctx *RoutePlottingContext,
	candidates []*services.GalaxySystem,
	target vec.Vec3,
	fuel float64,
	shouldScoop bool,
	tag string,
	depth int,
) (*services.GalaxySystem, error) {
	best := [3]int{-1, -1, -1}
	bestDst := [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}

	trackBest := func(typ, i int, dst float64) {
		if dst < bestDst[typ] {
			best[typ] = i
			bestDst[typ] = dst
		}
	}

	skippedScoopable := 0
	skippedRange := 0
	skippedFuel := 0
	remaining := ctx.from.Position.Distance(ctx.to.Position)
	for i, s := range candidates {
		if s.Position.Distance(ctx.to.Position) >= remaining {
			continue
		}
		if shouldScoop && !s.IsScoopable() {
			skippedScoopable++
			continue
		}
		if ctx.from.Position.Distance(s.Position) > ctx.maxRange {
			skippedRange++
			continue
		}
		dst := target.Distance(s.Position)
		fCost := fuelCost(ctx.loadout, ctx.fsd, ctx.maxRange, ctx.from.Position.Distance(s.Position))
		if fuel-fCost < 0.2 {
			skippedFuel++
			continue
		}

		if slices.Contains(ctx.starClasses, s.StarClass) {
			trackBest(0, i, dst)
		}
		if s.IsScoopable() {
			trackBest(1, i, dst)
		}
		trackBest(2, i, dst)
	}

	ctx.logger.Debug(fmt.Sprintf("%s findBest[%d]: %d candidates, skipped: %d non-scoopable, %d out-of-range, %d insufficient-fuel",
		tag, depth, len(candidates), skippedScoopable, skippedRange, skippedFuel))

	type choice struct {
		label string
		idx   int
		dst   float64
	}
	var pick *choice
	if best[0] != -1 {
		pick = &choice{"preferred", best[0], bestDst[0]}
	} else if best[1] != -1 {
		pick = &choice{"scoopable-fallback", best[1], bestDst[1]}
	} else if best[2] != -1 {
		pick = &choice{"any-fallback", best[2], bestDst[2]}
	}

	if pick == nil {
		ctx.logger.Debug(fmt.Sprintf("%s findBest[%d]: no suitable system found", tag, depth))
		return nil, fmt.Errorf("Failed to find a suitable system.\n")
	}

	ctx.logger.Debug(fmt.Sprintf("%s findBest[%d]: %s=%q dist_from_target=%.2f ly",
		tag, depth, pick.label, candidates[pick.idx].Name, pick.dst))
	return candidates[pick.idx], nil
}
