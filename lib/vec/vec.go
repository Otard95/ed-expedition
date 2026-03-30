package vec

import (
	"golang.org/x/exp/constraints"
	"math"
)

type Number interface {
	constraints.Float | constraints.Integer
}

type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func NewVec3[T Number](x, y, z T) Vec3 {
	return Vec3{float64(x), float64(y), float64(z)}
}

func NewVec3FromSlice[T Number](s []T) Vec3 {
	if len(s) != 3 {
		panic("NewVec3FromSlice: s was not length 3")
	}
	return Vec3{float64(s[0]), float64(s[1]), float64(s[2])}
}

func UnpackAs[T Number](v Vec3) (x, y, z T) {
	return T(v.X), T(v.Y), T(v.Z)
}

func SliceOf[T Number](v Vec3) []T {
	return []T{T(v.X), T(v.Y), T(v.Z)}
}

func (v Vec3) Unpack() (x, y, z float64) {
	return v.X, v.Y, v.Z
}

func (v Vec3) Slice() []float64 {
	return []float64{v.X, v.Y, v.Z}
}

func (v Vec3) Clone() Vec3 {
	return v
}

func (v Vec3) Len() float64 {
	return math.Sqrt(v.SqLen())
}

func (v Vec3) SqLen() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3) Add(v2 Vec3) Vec3 {
	return NewVec3(v.X+v2.X, v.Y+v2.Y, v.Z+v2.Z)
}

func (v Vec3) Sub(v2 Vec3) Vec3 {
	return NewVec3(v.X-v2.X, v.Y-v2.Y, v.Z-v2.Z)
}

func (v Vec3) Scale(f float64) Vec3 {
	return NewVec3(v.X*f, v.Y*f, v.Z*f)
}

func (v Vec3) Mag(f float64) Vec3 {
	return v.Scale(f / v.Len())
}

func (v Vec3) Norm() Vec3 {
	l := v.Len()
	return NewVec3(v.X/l, v.Y/l, v.Z/l)
}

func (v Vec3) Neg() Vec3 {
	return NewVec3(-v.X, -v.Y, -v.Z)
}

func (v Vec3) Distance(v2 Vec3) float64 {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	dz := v.Z - v2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func (v Vec3) SqDistance(v2 Vec3) float64 {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	dz := v.Z - v2.Z
	return dx*dx + dy*dy + dz*dz
}

func (v Vec3) Dot(v2 Vec3) float64 {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
}

func (v Vec3) Cross(v2 Vec3) Vec3 {
	return Vec3{
		v.Y*v2.Z - v.Z*v2.Y,
		v.Z*v2.X - v.X*v2.Z,
		v.X*v2.Y - v.Y*v2.X,
	}
}
