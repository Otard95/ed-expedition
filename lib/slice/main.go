package slice

import (
	"slices"
)

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

func Find[T any](s []T, predicate func(T) bool) *T {
	i := slices.IndexFunc(s, predicate)
	if i < 0 {
		return nil
	}
	return &s[i]
}

func Map[T any, R any](s []T, transform func(*T) R) []R {
	result := make([]R, 0, len(s))
	for _, item := range s {
		result = append(result, transform(&item))
	}
	return result
}
