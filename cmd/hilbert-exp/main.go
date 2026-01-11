package main

import (
	"fmt"

	"gonum.org/v1/gonum/spatial/curve"
)

func main() {
	h, err := curve.NewHilbert3D(4)
	if err != nil {
		fmt.Printf("ERROR: %s", err.Error())
	}

	for x := range 3 {
		for y := range 3 {
			for z := range 3 {
				fmt.Printf("(%d %d %d) = %d\n", x-1, y-1, z-1, h.Pos([]int{x - 1, y - 1, z - 1}))
			}
		}
	}
}
