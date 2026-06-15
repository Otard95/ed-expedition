package plotters

import (
	"ed-expedition/database"
	"ed-expedition/lib/form"
	"ed-expedition/lib/job"
	"ed-expedition/lib/ptr"
	"ed-expedition/models"
	"ed-expedition/services"
	"math"
	"slices"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type Plotter interface {
	Plot(
		from, to string,
		inputs form.InputValues,
		loadout *models.Loadout,
		logger wailsLogger.Logger,
		tracker *job.ProgressTracker,
	) (*models.Route, error)
	ProgressType() job.PhaseType
	InputConfig() form.InputConfig
	String() string
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

func containsAnyOfClasses(systems []*services.GalaxySystem, classes []database.StarClass) bool {
	return slices.ContainsFunc(systems, func(s *services.GalaxySystem) bool { return slices.Contains(classes, s.StarClass) })
}

func computeFSDBoostForRoute(route *models.Route, maxJumpRange float64) {
	for i := 0; i < len(route.Jumps)-1; i++ {
		if route.Jumps[i].FSDBoost != nil {
			continue
		}
		if route.Jumps[i+1].Distance >= maxJumpRange {
			route.Jumps[i].FSDBoost = ptr.New(calculateMinFSDBoost(
				route.Jumps[i+1].Distance, maxJumpRange,
			))
		}
	}
}

func calculateMinFSDBoost(dst, maxJumpRange float64) models.FSDBoost {
	boost := int(math.Ceil((dst/maxJumpRange - 1) * 4))
	switch boost {
	case 1:
		return models.FSDBoostInjectionBasic
	case 2:
		return models.FSDBoostInjectionStandard
	case 3, 4:
		return models.FSDBoostInjectionPremium
	default:
		return models.FSDBoostNeutron
	}
}
