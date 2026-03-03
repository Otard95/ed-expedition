package database

import ()

const (
	// Hilbert curve: order 20 for ~0.1 ly precision (fits in 60 bits)
	HilbertOrder = 20
	HilbertBits  = 60

	// Galaxy coordinate origin and scaling
	OriginX    = -43000.0
	OriginY    = -30000.0
	OriginZ    = -24000.0
	CoordScale = 10 // 0.1 ly precision
)
