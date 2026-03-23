package services

import (
	"ed-expedition/database"
	"ed-expedition/lib/vec"
)

type GalaxySystem struct {
	Id        uint64
	Name      string
	Position  vec.Vec3[float64]
	StarClass uint8
}

func (s GalaxySystem) StarClassName() string {
	return database.StarClassName(s.StarClass)
}

func (s GalaxySystem) IsScoopable() bool {
	return database.IsScoopableStarClass(s.StarClass)
}
