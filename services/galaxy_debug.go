package services

import (
	"ed-expedition/database"
	"ed-expedition/lib/slice"
	"ed-expedition/lib/vec"
	"fmt"
	"math"
	"slices"
	"sync"

	"time"
)

type HilbertGroupDebug struct {
	Position            vec.Vec3
	Points              []vec.Vec3
	HilbertPoints       [][]uint32
	DurationsMs         map[string]int64
	Radius              float64
	Indices             []int
	SortedIndices       []int
	Groups              [][]int
	AvgDiff             int
	Diffs               []int
	Ranges              [][2]int
	NumSystemsPreFilter int
	Systems             []*GalaxySystem
}

func (s *GalaxyService) DebugHilbertGroups(pos vec.Vec3, radius float64, useParallelQueries bool) *HilbertGroupDebug {
	start := time.Now()
	durations := map[string]int64{}
	mark := func(name string, since time.Time) {
		durations[name] = time.Since(since).Milliseconds()
	}

	phaseStart := time.Now()
	points := make([]vec.Vec3, 0, 9)
	for _, point := range sphere {
		points = append(
			points,
			point.Scale(radius).Add(pos),
			point.Scale(radius/2).Add(pos),
		)
	}
	points = append(points, pos)
	mark("points", phaseStart)

	phaseStart = time.Now()
	hilbertPoints := make([][]uint32, 0, len(points))
	for _, point := range points {
		hilbertPoints = append(
			hilbertPoints,
			vec.SliceOf[uint32](database.NormalizeCoord(point)),
		)
	}
	mark("normalize", phaseStart)

	phaseStart = time.Now()
	hilbertIndices := make([]int, 0, len(points))
	for _, point := range hilbertPoints {
		hilbertIndices = append(
			hilbertIndices,
			database.Hilbert(point[0], point[1], point[2]),
		)
	}
	mark("hilbert", phaseStart)

	phaseStart = time.Now()
	sortedIndices := slices.Clone(hilbertIndices)
	slices.Sort(sortedIndices)

	indexDiffs := diffs(sortedIndices)
	avgDiff := avg(indexDiffs)

	groups := make([][]int, 0, 4)

	g := []int{}
	for i := 0; i < len(sortedIndices)-1; i++ {
		g = append(g, sortedIndices[i])
		if abs(sortedIndices[i]-sortedIndices[i+1]) > int(float64(avgDiff)*2) {
			groups = append(groups, g)
			g = []int{}
		}
	}
	g = append(g, sortedIndices[len(sortedIndices)-1])
	groups = append(groups, g)

	slices.SortFunc(groups, func(a, b []int) int { return len(b) - len(a) })
	mark("grouping", phaseStart)

	phaseStart = time.Now()
	avgDiffLargestGroup := avg(diffs(groups[0]))

	ranges := make([][2]int, len(groups))

	for i, g := range groups {
		ranges[i] = [2]int{
			slices.Min(g) - avgDiffLargestGroup*3,
			slices.Max(g) + avgDiffLargestGroup*3,
		}
	}
	mark("ranges", phaseStart)

	phaseStart = time.Now()
	var systems []*database.System

	if useParallelQueries {
		systemsCh := make(chan []*database.System, len(ranges))
		wg := sync.WaitGroup{}
		for _, r := range ranges {
			wg.Add(1)
			go func() {
				defer wg.Done()
				s, err := s.db.SystemsByHilbertRange(r[0], r[1])
				if err != nil {
					return
				}
				systemsCh <- s
			}()
		}
		wg.Wait()
		close(systemsCh)

		systems = make([]*database.System, 0, 128)
		for s := range systemsCh {
			systems = append(systems, s...)
		}
	} else {
		var err error
		systems, err = s.db.SystemsByHilbertRanges(ranges)
		if err != nil {
			systems = []*database.System{}
		}
	}

	mark("query", phaseStart)

	numSystemsPreFilter := len(systems)
	normPos := database.NormalizeCoord(pos)
	systems = slice.Filter(systems, func(s *database.System) bool {
		return vec.NewVec3(s.X, s.Y, s.Z).SqDistance(normPos) <= math.Pow(radius*10, 2)
	})

	mark("total", start)

	s.logger.Info(fmt.Sprintf(
		"[GalaxyService](DebugHilbertGroups) pos=(%.1f, %.1f, %.1f) radius=%.1f points=%d groups=%d ranges=%d systems=%d durations_ms=%v",
		pos.X, pos.Y, pos.Z, radius, len(points), len(groups), len(ranges), len(systems), durations,
	))

	return &HilbertGroupDebug{
		Position:            pos,
		Points:              points,
		HilbertPoints:       hilbertPoints,
		DurationsMs:         durations,
		Radius:              radius,
		Indices:             hilbertIndices,
		SortedIndices:       sortedIndices,
		Groups:              groups,
		AvgDiff:             avgDiff,
		Diffs:               indexDiffs,
		Ranges:              ranges,
		NumSystemsPreFilter: numSystemsPreFilter,
		Systems:             slice.Map(systems, transformDatabaseSystemToGalaxySystem),
	}
}
