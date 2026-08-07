// region-boundaries processes manually marked region boundary points into clean
// geometric boundary definitions (radial lines and arcs from the galactic centre).
//
// Usage:
//
//	go run ./cmd/region-boundaries <regions-dir> <final-output-dir>
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ── Data types ───────────────────────────────────────────────────────────────

type Mark struct {
	ID    int     `json:"id"`
	R     float64 `json:"r"`
	Theta float64 `json:"theta"`
	X     float64 `json:"x"`
	Z     float64 `json:"z"`
}

type RegionData struct {
	name        string // directory slug
	displayName string // from name.txt
	nameX, nameZ float64
	hasNameCoords bool
	marks       []Mark
}

type TaggedMark struct {
	region string
	mark   Mark
}

type Cluster struct {
	ID      int      `json:"id"`
	Sources []string `json:"sources"`
	R       float64  `json:"r"`
	Theta   float64  `json:"theta"`
	X       float64  `json:"x"`
	Z       float64  `json:"z"`
}

type Edge struct {
	From  int     `json:"from"`
	To    int     `json:"to"`
	Type  string  `json:"type"`
	Theta float64 `json:"theta,omitempty"`
	R     float64 `json:"r,omitempty"`
}

type regionResult struct {
	reg        RegionData
	clusterIDs []int
	snapped    []Mark
	edges      []Edge
}

type RegionBoundary struct {
	Name   string `json:"name"`
	Points []int  `json:"points"`
	Edges  []Edge `json:"edges"`
}

// BoundaryOutput is the authoritative structure-of-arrays format.
// Vertices are polar [r, θ] from the galactic centre — single source of truth.
// Cartesian and edge parameters are derived at render time.
type BoundaryOutput struct {
	Vertices      [][]float64   `json:"vertices"`       // [[r, θ], ...] indexed by cluster ID
	Names         []string      `json:"names"`          // display name per region
	Edges         [][]int       `json:"edges"`          // vertex indices per region (loops back to start)
	EdgeType      [][]string    `json:"edge_type"`      // "arc"|"radial" per edge
	NamePositions []*[2]float64 `json:"name_positions"` // ED [x, z] override; null = use centroid
}

// ── Main ─────────────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: region-boundaries <regions-dir> <final-output-dir>")
		os.Exit(1)
	}
	regionsDir := os.Args[1]
	finalDir := os.Args[2]



	if err := os.MkdirAll(finalDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// ── 1. Load all marks ────────────────────────────────────────────────────

	entries, err := os.ReadDir(regionsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var regions []RegionData
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(regionsDir, e.Name(), "marks.json"))
		if err != nil {
			continue
		}
		var marks []Mark
		if err := json.Unmarshal(data, &marks); err != nil || len(marks) == 0 {
			continue
		}
		for i := range marks {
			marks[i].R, marks[i].Theta = toPolar(marks[i].X, marks[i].Z)
		}
		displayName := e.Name()
		if nameBytes, err := os.ReadFile(filepath.Join(regionsDir, e.Name(), "name.txt")); err == nil {
			displayName = strings.TrimSpace(string(nameBytes))
		}
		reg := RegionData{name: e.Name(), displayName: displayName, marks: marks}
		if coordData, err := os.ReadFile(filepath.Join(regionsDir, e.Name(), "name-coords.json")); err == nil {
			var coords struct {
				X float64 `json:"x"`
				Z float64 `json:"z"`
			}
			if err := json.Unmarshal(coordData, &coords); err == nil {
				reg.nameX, reg.nameZ = coords.X, coords.Z
				reg.hasNameCoords = true
			}
		}
		regions = append(regions, reg)
		fmt.Printf("Loaded %-30s %d marks\n", e.Name(), len(marks))
	}
	fmt.Printf("\n%d regions loaded\n\n", len(regions))

	// ── 2. Cluster co-located marks ──────────────────────────────────────────

	var all []TaggedMark
	for _, reg := range regions {
		for _, m := range reg.marks {
			all = append(all, TaggedMark{region: reg.name, mark: m})
		}
	}

	fmt.Println("Clustering shared intersections:")
	clusters, markToCluster := clusterMarks(all)

	sharedCount := 0
	for _, c := range clusters {
		if len(c.Sources) > 1 {
			sharedCount++
		}
	}
	fmt.Printf("\n%d intersection points total, %d shared\n\n", len(clusters), sharedCount)

	// ── 3. First-pass snap per region using cluster centroid positions ────────

	results := make([]regionResult, len(regions))
	for ri, reg := range regions {
		nm := len(reg.marks)
		clusterIDs := make([]int, nm)
		clusterPos := make([]Mark, nm)
		for i, mk := range reg.marks {
			cid := markToCluster[fmt.Sprintf("%s:%d", reg.name, mk.ID)]
			clusterIDs[i] = cid
			c := clusters[cid]
			clusterPos[i] = Mark{ID: mk.ID, R: c.R, Theta: c.Theta, X: c.X, Z: c.Z}
		}
		snapped, edges := firstPassSnap(clusterPos, clusterIDs)
		results[ri] = regionResult{
			reg:        reg,
			clusterIDs: clusterIDs,
			snapped:    snapped,
			edges:      edges,
		}
	}

	// ── 4. Global optimisation (3 passes) ────────────────────────────────────

	fmt.Println("Running global optimisation...")
	finalPos, globalEdge := globalOptimise(results, clusters)
	fmt.Println("Done.\n")

	// ── 5. Validation ────────────────────────────────────────────────────────

	fmt.Println("Validating connectivity:")
	issues := validate(results, finalPos, globalEdge)
	printValidation(issues)
	fmt.Println()

	// ── 6. Generate final segments ───────────────────────────────────────────

	var allSegments []float64
	seenEdges := map[edgeKey]bool{}
	var regionBoundaries []RegionBoundary

	for _, res := range results {
		ne := len(res.clusterIDs)
		var regSegs []float64
		finalEdges := make([]Edge, ne)

		for i := range res.clusterIDs {
			aci := res.clusterIDs[i]
			bci := res.clusterIDs[(i+1)%ne]
			ek := newEdgeKey(aci, bci)
			ge := globalEdge[ek]
			finalEdges[i] = Edge{From: aci, To: bci, Type: ge.Type, Theta: ge.Theta, R: ge.R}

			a := finalPos[aci]
			b := finalPos[bci]

			var segs [][4]float64
			if ge.Type == "radial" {
				segs = [][4]float64{{a.X, a.Z, b.X, b.Z}}
			} else {
				segs = arcSegments(ge.R, a.Theta, b.Theta, a.X, a.Z, b.X, b.Z)
			}

			for _, s := range segs {
				regSegs = append(regSegs, s[0], s[1], s[2], s[3])
			}
			if !seenEdges[ek] {
				seenEdges[ek] = true
				for _, s := range segs {
					allSegments = append(allSegments, s[0], s[1], s[2], s[3])
				}
			}
		}

		// Write debug files per region (intermediate, not the authoritative output)
		writeJSON(filepath.Join(regionsDir, res.reg.name, "boundary.json"), map[string]any{
			"points": res.clusterIDs,
			"edges":  finalEdges,
		})
		writeJSON(filepath.Join(regionsDir, res.reg.name, "segments.json"), regSegs)

		regionBoundaries = append(regionBoundaries, RegionBoundary{
			Name:   res.reg.displayName,
			Points: res.clusterIDs,
			Edges:  finalEdges,
		})
		fmt.Printf("%-30s %d points, %d segments\n", res.reg.displayName, ne, len(regSegs)/4)
	}

	// ── 7. Write final output ────────────────────────────────────────────────

	verts := make([][]float64, len(clusters))
	for i, c := range clusters {
		verts[i] = []float64{c.R, c.Theta}
	}
	nr := len(regionBoundaries)
	bout := BoundaryOutput{
		Vertices:      verts,
		Names:         make([]string, nr),
		Edges:         make([][]int, nr),
		EdgeType:      make([][]string, nr),
		NamePositions: make([]*[2]float64, nr),
	}
	for i, rb := range regionBoundaries {
		bout.Names[i] = rb.Name
		bout.Edges[i] = rb.Points
		types := make([]string, len(rb.Edges))
		for j, e := range rb.Edges {
			types[j] = e.Type
		}
		bout.EdgeType[i] = types
		if results[i].reg.hasNameCoords {
			coords := [2]float64{results[i].reg.nameX, results[i].reg.nameZ}
			bout.NamePositions[i] = &coords
		}
	}
	writeJSON(filepath.Join(finalDir, "boundaries.json"), bout)

	fmt.Printf("\n%d regions written to %s\n", len(regions), finalDir)
}



func writeJSON(path string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal:", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
	}
}
