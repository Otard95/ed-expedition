package vec

import (
	"golang.org/x/exp/constraints"
	"math"
)

type Vec3[T constraints.Float] struct {
	X T `json:"x"`
	Y T `json:"y"`
	Z T `json:"z"`
}

func NewVec3[T constraints.Float](x, y, z T) Vec3[T] {
	return Vec3[T]{x, y, z}
}

func NewVec3FromSlice[T constraints.Float](s []T) Vec3[T] {
	if len(s) != 3 {
		panic("NewVec3FromSlice: s was not length 3")
	}
	return Vec3[T]{s[0], s[1], s[2]}
}

func (v Vec3[T]) Unpack() (x, y, z T) {
	return v.X, v.Y, v.Z
}

func (v Vec3[T]) Slice() []T {
	return []T{v.X, v.Y, v.Z}
}

func (v Vec3[T]) Clone() Vec3[T] {
	return v
}

func (v Vec3[T]) Len() T {
	return T(math.Sqrt(float64(v.SqLen())))
}

func (v Vec3[T]) SqLen() T {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3[T]) Add(v2 Vec3[T]) Vec3[T] {
	return NewVec3(v.X+v2.X, v.Y+v2.Y, v.Z+v2.Z)
}

func (v Vec3[T]) Sub(v2 Vec3[T]) Vec3[T] {
	return NewVec3(v.X-v2.X, v.Y-v2.Y, v.Z-v2.Z)
}

func (v Vec3[T]) Scale(f T) Vec3[T] {
	return NewVec3(v.X*f, v.Y*f, v.Z*f)
}

func (v Vec3[T]) Norm() Vec3[T] {
	l := v.Len()
	return NewVec3(v.X/l, v.Y/l, v.Z/l)
}

func (v Vec3[T]) Neg() Vec3[T] {
	return NewVec3(-v.X, -v.Y, -v.Z)
}

func (v Vec3[T]) Distance(v2 Vec3[T]) T {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	dz := v.Z - v2.Z
	return T(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

func (v Vec3[T]) SqDistance(v2 Vec3[T]) T {
	dx := v.X - v2.X
	dy := v.Y - v2.Y
	dz := v.Z - v2.Z
	return dx*dx + dy*dy + dz*dz
}

func (v Vec3[T]) Dot(v2 Vec3[T]) T {
	return v.X*v2.X + v.Y*v2.Y + v.Z*v2.Z
}

func (v Vec3[T]) Cross(v2 Vec3[T]) Vec3[T] {
	return Vec3[T]{
		v.Y*v2.Z - v.Z*v2.Y,
		v.Z*v2.X - v.X*v2.Z,
		v.X*v2.Y - v.Y*v2.X,
	}
}
