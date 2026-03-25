package database

import "ed-expedition/lib/vec"

func NormalizeCoord(v vec.Vec3) vec.Vec3 {
	return v.Sub(Origin).Scale(CoordScale)
}

func DenormalizeCoord(v vec.Vec3) vec.Vec3 {
	return Origin.Add(v.Scale(1.0 / CoordScale))
}
