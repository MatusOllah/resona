package testutil

import (
	"math"
	"slices"
)

func EqualWithinTolerance(a, b, epsilon float64) bool {
	if a == b {
		return true
	}

	return math.Abs(a-b) <= epsilon
}

func EqualSliceWithinTolerance(a, b []float64, epsilon float64) bool {
	if a == nil && b == nil {
		return true
	}

	return slices.EqualFunc(a, b, func(aa, bb float64) bool {
		return EqualWithinTolerance(aa, bb, epsilon)
	})
}
