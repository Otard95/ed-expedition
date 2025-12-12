package slice

import "slices"

func Split[T comparable](s []T, delim T) [][]T {
	n := Count(s, delim)
	result := make([][]T, n+1)

	i := 0
	for i < n {
		j := slices.Index(s, delim)
		result[i] = s[:j]
		s = s[j+1:]
		i++
	}
	result[i] = s

	return result
}

func Count[T comparable](s []T, el T) int {
	c := 0
	for _, e := range s {
		if e == el {
			c++
		}
	}
	return c
}
