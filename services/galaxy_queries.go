package services

import (
	"ed-expedition/database"
	"ed-expedition/lib/slice"
	"ed-expedition/lib/vec"
	"errors"
	"math"
	"slices"

	"golang.org/x/exp/constraints"
)

var ErrGalaxyNotReady = errors.New("galaxy database is not ready")
var (
	x      = vec.NewVec3[float64](1, 0, 0)
	y      = vec.NewVec3[float64](0, 1, 0)
	z      = vec.NewVec3[float64](0, 0, 1)
	sphere = []vec.Vec3{ // Uniform opints on a unit-sphere
		vec.NewVec3[float64](1, 1, 1).Norm(),
		vec.NewVec3[float64](1, 1, -1).Norm(),
		vec.NewVec3[float64](1, -1, 1).Norm(),
		vec.NewVec3[float64](1, -1, -1).Norm(),
		vec.NewVec3[float64](-1, 1, 1).Norm(),
		vec.NewVec3[float64](-1, 1, -1).Norm(),
		vec.NewVec3[float64](-1, -1, 1).Norm(),
		vec.NewVec3[float64](-1, -1, -1).Norm(),
		vec.NewVec3[float64](1, 0, 0),
		vec.NewVec3[float64](0, 1, 0),
		vec.NewVec3[float64](0, 0, 1),
		vec.NewVec3[float64](-1, 0, 0),
		vec.NewVec3[float64](0, -1, 0),
		vec.NewVec3[float64](0, 0, -1),
	}
)

func (s *GalaxyService) ValidateSystemName(name string) (canonicalName string, valid bool, err error) {
	if s.state != GalaxyStateReady || s.db == nil {
		return "", false, ErrGalaxyNotReady
	}

	sys, err := s.db.SystemByName(name)
	if err != nil {
		return "", false, err
	}
	if sys == nil {
		return "", false, nil
	}

	return sys.Name, true, nil
}

func (s *GalaxyService) AutocompleteSystems(prefix string, limit int) ([]string, error) {
	if s.state != GalaxyStateReady || s.db == nil {
		return nil, ErrGalaxyNotReady
	}

	systems, err := s.db.SystemsByPrefix(prefix, limit)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(systems))
	for i, sys := range systems {
		names[i] = sys.Name
	}
	return names, nil
}

func (s *GalaxyService) GetSystemWithName(name string) (*GalaxySystem, error) {
	if s.state != GalaxyStateReady || s.db == nil {
		return nil, ErrGalaxyNotReady
	}

	system, err := s.db.SystemByName(name)
	if err != nil {
		return nil, err
	}

	return transformDatabaseSystemToGalaxySystem(system), nil
}

func (s *GalaxyService) GetSystemsAround(pos vec.Vec3, radius float64) ([]*GalaxySystem, error) {
	if s.state != GalaxyStateReady || s.db == nil {
		return nil, ErrGalaxyNotReady
	}

	hilbertIndices := make([]int, 0, len(sphere)*2+1)
	for _, point := range sphere {
		hilbertIndices = append(
			hilbertIndices,
			database.Hilbert(vec.UnpackAs[uint32](database.NormalizeCoord(point.Scale(radius).Add(pos)))),
			database.Hilbert(vec.UnpackAs[uint32](database.NormalizeCoord(point.Scale(radius/2).Add(pos)))),
		)
	}
	hilbertIndices = append(hilbertIndices, database.Hilbert(vec.UnpackAs[uint32](database.NormalizeCoord(pos))))

	slices.Sort(hilbertIndices)

	avgDiff := avg(diffs(hilbertIndices))

	groups := make([][]int, 0, 8)

	g := []int{}
	for i := 0; i < len(hilbertIndices)-1; i++ {
		g = append(g, hilbertIndices[i])
		if abs(hilbertIndices[i]-hilbertIndices[i+1]) > avgDiff*2 {
			groups = append(groups, g)
			g = []int{}
		}
	}
	g = append(g, hilbertIndices[len(hilbertIndices)-1])
	groups = append(groups, g)

	largestGroupIndex := 0
	for i, g := range groups {
		if len(g) > len(groups[largestGroupIndex]) {
			largestGroupIndex = i
		}
	}

	avgDiffLargestGroup := avg(diffs(groups[largestGroupIndex]))

	ranges := make([][2]int, len(groups))
	for i, g := range groups {
		ranges[i] = [2]int{
			slices.Min(g) - avgDiffLargestGroup*3,
			slices.Max(g) + avgDiffLargestGroup*3,
		}
	}

	systems := slice.Flatten(slice.MapParallel(ranges, func(r [2]int) []*GalaxySystem {
		s, err := s.db.SystemsByHilbertRange(r[0], r[1])
		// TODO: Handle this
		if err != nil {
			return []*GalaxySystem{}
		}
		return slice.Map(s, transformDatabaseSystemToGalaxySystem)
	}))

	sqRadius := math.Pow(radius, 2)
	systems = slice.Filter(systems, func(s *GalaxySystem) bool {
		return s.Position.SqDistance(pos) <= sqRadius
	})

	return systems, nil
}

func transformDatabaseSystemToGalaxySystem(s *database.System) *GalaxySystem {
	return &GalaxySystem{
		Id:        s.Id,
		Name:      s.Name,
		Position:  database.DenormalizeCoord(vec.NewVec3(s.X, s.Y, s.Z)),
		StarClass: s.StarClass,
	}
}

func diffs(values []int) []int {
	if len(values) < 2 {
		return []int{}
	}

	d := make([]int, 0, len(values)-1)

	for i := 0; i < len(values)-1; i++ {
		d = append(d, abs(values[i]-values[i+1]))
	}

	return d
}

func abs[T constraints.Integer](v T) T {
	if v < 0 {
		return -v
	}
	return v
}

func sum(values []int) int {
	result := 0
	for _, v := range values {
		result += v
	}
	return result
}

func avg(values []int) int {
	if len(values) == 0 {
		return 0
	}
	return sum(values) / len(values)
}

func pack[T any](x, y, z T) [3]T {
	return [3]T{x, y, z}
}
