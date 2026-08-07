package main

import "math"

const (
	gcX = 436.14
	gcZ = 26826.65

	arcStepRad = 0.008 // radians per interpolated arc step
)

func toPolar(x, z float64) (r, theta float64) {
	dx, dz := x-gcX, z-gcZ
	return math.Sqrt(dx*dx + dz*dz), math.Atan2(dx, dz)
}

func fromPolar(r, theta float64) (x, z float64) {
	return gcX + r*math.Sin(theta), gcZ + r*math.Cos(theta)
}

func edist(a, b Mark) float64 {
	dx, dz := a.X-b.X, a.Z-b.Z
	return math.Sqrt(dx*dx + dz*dz)
}

// classifyEdge decides whether the edge from a to b is RADIAL (constant θ from
// the galactic centre) or ARC (constant r). The heuristic compares |Δθ|*avgR
// (angular displacement in ly) against |Δr| (radial displacement in ly).
func classifyEdge(a, b Mark) (typ string, theta, r float64) {
	avgR := (a.R + b.R) / 2
	dr := math.Abs(b.R - a.R)
	dt := math.Abs(b.Theta - a.Theta)
	if dt*avgR < dr {
		return "radial", (a.Theta + b.Theta) / 2, 0
	}
	return "arc", 0, avgR
}

// snapPoint returns the exact (r, θ) for a vertex given its two adjacent edge
// primitives. The vertex lies at their intersection.
func snapPoint(mark Mark, prev, next Edge) (r, theta float64) {
	switch {
	case prev.Type == "arc" && next.Type == "radial":
		return prev.R, next.Theta
	case prev.Type == "radial" && next.Type == "arc":
		return next.R, prev.Theta
	case prev.Type == "radial":
		// Intermediate on a shared radial — snap θ, preserve r from mark
		return mark.R, (prev.Theta + next.Theta) / 2
	default:
		// Intermediate on a shared arc — snap r, preserve θ from mark
		return (prev.R + next.R) / 2, mark.Theta
	}
}

// arcSegments interpolates an arc from (x0,z0) to (x1,z1) at constant radius r.
// The first and last points are the exact supplied coordinates to guarantee
// connectivity with adjacent segments regardless of floating-point precision.
func arcSegments(r, theta0, theta1, x0, z0, x1, z1 float64) [][4]float64 {
	dt := theta1 - theta0
	for dt > math.Pi {
		dt -= 2 * math.Pi
	}
	for dt < -math.Pi {
		dt += 2 * math.Pi
	}
	steps := int(math.Ceil(math.Abs(dt) / arcStepRad))
	if steps < 2 {
		steps = 2
	}
	segs := make([][4]float64, 0, steps)
	px, pz := x0, z0
	for i := 1; i <= steps; i++ {
		var cx, cz float64
		if i == steps {
			cx, cz = x1, z1 // exact endpoint guarantees connectivity
		} else {
			t := theta0 + dt*float64(i)/float64(steps)
			cx, cz = fromPolar(r, t)
		}
		segs = append(segs, [4]float64{px, pz, cx, cz})
		px, pz = cx, cz
	}
	return segs
}
