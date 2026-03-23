package services

import (
	"ed-expedition/database"
	"ed-expedition/lib/vec"
	"errors"

	"golang.org/x/exp/constraints"
)

var ErrGalaxyNotReady = errors.New("galaxy database is not ready")
var (
	x       = vec.NewVec3[float64](1, 0, 0)
	y       = vec.NewVec3[float64](0, 1, 0)
	z       = vec.NewVec3[float64](0, 0, 1)
	corners = []vec.Vec3[float64]{ // Uniform opints on a unit-sphere
		vec.NewVec3[float64](1, 1, 1).Norm(),
		vec.NewVec3[float64](1, 1, -1).Norm(),
		vec.NewVec3[float64](1, -1, 1).Norm(),
		vec.NewVec3[float64](1, -1, -1).Norm(),
		vec.NewVec3[float64](-1, 1, 1).Norm(),
		vec.NewVec3[float64](-1, 1, -1).Norm(),
		vec.NewVec3[float64](-1, -1, 1).Norm(),
		vec.NewVec3[float64](-1, -1, -1).Norm(),
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

func (s *GalaxyService) GetSystemsAround(pos vec.Vec3[float64], radius float64) []*GalaxySystem {
	hilbertIndices := make([]int, 0, 9)
	for _, corner := range corners {
		hilbertIndices = append(
			hilbertIndices,
			database.Hilbert(database.NormalizeCoord(corner.Scale(radius).Add(pos).Unpack())),
		)
	}
	hilbertIndices = append(hilbertIndices, database.Hilbert(database.NormalizeCoord(pos.Unpack())))
	avgDiff := avg(diffs(hilbertIndices))

	groups := make([][]int, 0, 4)

	g := []int{}
	for i := 0; i < len(hilbertIndices)-1; i++ {
		g = append(g, hilbertIndices[i])
		if abs(hilbertIndices[i]-hilbertIndices[i+1]) > avgDiff {
			groups = append(groups, g)
			g = []int{}
		}
	}
	g = append(g, hilbertIndices[len(hilbertIndices)-1])
	groups = append(groups, g)

	return []*GalaxySystem{}
}

func diffs(values []int) []int {
	d := make([]int, 0, len(values)*(len(values)-1))

	for i, v1 := range values {
		for j, v2 := range values {
			if i == j {
				continue
			}

			d = append(d, abs(v1-v2))
		}
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
	return sum(values) / len(values)
}
