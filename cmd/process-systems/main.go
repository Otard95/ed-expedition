package main

import (
	"bufio"
	"ed-expedition/lib/fs"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

var Parallel int
var TopN int

type System struct {
	Id       int64   `json:"id64"`
	Name     string  `json:"name"`
	MainStar *string `json:"mainStar,omitempty"`
	UpTime   string  `json:"updateTime,omitempty"`
	Coords   Coord   `json:"coords"`
}
type Coord struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func main() {
	flag.IntVar(&Parallel, "parallel", 10, "Number of parallel JSON parsers")
	flag.IntVar(&TopN, "n", 3, "Number of closest pairs to find")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("Not enough arguments. Requires sub-command and filePath")
		os.Exit(1)
	}

	cmd := args[0]
	filePath := args[1]

	if !fs.IsFile(filePath) {
		fmt.Printf("No such filePath: %s", filePath)
		os.Exit(1)
	}

	start := time.Now()

	switch cmd {
	case "minmax":
		minmax(filePath)
	case "countcube":
		countCube(filePath)
	case "closestpair":
		closestPair(filePath)
	case "startypes":
		starTypes(filePath)
	default:
		fmt.Printf("Unknown sub-command: %s", cmd)
		os.Exit(1)
	}

	fmt.Printf("\nCompleted in %s\n", time.Since(start))
}

type MinMax struct {
	min, max float64
}

func NewMinMax() MinMax {
	return MinMax{math.MaxFloat64, -math.MaxFloat64}
}

const (
	X int = iota
	Y
	Z
)

// Sagittarius A* coordinates
var SagA = struct{ X, Y, Z float64 }{25.21875, -20.90625, 25899.96875}

const (
	CellSize      = 20.0
	MaxPairDist   = 2.0
	MaxPairDistSq = MaxPairDist * MaxPairDist
	CubeHalfSize  = 500.0
)

type CellKey struct {
	X, Y, Z int
}

func coordToCell(x, y, z float64) CellKey {
	return CellKey{
		X: int(math.Floor(x / CellSize)),
		Y: int(math.Floor(y / CellSize)),
		Z: int(math.Floor(z / CellSize)),
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

type SystemPair struct {
	A, B   System
	DistSq float64
}

type TopPairs struct {
	pairs []SystemPair
	n     int
}

func NewTopPairs(n int) *TopPairs {
	return &TopPairs{pairs: make([]SystemPair, 0, n), n: n}
}

func (tp *TopPairs) MaxDistSq() float64 {
	if len(tp.pairs) < tp.n {
		return MaxPairDistSq
	}
	return tp.pairs[len(tp.pairs)-1].DistSq
}

func (tp *TopPairs) TryInsert(pair SystemPair) bool {
	if pair.DistSq >= tp.MaxDistSq() {
		return false
	}

	// Find insertion point (keep sorted by DistSq ascending)
	insertIdx := len(tp.pairs)
	for i, p := range tp.pairs {
		if pair.DistSq < p.DistSq {
			insertIdx = i
			break
		}
	}

	// Insert
	if len(tp.pairs) < tp.n {
		tp.pairs = append(tp.pairs, SystemPair{})
	}
	copy(tp.pairs[insertIdx+1:], tp.pairs[insertIdx:])
	tp.pairs[insertIdx] = pair

	// Trim if over capacity
	if len(tp.pairs) > tp.n {
		tp.pairs = tp.pairs[:tp.n]
	}

	return true
}

func closestPair(filePath string) error {
	systemChan, err := iterateSystems(filePath)
	if err != nil {
		return fmt.Errorf("Failed to create system iterator: %s", err.Error())
	}

	// Cube bounds around Sag A*
	minX, maxX := SagA.X-CubeHalfSize, SagA.X+CubeHalfSize
	minY, maxY := SagA.Y-CubeHalfSize, SagA.Y+CubeHalfSize
	minZ, maxZ := SagA.Z-CubeHalfSize, SagA.Z+CubeHalfSize

	// Load systems into grid
	grid := make(map[CellKey][]System)
	count := 0
	for system := range systemChan {
		c := system.Coords
		if c.X < minX || c.X > maxX || c.Y < minY || c.Y > maxY || c.Z < minZ || c.Z > maxZ {
			continue
		}
		key := coordToCell(c.X, c.Y, c.Z)
		grid[key] = append(grid[key], system)
		count++
	}
	fmt.Printf("Loaded %d systems into %d cells\n", count, len(grid))

	// Find closest pairs
	top := NewTopPairs(TopN)

	for key, systems := range grid {
		cellMaxX := float64(key.X+1) * CellSize
		cellMaxY := float64(key.Y+1) * CellSize
		cellMaxZ := float64(key.Z+1) * CellSize

		for i, a := range systems {
			// Compare with remaining systems in same cell
			for j := i + 1; j < len(systems); j++ {
				if distSq, ok := checkPair(a, systems[j], top.MaxDistSq()); ok {
					top.TryInsert(SystemPair{A: a, B: systems[j], DistSq: distSq})
				}
			}

			// Check which boundaries we're near
			nearMaxX := a.Coords.X > cellMaxX-MaxPairDist
			nearMaxY := a.Coords.Y > cellMaxY-MaxPairDist
			nearMaxZ := a.Coords.Z > cellMaxZ-MaxPairDist

			// +X neighbor
			if nearMaxX {
				if neighbors, ok := grid[CellKey{key.X + 1, key.Y, key.Z}]; ok {
					for _, b := range neighbors {
						if distSq, ok := checkPair(a, b, top.MaxDistSq()); ok {
							top.TryInsert(SystemPair{A: a, B: b, DistSq: distSq})
						}
					}
				}
			}

			// +Y neighbor
			if nearMaxY {
				if neighbors, ok := grid[CellKey{key.X, key.Y + 1, key.Z}]; ok {
					for _, b := range neighbors {
						if distSq, ok := checkPair(a, b, top.MaxDistSq()); ok {
							top.TryInsert(SystemPair{A: a, B: b, DistSq: distSq})
						}
					}
				}
			}

			// +Z neighbor
			if nearMaxZ {
				if neighbors, ok := grid[CellKey{key.X, key.Y, key.Z + 1}]; ok {
					for _, b := range neighbors {
						if distSq, ok := checkPair(a, b, top.MaxDistSq()); ok {
							top.TryInsert(SystemPair{A: a, B: b, DistSq: distSq})
						}
					}
				}
			}

			// +XY diagonal
			if nearMaxX && nearMaxY {
				if neighbors, ok := grid[CellKey{key.X + 1, key.Y + 1, key.Z}]; ok {
					for _, b := range neighbors {
						if distSq, ok := checkPair(a, b, top.MaxDistSq()); ok {
							top.TryInsert(SystemPair{A: a, B: b, DistSq: distSq})
						}
					}
				}
			}

			// +XZ diagonal
			if nearMaxX && nearMaxZ {
				if neighbors, ok := grid[CellKey{key.X + 1, key.Y, key.Z + 1}]; ok {
					for _, b := range neighbors {
						if distSq, ok := checkPair(a, b, top.MaxDistSq()); ok {
							top.TryInsert(SystemPair{A: a, B: b, DistSq: distSq})
						}
					}
				}
			}

			// +YZ diagonal
			if nearMaxY && nearMaxZ {
				if neighbors, ok := grid[CellKey{key.X, key.Y + 1, key.Z + 1}]; ok {
					for _, b := range neighbors {
						if distSq, ok := checkPair(a, b, top.MaxDistSq()); ok {
							top.TryInsert(SystemPair{A: a, B: b, DistSq: distSq})
						}
					}
				}
			}

			// +XYZ corner
			if nearMaxX && nearMaxY && nearMaxZ {
				if neighbors, ok := grid[CellKey{key.X + 1, key.Y + 1, key.Z + 1}]; ok {
					for _, b := range neighbors {
						if distSq, ok := checkPair(a, b, top.MaxDistSq()); ok {
							top.TryInsert(SystemPair{A: a, B: b, DistSq: distSq})
						}
					}
				}
			}
		}
	}

	if len(top.pairs) == 0 {
		fmt.Println("No pairs found within 2 ly")
	} else {
		fmt.Printf("Top %d closest pairs:\n", len(top.pairs))
		for i, pair := range top.pairs {
			fmt.Printf("\n%d. %.6f ly\n", i+1, math.Sqrt(pair.DistSq))
			fmt.Printf("   %s (%d) at (%.2f, %.2f, %.2f)\n",
				pair.A.Name, pair.A.Id, pair.A.Coords.X, pair.A.Coords.Y, pair.A.Coords.Z)
			fmt.Printf("   %s (%d) at (%.2f, %.2f, %.2f)\n",
				pair.B.Name, pair.B.Id, pair.B.Coords.X, pair.B.Coords.Y, pair.B.Coords.Z)
		}
	}

	return nil
}

func checkPair(a, b System, maxDistSq float64) (distSq float64, closer bool) {
	dx := abs(a.Coords.X - b.Coords.X)
	if dx > MaxPairDist {
		return 0, false
	}
	dy := abs(a.Coords.Y - b.Coords.Y)
	if dy > MaxPairDist {
		return 0, false
	}
	dz := abs(a.Coords.Z - b.Coords.Z)
	if dz > MaxPairDist {
		return 0, false
	}

	distSq = dx*dx + dy*dy + dz*dz
	if distSq >= maxDistSq {
		return 0, false
	}
	return distSq, true
}

func starTypes(filePath string) error {
	systemChan, err := iterateSystems(filePath)
	if err != nil {
		return fmt.Errorf("Failed to create system iterator: %s", err.Error())
	}

	types := make(map[string]bool, 255)
	for system := range systemChan {
		if system.MainStar != nil {
			types[*system.MainStar] = true
		}
	}

	fmt.Printf("Found %d unique MainStar types:\n", len(types))
	for t := range types {
		fmt.Println(t)
	}
	return nil
}

func countCube(filePath string) error {
	systemChan, err := iterateSystems(filePath)
	if err != nil {
		return fmt.Errorf("Failed to create system iterator: %s", err.Error())
	}

	const halfSize = 500.0
	minX, maxX := SagA.X-halfSize, SagA.X+halfSize
	minY, maxY := SagA.Y-halfSize, SagA.Y+halfSize
	minZ, maxZ := SagA.Z-halfSize, SagA.Z+halfSize

	count := 0
	for system := range systemChan {
		if system.Coords.X >= minX && system.Coords.X <= maxX &&
			system.Coords.Y >= minY && system.Coords.Y <= maxY &&
			system.Coords.Z >= minZ && system.Coords.Z <= maxZ {
			count++
		}
	}

	fmt.Printf("Systems in 1000x1000x1000 cube around Sagittarius A*: %d\n", count)
	return nil
}

func minmax(filePath string) error {
	systemChan, err := iterateSystems(filePath)
	if err != nil {
		return fmt.Errorf("Failed to create system iterator: %s", err.Error())
	}

	mm := make([]MinMax, 3)
	for i := range mm {
		mm[i] = NewMinMax()
	}

	for system := range systemChan {
		mm[X].min = min(mm[X].min, system.Coords.X)
		mm[X].max = max(mm[X].max, system.Coords.X)
		mm[Y].min = min(mm[Y].min, system.Coords.Y)
		mm[Y].max = max(mm[Y].max, system.Coords.Y)
		mm[Z].min = min(mm[Z].min, system.Coords.Z)
		mm[Z].max = max(mm[Z].max, system.Coords.Z)
	}

	fmt.Printf("\nX min: %f, max: %f, range: %f\n", mm[X].min, mm[X].max, mm[X].max-mm[X].min)
	fmt.Printf("Y min: %f, max: %f, range: %f\n", mm[Y].min, mm[Y].max, mm[Y].max-mm[Y].min)
	fmt.Printf("Z min: %f, max: %f, range: %f\n", mm[Z].min, mm[Z].max, mm[Z].max-mm[Z].min)
	return nil
}

func iterateSystems(filePath string) (chan System, error) {
	lineChan, err := lineIterator(filePath)
	if err != nil {
		return nil, fmt.Errorf("Could not create scanner: %s", err.Error())
	}

	systemChan := make(chan System, Parallel*2)
	var wg sync.WaitGroup

	for range Parallel {
		wg.Add(1)
		go func() {
			for line := range lineChan {
				var system System
				err := json.Unmarshal([]byte(line), &system)
				if err != nil {
					fmt.Printf("ERROR: Failed to Unmarshal:\n  %s\n  Error: %s\n", line, err.Error())
					continue
				}

				systemChan <- system
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(systemChan)
	}()

	return systemChan, nil
}

func lineIterator(filePath string) (chan string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %s", err.Error())
	}

	scanner := bufio.NewScanner(file)
	lineChan := make(chan string, 10)

	go func() {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "{") {
				continue
			}
			if strings.HasSuffix(line, "},") {
				line = line[:len(line)-1]
			}

			lineChan <- line
		}

		close(lineChan)
		file.Close()
	}()

	return lineChan, nil
}
