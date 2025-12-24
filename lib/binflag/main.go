package binflag

import "golang.org/x/exp/constraints"

func Has[T constraints.Integer](flags, flag T) bool {
	return flags&flag == flag
}
