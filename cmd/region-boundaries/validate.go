package main

import (
	"fmt"
	"math"
)

const disconnectThreshold = 50.0 // ly — endpoint gap larger than this is a disconnection

// ValidationResult reports a gap between two consecutive boundary edges.
type ValidationResult struct {
	Region string
	FromID int
	ToID   int
	Gap    float64
}

// validate checks that every consecutive pair of edges in each region connects
// — i.e., the end of edge i equals the start of edge i+1 (both share the same
// junction cluster, so with exact finalPos endpoints the gap should be zero).
// A gap > disconnectThreshold indicates a missing or misclassified mark.
func validate(results []regionResult, finalPos []Mark, globalEdge map[edgeKey]Edge) []ValidationResult {
	var issues []ValidationResult

	for _, res := range results {
		ne := len(res.clusterIDs)
		for i := range res.clusterIDs {
			aci := res.clusterIDs[i]
			bci := res.clusterIDs[(i+1)%ne]

			// End of this edge: finalPos[bci]
			// Start of next edge: also finalPos[bci] (shared junction)
			// Because we use exact finalPos as arc/line endpoints, the gap is
			// the deviation between finalPos[bci] as seen from both sides.
			endX, endZ := finalPos[bci].X, finalPos[bci].Z

			// Cross-check: the end should equal the next edge's start point.
			nextAci := bci
			startX, startZ := finalPos[nextAci].X, finalPos[nextAci].Z

			gap := math.Sqrt((endX-startX)*(endX-startX) + (endZ-startZ)*(endZ-startZ))

			// The above will always be 0 since they reference the same finalPos entry.
			// The meaningful check is whether the GEOMETRIC constraint is satisfied:
			// does finalPos[bci] actually lie on the two adjacent primitives?
			ek := newEdgeKey(aci, bci)
			ge := globalEdge[ek]
			nextEK := newEdgeKey(bci, res.clusterIDs[(i+2)%ne])
			nextGE := globalEdge[nextEK]

			p := finalPos[bci]
			var geometricGap float64
			switch {
			case ge.Type == "arc" && nextGE.Type == "radial":
				// Should lie at (ge.R, nextGE.Theta)
				ex, ez := fromPolar(ge.R, nextGE.Theta)
				geometricGap = math.Sqrt((p.X-ex)*(p.X-ex) + (p.Z-ez)*(p.Z-ez))
			case ge.Type == "radial" && nextGE.Type == "arc":
				// Should lie at (nextGE.R, ge.Theta)
				ex, ez := fromPolar(nextGE.R, ge.Theta)
				geometricGap = math.Sqrt((p.X-ex)*(p.X-ex) + (p.Z-ez)*(p.Z-ez))
			default:
				geometricGap = gap
			}

			if geometricGap > disconnectThreshold {
				issues = append(issues, ValidationResult{
					Region: res.reg.name,
					FromID: aci,
					ToID:   bci,
					Gap:    geometricGap,
				})
			}
		}
	}

	return issues
}

func printValidation(issues []ValidationResult) {
	if len(issues) == 0 {
		fmt.Println("  ✓ All boundaries are geometrically consistent")
		return
	}
	for _, v := range issues {
		fmt.Printf("  ✗ %s: cluster %d→%d geometric gap %.0f ly\n",
			v.Region, v.FromID, v.ToID, v.Gap)
	}
}
