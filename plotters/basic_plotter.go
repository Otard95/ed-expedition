package plotters

import (
	"ed-expedition/lib/slice"
	"ed-expedition/lib/vec"
	"ed-expedition/models"
	"ed-expedition/services"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
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

	searchRadius := 20.0
	targetJumpDistance := parseFloat(getNumberInput(inputs, "target_jump_distance", "20"))
	scoopableOnly := getBoolInput(inputs, "scoopable_only", false) == "1"

	vFromTo := toSystem.Position.Sub(fromSystem.Position)
	vDir := vFromTo.Norm()
	vJump := vDir.Scale(targetJumpDistance)
	totalDist := vFromTo.Len()
	estimatedSteps := int(math.Ceil(totalDist / targetJumpDistance))

	logger.Debug(fmt.Sprintf("%s total_distance=%.2f ly target_jump=%.2f ly estimated_steps=%d search_radius=%.1f scoopable_only=%v",
		tag, totalDist, targetJumpDistance, estimatedSteps, searchRadius, scoopableOnly))
	logger.Debug(fmt.Sprintf("%s direction vector=%v jump vector=%v", tag, vDir, vJump))

	type candidateResult struct {
		candidates []*services.GalaxySystem
		target     vec.Vec3
		step       int
	}
	candidatesCh := make(chan candidateResult, 2)
	go func() {
		defer close(candidatesCh)
		for i := 1; float64(i)*targetJumpDistance < totalDist; i++ {
			target := fromSystem.Position.Add(vJump.Scale(float64(i)))
			logger.Debug(fmt.Sprintf("%s step %d: searching around %v radius=%.1f", tag, i, target, searchRadius))

			systems, err := p.GalaxyQuerier.GetSystemsAround(target, searchRadius)
			if err != nil || len(systems) == 0 {
				if err != nil {
					logger.Debug(fmt.Sprintf("%s step %d: initial search failed: %v, retrying with radius=%.1f", tag, i, err, searchRadius*2))
				} else {
					logger.Debug(fmt.Sprintf("%s step %d: no systems found, retrying with radius=%.1f", tag, i, searchRadius*2))
				}
				systems, err = p.GalaxyQuerier.GetSystemsAround(target, searchRadius*2)
				if err != nil || len(systems) == 0 {
					if err != nil {
						logger.Debug(fmt.Sprintf("%s step %d: retry also failed: %v, aborting", tag, i, err))
					} else {
						logger.Debug(fmt.Sprintf("%s step %d: retry also found 0 systems, aborting", tag, i))
					}
					return
				}
			}
			logger.Debug(fmt.Sprintf("%s step %d: found %d candidates", tag, i, len(systems)))

			candidatesCh <- candidateResult{
				candidates: systems,
				target:     target,
				step:       i,
			}
		}
	}()

	systems := make([]*services.GalaxySystem, 0, estimatedSteps+1)
	systems = append(systems, fromSystem)
	for result := range candidatesCh {
		best := 0
		bestDst := math.Inf(1)
		for i, candidate := range result.candidates {
			if scoopableOnly && !candidate.IsScoopable() {
				continue
			}
			dst := result.target.Distance(candidate.Position)
			if dst < bestDst {
				best = i
				bestDst = dst
			}
		}
		selected := result.candidates[best]
		logger.Debug(fmt.Sprintf("%s step %d: selected %q id=%d dist_from_target=%.2f ly scoopable=%v",
			tag, result.step, selected.Name, selected.Id, bestDst, selected.IsScoopable()))
		systems = append(systems, selected)
	}
	systems = append(systems, toSystem)

	plotterParams := make(map[string]any, len(inputs)+2)
	plotterParams["from"] = from
	plotterParams["to"] = to
	for key, value := range inputs {
		plotterParams[key] = value
	}

	plotterMetadata := make(map[string]any, 1)
	plotterMetadata["search_radius"] = searchRadius

	var prevPos *vec.Vec3
	jumps := slice.Map(systems, func(s *services.GalaxySystem) models.RouteJump {
		dst := 0.0
		if prevPos != nil {
			dst = s.Position.Distance(*prevPos)
		}
		prevPos = &s.Position
		pos := s.Position.Clone()

		return models.RouteJump{
			SystemName: s.Name,
			SystemID:   int64(s.Id),
			Scoopable:  s.IsScoopable(),
			MustRefuel: s.IsScoopable(),
			Distance:   dst,
			FuelInTank: nil,
			FuelUsed:   nil,
			HasNeutron: nil,
			Position:   &pos,
		}
	})

	route := models.Route{
		ID:              uuid.New().String(),
		Name:            fmt.Sprintf("%s → %s", fromSystem.Name, toSystem.Name),
		Plotter:         "basic_plotter",
		PlotterParams:   plotterParams,
		PlotterMetadata: plotterMetadata,
		Jumps:           jumps,
		CreatedAt:       time.Now(),
	}

	logger.Info(fmt.Sprintf("%s route generated: %d jumps, total_distance=%.2f ly", tag, len(jumps), totalDist))
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
