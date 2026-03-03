package database

type System struct {
	Id         uint64
	hilbertKey uint64
	Name       string
	X, Y, Z    uint32
	StarClass  uint8
}
