package database

type System struct {
	Id         uint64 `json:"id"`
	hilbertKey uint64
	Name       string `json:"name"`
	X          uint32 `json:"x"`
	Y          uint32 `json:"y"`
	Z          uint32 `json:"z"`
	StarClass  uint8  `json:"star_class"`
}
