package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const clusterThreshold = 300.0 // ly — marks closer than this are the same intersection

type dsu struct{ parent []int }

func newDSU(n int) *dsu {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &dsu{p}
}

func (d *dsu) find(x int) int {
	for d.parent[x] != x {
		d.parent[x] = d.parent[d.parent[x]]
		x = d.parent[x]
	}
	return x
}

func (d *dsu) union(x, y int) {
	if rx, ry := d.find(x), d.find(y); rx != ry {
		d.parent[ry] = rx
	}
}

// clusterMarks groups co-located marks from different regions into clusters.
// Returns the cluster slice and a lookup from "region:markID" → cluster index.
func clusterMarks(all []TaggedMark) (clusters []Cluster, markToCluster map[string]int) {
	n := len(all)
	ds := newDSU(n)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if all[i].region != all[j].region &&
				edist(all[i].mark, all[j].mark) < clusterThreshold {
				ds.union(i, j)
			}
		}
	}

	clusterMap := map[int][]int{}
	for i := 0; i < n; i++ {
		root := ds.find(i)
		clusterMap[root] = append(clusterMap[root], i)
	}

	var roots []int
	for root := range clusterMap {
		roots = append(roots, root)
	}
	sort.Ints(roots)

	clusters = make([]Cluster, len(roots))
	rootToID := map[int]int{}

	for idx, root := range roots {
		members := clusterMap[root]
		var sumX, sumZ float64
		var sources []string
		for _, mi := range members {
			tm := all[mi]
			sumX += tm.mark.X
			sumZ += tm.mark.Z
			sources = append(sources, fmt.Sprintf("%s:%d", tm.region, tm.mark.ID))
		}
		avgX, avgZ := sumX/float64(len(members)), sumZ/float64(len(members))
		r, theta := toPolar(avgX, avgZ)

		clusters[idx] = Cluster{ID: idx, Sources: sources, R: r, Theta: theta, X: avgX, Z: avgZ}
		rootToID[root] = idx

		if len(members) > 1 {
			var rms float64
			for _, mi := range members {
				dx, dz := all[mi].mark.X-avgX, all[mi].mark.Z-avgZ
				rms += dx*dx + dz*dz
			}
			rms = math.Sqrt(rms / float64(len(members)))
			fmt.Printf("  Shared cluster %2d (%d marks, RMS %3.0f ly): %s\n",
				idx, len(members), rms, strings.Join(sources, ", "))
		}
	}

	markToCluster = map[string]int{}
	for i, tm := range all {
		root := ds.find(i)
		markToCluster[fmt.Sprintf("%s:%d", tm.region, tm.mark.ID)] = rootToID[root]
	}

	return clusters, markToCluster
}
