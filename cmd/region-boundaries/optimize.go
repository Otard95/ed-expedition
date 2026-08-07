package main

// firstPassSnap classifies edges, merges consecutive same-type runs, and snaps
// each vertex to the exact primitive intersection — all using the cluster
// centroid positions as input.
func firstPassSnap(clusterPos []Mark, clusterIDs []int) (snapped []Mark, edges []Edge) {
	n := len(clusterPos)
	m := clusterPos

	type raw struct {
		typ   string
		theta float64
		r     float64
	}
	rawEdges := make([]raw, n)
	for i := range m {
		typ, theta, r := classifyEdge(m[i], m[(i+1)%n])
		rawEdges[i] = raw{typ, theta, r}
	}

	edges = make([]Edge, n)
	for i := range m {
		edges[i] = Edge{
			From:  clusterIDs[i],
			To:    clusterIDs[(i+1)%n],
			Type:  rawEdges[i].typ,
			Theta: rawEdges[i].theta,
			R:     rawEdges[i].r,
		}
	}

	// Merge consecutive same-type runs → average their geometric parameters
	visited := make([]bool, n)
	for start := 0; start < n; start++ {
		if visited[start] {
			continue
		}
		typ := edges[start].Type
		run := []int{start}
		for {
			next := (run[len(run)-1] + 1) % n
			if next == start || edges[next].Type != typ || visited[next] {
				break
			}
			run = append(run, next)
		}
		if len(run) < 2 {
			continue
		}
		for _, k := range run {
			visited[k] = true
		}
		if typ == "radial" {
			var sum float64
			for _, k := range run {
				sum += edges[k].Theta
			}
			avg := sum / float64(len(run))
			for _, k := range run {
				edges[k].Theta = avg
			}
		} else {
			var sum float64
			for _, k := range run {
				sum += edges[k].R
			}
			avg := sum / float64(len(run))
			for _, k := range run {
				edges[k].R = avg
			}
		}
	}

	snapped = make([]Mark, n)
	for i := range m {
		prev := edges[(i-1+n)%n]
		next := edges[i]
		r, theta := snapPoint(m[i], prev, next)
		x, z := fromPolar(r, theta)
		snapped[i] = Mark{ID: m[i].ID, R: r, Theta: theta, X: x, Z: z}
	}
	return snapped, edges
}

// edgeKey is the canonical (unordered) pair of cluster IDs identifying an edge.
type edgeKey struct{ a, b int }

func newEdgeKey(a, b int) edgeKey {
	if a > b {
		a, b = b, a
	}
	return edgeKey{a, b}
}

// globalOptimise computes optimal geometric parameters for every arc and radial
// chain in the boundary graph, then positions every vertex at the intersection
// of its arc chain's r and its radial chain's θ.
//
// Passes:
//   A – average first-pass snapped positions per cluster
//   B – average edge parameters across regions sharing the same edge
//   C – global chain solve: find arc and radial connected components,
//       compute the least-squares (mean) parameter for each chain,
//       and set every vertex in the chain to that parameter.
//
// Simplification: arc and radial chains are treated as decoupled — a valid
// approximation that keeps every shared vertex at exactly one position.
func globalOptimise(results []regionResult, clusters []Cluster) (finalPos []Mark, globalEdge map[edgeKey]Edge) {
	// ── Pass A: average snapped positions per cluster ────────────────────────
	snapX := map[int][]float64{}
	snapZ := map[int][]float64{}
	for _, res := range results {
		for i, cid := range res.clusterIDs {
			snapX[cid] = append(snapX[cid], res.snapped[i].X)
			snapZ[cid] = append(snapZ[cid], res.snapped[i].Z)
		}
	}
	globalPos := make([]Mark, len(clusters))
	for cid := range clusters {
		xs, zs := snapX[cid], snapZ[cid]
		var sumX, sumZ float64
		for i := range xs {
			sumX += xs[i]
			sumZ += zs[i]
		}
		avgX, avgZ := sumX/float64(len(xs)), sumZ/float64(len(zs))
		r, theta := toPolar(avgX, avgZ)
		globalPos[cid] = Mark{R: r, Theta: theta, X: avgX, Z: avgZ}
	}

	// ── Pass B: average edge parameters across regions ───────────────────────
	edgeParams := map[edgeKey][]Edge{}
	for _, res := range results {
		ne := len(res.edges)
		for i, e := range res.edges {
			ek := newEdgeKey(res.clusterIDs[i], res.clusterIDs[(i+1)%ne])
			edgeParams[ek] = append(edgeParams[ek], e)
		}
	}
	globalEdge = map[edgeKey]Edge{}
	for ek, edges := range edgeParams {
		avg := edges[0]
		if len(edges) > 1 {
			if avg.Type == "radial" {
				var sum float64
				for _, e := range edges {
					sum += e.Theta
				}
				avg.Theta = sum / float64(len(edges))
			} else {
				var sum float64
				for _, e := range edges {
					sum += e.R
				}
				avg.R = sum / float64(len(edges))
			}
		}
		avg.From, avg.To = ek.a, ek.b
		globalEdge[ek] = avg
	}

	// ── Pass C: global chain solve ───────────────────────────────────────────
	//
	// Arc chain:    all vertices globally connected by arc edges → share one r
	// Radial chain: all vertices globally connected by radial edges → share one θ
	//
	// Each chain's optimal parameter = mean of all vertex measurements in it
	// (the least-squares solution under equal-weight observations).
	// Each vertex sits at fromPolar(r_from_arc_chain, θ_from_radial_chain).

	nc := len(clusters)
	arcDSU := newDSU(nc)
	radDSU := newDSU(nc)
	for ek, ge := range globalEdge {
		if ge.Type == "arc" {
			arcDSU.union(ek.a, ek.b)
		} else {
			radDSU.union(ek.a, ek.b)
		}
	}

	// Collect r measurements per arc chain
	arcChainR := map[int][]float64{}
	for cid := range clusters {
		root := arcDSU.find(cid)
		arcChainR[root] = append(arcChainR[root], globalPos[cid].R)
	}
	arcMean := map[int]float64{}
	for root, rs := range arcChainR {
		var sum float64
		for _, r := range rs {
			sum += r
		}
		arcMean[root] = sum / float64(len(rs))
	}

	// Collect θ measurements per radial chain
	radChainT := map[int][]float64{}
	for cid := range clusters {
		root := radDSU.find(cid)
		radChainT[root] = append(radChainT[root], globalPos[cid].Theta)
	}
	radMean := map[int]float64{}
	for root, ts := range radChainT {
		var sum float64
		for _, t := range ts {
			sum += t
		}
		radMean[root] = sum / float64(len(ts))
	}

	// A vertex with no arc edges is its own arc chain (singleton) — preserve r.
	// Same for radial singletons.
	arcConnected := map[int]bool{}
	radConnected := map[int]bool{}
	for ek, ge := range globalEdge {
		if ge.Type == "arc" {
			arcConnected[ek.a] = true
			arcConnected[ek.b] = true
		} else {
			radConnected[ek.a] = true
			radConnected[ek.b] = true
		}
	}

	finalPos = make([]Mark, nc)
	for cid := range clusters {
		r := globalPos[cid].R
		if arcConnected[cid] {
			r = arcMean[arcDSU.find(cid)]
		}
		theta := globalPos[cid].Theta
		if radConnected[cid] {
			theta = radMean[radDSU.find(cid)]
		}
		x, z := fromPolar(r, theta)
		finalPos[cid] = Mark{R: r, Theta: theta, X: x, Z: z}
		clusters[cid].R, clusters[cid].Theta = r, theta
		clusters[cid].X, clusters[cid].Z = x, z
	}

	return finalPos, globalEdge
}
