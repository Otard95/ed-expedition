package database

func NormalizeCoord(x, y, z float64) (nx uint32, ny uint32, nz uint32) {
	nx = uint32(int(x-OriginX) * CoordScale)
	ny = uint32(int(y-OriginY) * CoordScale)
	nz = uint32(int(z-OriginZ) * CoordScale)
	return nx, ny, nz
}

func DenormalizeCoord(x, y, z uint32) (dx float64, dy float64, dz float64) {
	dx = OriginX + (float64(x) / float64(CoordScale))
	dy = OriginY + (float64(y) / float64(CoordScale))
	dz = OriginZ + (float64(z) / float64(CoordScale))
	return dx, dy, dz
}
